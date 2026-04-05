package reader

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"gpdf/model"
	"gpdf/security"
	"gpdf/stream"
	"gpdf/syntax/impl"
	"gpdf/xref"
)

// documentCore holds the low-level state for resolving, caching, and decrypting
// indirect objects from a PDF file.
type documentCore struct {
	r               io.ReaderAt
	size            int64
	xref            xref.Table
	trailer         model.Trailer
	startXRefOffset int64
	maxStreamBytes  int
	maxFilterChain  int
	filters         stream.FilterRegistry
	decryptor       security.Decryptor
	version         PDFVersion
	linearization   *LinearizationInfo

	mu         sync.Mutex
	objects    map[model.Ref]model.Object
	cacheOrder []model.Ref
	cacheLimit int
}

func (c *documentCore) Trailer() model.Trailer            { return c.trailer }
func (c *documentCore) StartXRefOffset() int64            { return c.startXRefOffset }
func (c *documentCore) ObjectNumbers() []int              { return c.xref.ObjectNumbers() }
func (c *documentCore) Version() PDFVersion               { return c.version }
func (c *documentCore) Linearization() *LinearizationInfo { return c.linearization }

// resolveRaw parses and returns the object at ref without decryption or caching (used for Encrypt dict).
func (c *documentCore) resolveRaw(ref model.Ref) (model.Object, error) {
	e, ok := c.xref.Get(ref.ObjectNumber)
	if !ok || !e.InUse {
		return nil, fmt.Errorf("object %d %d R not in xref", ref.ObjectNumber, ref.Generation)
	}
	p := impl.NewParser(c.r, c.size)
	if err := p.SetPosition(e.Offset); err != nil {
		return nil, err
	}
	indirect, inline, err := p.ParseObject()
	if err != nil {
		return nil, err
	}
	if indirect == nil {
		return nil, fmt.Errorf("object %d %d R (offset %d): expected indirect object, got %T",
			ref.ObjectNumber, ref.Generation, e.Offset, inline)
	}
	if indirect.ObjectNumber != ref.ObjectNumber || indirect.Generation != ref.Generation {
		return nil, fmt.Errorf("object %d %d R (offset %d): xref points to %d %d obj",
			ref.ObjectNumber, ref.Generation, e.Offset, indirect.ObjectNumber, indirect.Generation)
	}
	return indirect.Value, nil
}

func (c *documentCore) Resolve(ref model.Ref) (model.Object, error) {
	c.mu.Lock()
	if obj, ok := c.objects[ref]; ok {
		c.mu.Unlock()
		return obj, nil
	}
	c.mu.Unlock()

	e, ok := c.xref.Get(ref.ObjectNumber)
	if !ok || !e.InUse {
		return nil, fmt.Errorf("object %d %d R not in xref", ref.ObjectNumber, ref.Generation)
	}
	if !e.Compressed && e.Generation != ref.Generation {
		return nil, fmt.Errorf("object %d: generation mismatch (xref %d, ref %d)",
			ref.ObjectNumber, e.Generation, ref.Generation)
	}

	var obj model.Object
	var err error
	if e.Compressed {
		obj, err = c.resolveFromObjectStream(e.StreamObjectNumber, e.IndexInStream)
		if err != nil {
			return nil, fmt.Errorf("object %d %d R (object stream %d, index %d): %w",
				ref.ObjectNumber, ref.Generation, e.StreamObjectNumber, e.IndexInStream, err)
		}
	} else {
		obj, err = c.resolveIndirect(ref, e)
		if err != nil {
			return nil, err
		}
	}

	c.mu.Lock()
	if cached, ok := c.objects[ref]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.objects[ref] = obj
	c.cacheOrder = append(c.cacheOrder, ref)
	c.enforceObjectCacheLimitLocked()
	c.mu.Unlock()

	return obj, nil
}

func (c *documentCore) enforceObjectCacheLimitLocked() {
	if c.cacheLimit <= 0 {
		return
	}
	for len(c.cacheOrder) > c.cacheLimit {
		oldest := c.cacheOrder[0]
		c.cacheOrder = c.cacheOrder[1:]
		delete(c.objects, oldest)
	}
}

func (c *documentCore) resolveIndirect(ref model.Ref, e xref.Entry) (model.Object, error) {
	p := impl.NewParser(c.r, c.size)
	if err := p.SetPosition(e.Offset); err != nil {
		return nil, err
	}
	indirect, inline, err := p.ParseObject()
	if err != nil {
		return nil, fmt.Errorf("object %d %d R (offset %d): %w", ref.ObjectNumber, ref.Generation, e.Offset, err)
	}
	if indirect == nil {
		return nil, fmt.Errorf("object %d %d R (offset %d): expected indirect object, got %T",
			ref.ObjectNumber, ref.Generation, e.Offset, inline)
	}
	if indirect.ObjectNumber != ref.ObjectNumber || indirect.Generation != ref.Generation {
		return nil, fmt.Errorf("object %d %d R (offset %d): xref points to %d %d obj",
			ref.ObjectNumber, ref.Generation, e.Offset, indirect.ObjectNumber, indirect.Generation)
	}
	obj, err := c.decryptObject(indirect.Value, ref)
	if err != nil {
		return nil, fmt.Errorf("object %d %d R (offset %d): %w", ref.ObjectNumber, ref.Generation, e.Offset, err)
	}
	return obj, nil
}

func (c *documentCore) resolveFromObjectStream(streamObjNum, index int) (model.Object, error) {
	streamRef := model.Ref{ObjectNumber: streamObjNum, Generation: 0}

	c.mu.Lock()
	cachedStream, cached := c.objects[streamRef]
	c.mu.Unlock()

	var objStm *model.Stream
	if cached {
		var ok bool
		objStm, ok = cachedStream.(*model.Stream)
		if !ok {
			return nil, fmt.Errorf("object stream %d is not a stream (%T)", streamObjNum, cachedStream)
		}
	} else {
		s, err := c.parseObjectStream(streamObjNum, streamRef)
		if err != nil {
			return nil, err
		}
		objStm = s
		c.mu.Lock()
		c.objects[streamRef] = objStm
		c.mu.Unlock()
	}

	return c.extractFromObjectStream(objStm, streamObjNum, index)
}

func (c *documentCore) parseObjectStream(streamObjNum int, streamRef model.Ref) (*model.Stream, error) {
	e, ok := c.xref.Get(streamObjNum)
	if !ok || !e.InUse || e.Compressed {
		return nil, fmt.Errorf("object stream %d not found or is itself compressed", streamObjNum)
	}
	p := impl.NewParser(c.r, c.size)
	if err := p.SetPosition(e.Offset); err != nil {
		return nil, err
	}
	indirect, _, err := p.ParseObject()
	if err != nil {
		return nil, fmt.Errorf("object stream %d: %w", streamObjNum, err)
	}
	if indirect == nil {
		return nil, fmt.Errorf("object stream %d: expected indirect object", streamObjNum)
	}
	s, ok := indirect.Value.(*model.Stream)
	if !ok {
		return nil, fmt.Errorf("object stream %d: expected stream, got %T", streamObjNum, indirect.Value)
	}
	decoded, err := c.decodeStream(s, streamRef)
	if err != nil {
		return nil, fmt.Errorf("object stream %d decode: %w", streamObjNum, err)
	}
	s.Content = decoded
	return s, nil
}

func (c *documentCore) extractFromObjectStream(objStm *model.Stream, streamObjNum, index int) (model.Object, error) {
	n := intFromModelObj(objStm.Dict[model.Name("N")])
	first := intFromModelObj(objStm.Dict[model.Name("First")])
	contentLen := len(objStm.Content)

	if index < 0 || index >= n {
		return nil, fmt.Errorf("object stream %d: index %d out of range (N=%d)", streamObjNum, index, n)
	}
	if first < 0 || first > contentLen {
		return nil, fmt.Errorf("object stream %d: /First %d out of bounds (content length %d)", streamObjNum, first, contentLen)
	}

	header := string(objStm.Content[:first])
	fields := strings.Fields(header)
	if len(fields) < (index+1)*2 {
		return nil, fmt.Errorf("object stream %d: header too short for index %d", streamObjNum, index)
	}
	objOffset, err := strconv.Atoi(fields[index*2+1])
	if err != nil {
		return nil, fmt.Errorf("object stream %d: bad offset at index %d: %w", streamObjNum, index, err)
	}

	var endOffset int
	if index+1 < n && len(fields) >= (index+2)*2 {
		endOffset, _ = strconv.Atoi(fields[(index+1)*2+1])
	} else {
		endOffset = contentLen - first
	}

	absStart := first + objOffset
	absEnd := first + endOffset
	if absStart < 0 || absEnd > contentLen || absStart > absEnd {
		return nil, fmt.Errorf("object stream %d index %d: data range [%d:%d] out of bounds (content length %d)",
			streamObjNum, index, absStart, absEnd, contentLen)
	}

	objData := objStm.Content[absStart:absEnd]
	objReader := bytes.NewReader(objData)
	p := impl.NewParser(objReader, int64(len(objData)))
	_, inline, parseErr := p.ParseObject()
	if parseErr != nil {
		return nil, fmt.Errorf("object stream %d index %d parse: %w", streamObjNum, index, parseErr)
	}
	return inline, nil
}

func (c *documentCore) decryptObject(obj model.Object, ref model.Ref) (model.Object, error) {
	switch v := obj.(type) {
	case model.String:
		if c.decryptor == nil {
			return v, nil
		}
		plain, err := c.decryptor.DecryptString(ref, []byte(v))
		if err != nil {
			return nil, err
		}
		return model.String(plain), nil
	case model.Array:
		out := make(model.Array, len(v))
		for i, item := range v {
			decrypted, err := c.decryptObject(item, ref)
			if err != nil {
				return nil, err
			}
			out[i] = decrypted
		}
		return out, nil
	case model.Dict:
		out := make(model.Dict, len(v))
		for key, item := range v {
			decrypted, err := c.decryptObject(item, ref)
			if err != nil {
				return nil, err
			}
			out[key] = decrypted
		}
		return out, nil
	case *model.Stream:
		return c.decryptStream(v, ref)
	default:
		return obj, nil
	}
}

func (c *documentCore) decryptStream(v *model.Stream, ref model.Ref) (model.Object, error) {
	if v == nil {
		return v, nil
	}
	dictObj, err := c.decryptObject(v.Dict, ref)
	if err != nil {
		return nil, err
	}
	dict, ok := dictObj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("stream dictionary is not a dictionary")
	}
	streamCopy := &model.Stream{
		Dict:    dict,
		Content: bytes.Clone(v.Content),
	}
	decoded, err := c.decodeStream(streamCopy, ref)
	if err != nil {
		return nil, err
	}
	streamCopy.Content = decoded
	return streamCopy, nil
}

func (c *documentCore) decodeStream(s *model.Stream, ref model.Ref) ([]byte, error) {
	data := s.Content
	if c.decryptor != nil {
		dec, err := c.decryptor.DecryptStream(ref, data)
		if err != nil {
			return nil, err
		}
		data = dec
	}
	return applyFiltersWithLimits(data, s.Dict[model.Name("Filter")], c.filters, c.maxStreamBytes, c.maxFilterChain)
}

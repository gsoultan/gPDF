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

type pdfDocument struct {
	r               io.ReaderAt
	size            int64
	xref            xref.Table
	trailer         model.Trailer
	startXRefOffset int64
	filters         stream.FilterRegistry
	decryptor       security.Decryptor

	mu      sync.Mutex
	objects map[model.Ref]model.Object
}

func (d *pdfDocument) Trailer() model.Trailer {
	return d.trailer
}

func (d *pdfDocument) ObjectNumbers() []int {
	return d.xref.ObjectNumbers()
}

func (d *pdfDocument) StartXRefOffset() int64 {
	return d.startXRefOffset
}

// resolveRaw parses and returns the object at ref without decryption or caching (used for Encrypt dict).
func (d *pdfDocument) resolveRaw(ref model.Ref) (model.Object, error) {
	e, ok := d.xref.Get(ref.ObjectNumber)
	if !ok || !e.InUse {
		return nil, fmt.Errorf("object %d %d R not in xref", ref.ObjectNumber, ref.Generation)
	}
	p := impl.NewParser(d.r, d.size)
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

func (d *pdfDocument) Resolve(ref model.Ref) (model.Object, error) {
	d.mu.Lock()
	if obj, ok := d.objects[ref]; ok {
		d.mu.Unlock()
		return obj, nil
	}
	d.mu.Unlock()

	e, ok := d.xref.Get(ref.ObjectNumber)
	if !ok || !e.InUse {
		return nil, fmt.Errorf("object %d %d R not in xref", ref.ObjectNumber, ref.Generation)
	}
	if !e.Compressed && e.Generation != ref.Generation {
		return nil, fmt.Errorf("object %d: generation mismatch (xref %d, ref %d)",
			ref.ObjectNumber, e.Generation, ref.Generation)
	}

	var obj model.Object
	if e.Compressed {
		resolved, err := d.resolveFromObjectStream(e.StreamObjectNumber, e.IndexInStream)
		if err != nil {
			return nil, fmt.Errorf("object %d %d R (object stream %d, index %d): %w",
				ref.ObjectNumber, ref.Generation, e.StreamObjectNumber, e.IndexInStream, err)
		}
		obj = resolved
		// Objects from object streams are already decrypted (the stream itself was decrypted).
	} else {
		p := impl.NewParser(d.r, d.size)
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
		obj, err = d.decryptObject(indirect.Value, ref)
		if err != nil {
			return nil, fmt.Errorf("object %d %d R (offset %d): %w", ref.ObjectNumber, ref.Generation, e.Offset, err)
		}
	}

	d.mu.Lock()
	d.objects[ref] = obj
	d.mu.Unlock()

	return obj, nil
}

// resolveFromObjectStream extracts an object from an object stream (type-2 xref entry).
// streamObjNum is the object number of the ObjStm, index is the position within it.
func (d *pdfDocument) resolveFromObjectStream(streamObjNum, index int) (model.Object, error) {
	streamRef := model.Ref{ObjectNumber: streamObjNum, Generation: 0}

	d.mu.Lock()
	cachedStream, cached := d.objects[streamRef]
	d.mu.Unlock()

	var objStm *model.Stream
	if cached {
		var ok bool
		objStm, ok = cachedStream.(*model.Stream)
		if !ok {
			return nil, fmt.Errorf("object stream %d is not a stream (%T)", streamObjNum, cachedStream)
		}
	} else {
		e, ok := d.xref.Get(streamObjNum)
		if !ok || !e.InUse || e.Compressed {
			return nil, fmt.Errorf("object stream %d not found or is itself compressed", streamObjNum)
		}
		p := impl.NewParser(d.r, d.size)
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
		decoded, err := d.decodeStream(s, streamRef)
		if err != nil {
			return nil, fmt.Errorf("object stream %d decode: %w", streamObjNum, err)
		}
		s.Content = decoded
		objStm = s

		d.mu.Lock()
		d.objects[streamRef] = objStm
		d.mu.Unlock()
	}

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

func (d *pdfDocument) decryptObject(obj model.Object, ref model.Ref) (model.Object, error) {
	switch v := obj.(type) {
	case model.String:
		if d.decryptor == nil {
			return v, nil
		}
		plain, err := d.decryptor.DecryptString(ref, []byte(v))
		if err != nil {
			return nil, err
		}
		return model.String(plain), nil
	case model.Array:
		out := make(model.Array, len(v))
		for i, item := range v {
			decrypted, err := d.decryptObject(item, ref)
			if err != nil {
				return nil, err
			}
			out[i] = decrypted
		}
		return out, nil
	case model.Dict:
		out := make(model.Dict, len(v))
		for key, item := range v {
			decrypted, err := d.decryptObject(item, ref)
			if err != nil {
				return nil, err
			}
			out[key] = decrypted
		}
		return out, nil
	case *model.Stream:
		if v == nil {
			return v, nil
		}
		dictObj, err := d.decryptObject(v.Dict, ref)
		if err != nil {
			return nil, err
		}
		dict, ok := dictObj.(model.Dict)
		if !ok {
			return nil, fmt.Errorf("stream dictionary is not a dictionary")
		}

		streamCopy := &model.Stream{
			Dict:    dict,
			Content: append([]byte(nil), v.Content...),
		}
		decoded, err := d.decodeStream(streamCopy, ref)
		if err != nil {
			return nil, err
		}
		streamCopy.Content = decoded
		return streamCopy, nil
	default:
		return obj, nil
	}
}

func (d *pdfDocument) decodeStream(s *model.Stream, ref model.Ref) ([]byte, error) {
	data := s.Content
	if d.decryptor != nil {
		dec, err := d.decryptor.DecryptStream(ref, data)
		if err != nil {
			return nil, err
		}
		data = dec
	}
	return applyFilters(data, s.Dict[model.Name("Filter")], d.filters)
}

func (d *pdfDocument) Catalog() (*model.Catalog, error) {
	root := d.trailer.Root()
	if root == nil {
		return nil, fmt.Errorf("no trailer Root")
	}
	obj, err := d.Resolve(*root)
	if err != nil {
		return nil, err
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("catalog is not a dictionary")
	}
	return &model.Catalog{Dict: dict}, nil
}

func (d *pdfDocument) Pages() ([]model.Page, error) {
	cat, err := d.Catalog()
	if err != nil {
		return nil, err
	}
	pagesRef, ok := cat.Dict[model.Name("Pages")].(model.Ref)
	if !ok {
		return nil, fmt.Errorf("catalog has no Pages")
	}
	obj, err := d.Resolve(pagesRef)
	if err != nil {
		return nil, err
	}
	return d.collectPages(obj, nil)
}

// inheritableKeys are page-tree keys that children inherit from ancestor nodes (PDF spec 7.7.3.4).
var inheritableKeys = []model.Name{"MediaBox", "CropBox", "Resources", "Rotate"}

func (d *pdfDocument) collectPages(obj model.Object, inherited model.Dict) ([]model.Page, error) {
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, nil
	}

	merged := mergeInherited(inherited, dict)

	typeName, _ := dict[model.Name("Type")].(model.Name)
	if typeName == "Page" {
		for _, key := range inheritableKeys {
			if _, exists := dict[key]; !exists {
				if val, ok := merged[key]; ok {
					dict[key] = val
				}
			}
		}
		return []model.Page{{Dict: dict}}, nil
	}

	kidsObj, ok := dict[model.Name("Kids")].(model.Array)
	if !ok {
		return nil, nil
	}
	var pages []model.Page
	for _, k := range kidsObj {
		ref, ok := k.(model.Ref)
		if !ok {
			continue
		}
		child, err := d.Resolve(ref)
		if err != nil {
			return nil, err
		}
		sub, err := d.collectPages(child, merged)
		if err != nil {
			return nil, err
		}
		pages = append(pages, sub...)
	}
	return pages, nil
}

// mergeInherited builds a dict of inheritable properties, with current node values overriding parent.
func mergeInherited(parent model.Dict, current model.Dict) model.Dict {
	merged := make(model.Dict, len(inheritableKeys))
	for _, key := range inheritableKeys {
		if val, ok := parent[key]; ok {
			merged[key] = val
		}
	}
	for _, key := range inheritableKeys {
		if val, ok := current[key]; ok {
			merged[key] = val
		}
	}
	return merged
}

func (d *pdfDocument) Info() (model.Dict, error) {
	infoRef := d.trailer.Info()
	if infoRef == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*infoRef)
	if err != nil {
		return nil, err
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, nil
	}
	return dict, nil
}

func (d *pdfDocument) MetadataStream() ([]byte, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.MetadataRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	s, ok := obj.(*model.Stream)
	if !ok || s == nil {
		return nil, nil
	}
	return s.Content, nil
}

func (d *pdfDocument) Content() (string, error) {
	return ExtractText(d)
}

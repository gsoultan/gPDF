package reader

import (
	"bytes"
	"fmt"
	"io"

	"gpdf/model"
	"gpdf/security"
	"gpdf/stream"
	"gpdf/syntax"
	"gpdf/syntax/impl"
	"gpdf/xref"
)

type pdfDocument struct {
	r               io.ReaderAt
	size            int64
	xref            xref.Table
	trailer         model.Trailer
	startXRefOffset int64
	parser          syntax.Parser
	objects         map[model.Ref]model.Object
	filters         stream.FilterRegistry
	decryptor       security.Decryptor
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
	if indirect != nil {
		return indirect.Value, nil
	}
	return inline, nil
}

func (d *pdfDocument) Resolve(ref model.Ref) (model.Object, error) {
	if obj, ok := d.objects[ref]; ok {
		return obj, nil
	}
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
		return nil, fmt.Errorf("object %d %d R (offset %d): %w", ref.ObjectNumber, ref.Generation, e.Offset, err)
	}
	var obj model.Object
	if indirect != nil {
		obj = indirect.Value
		d.objects[ref] = obj
	} else {
		obj = inline
	}
	// Decrypt and decode
	if d.decryptor != nil {
		if s, ok := obj.(model.String); ok {
			plain, err := d.decryptor.DecryptString(ref, []byte(s))
			if err != nil {
				return nil, err
			}
			obj = model.String(plain)
			d.objects[ref] = obj
		}
	}
	if stream, ok := obj.(*model.Stream); ok && stream != nil {
		decoded, err := d.decodeStream(stream, ref)
		if err != nil {
			return nil, err
		}
		stream.Content = decoded
	}
	return obj, nil
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
	filterObj := s.Dict[model.Name("Filter")]
	if filterObj == nil {
		return data, nil
	}
	// Single filter name
	if name, ok := filterObj.(model.Name); ok {
		f := d.filters.Get(string(name))
		if f == nil {
			return data, nil
		}
		var out bytes.Buffer
		if err := f.Decode(&out, bytes.NewReader(data), string(name)); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	}
	// Array of filters - decode in order (first decoded is first in array)
	if arr, ok := filterObj.(model.Array); ok {
		for _, o := range arr {
			name, ok := o.(model.Name)
			if !ok {
				break
			}
			f := d.filters.Get(string(name))
			if f == nil {
				break
			}
			var out bytes.Buffer
			if err := f.Decode(&out, bytes.NewReader(data), string(name)); err != nil {
				return nil, err
			}
			data = out.Bytes()
		}
		return data, nil
	}
	return data, nil
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
	return d.collectPages(obj)
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
	stream, ok := obj.(*model.Stream)
	if !ok || stream == nil {
		return nil, nil
	}
	return stream.Content, nil
}

func (d *pdfDocument) collectPages(obj model.Object) ([]model.Page, error) {
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, nil
	}
	// Page tree node: /Type /Pages, /Kids array
	typeName, _ := dict[model.Name("Type")].(model.Name)
	if typeName == "Page" {
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
		sub, err := d.collectPages(child)
		if err != nil {
			return nil, err
		}
		pages = append(pages, sub...)
	}
	return pages, nil
}

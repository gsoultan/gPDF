package doc

import (
	"io"

	"gpdf/model"
	"gpdf/reader"
	"gpdf/writer"
)

// builtDocument is an in-memory document produced by DocumentBuilder.
type builtDocument struct {
	trailer model.Trailer
	objects map[int]model.Object
	size    int
}

func (d *builtDocument) Catalog() (*model.Catalog, error) {
	root := d.trailer.Root()
	if root == nil {
		return nil, nil
	}
	obj := d.objects[root.ObjectNumber]
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, nil
	}
	return &model.Catalog{Dict: dict}, nil
}

func (d *builtDocument) Pages() ([]model.Page, error) {
	root := d.trailer.Root()
	if root == nil {
		return nil, nil
	}
	// Find Pages ref from catalog
	catObj := d.objects[root.ObjectNumber]
	cat, ok := catObj.(model.Dict)
	if !ok {
		return nil, nil
	}
	pagesRef, ok := cat[model.Name("Pages")].(model.Ref)
	if !ok {
		return nil, nil
	}
	pagesObj := d.objects[pagesRef.ObjectNumber]
	pagesDict, ok := pagesObj.(model.Dict)
	if !ok {
		return nil, nil
	}
	kids, ok := pagesDict[model.Name("Kids")].(model.Array)
	if !ok {
		return nil, nil
	}
	var out []model.Page
	for _, k := range kids {
		ref, ok := k.(model.Ref)
		if !ok {
			continue
		}
		pageObj := d.objects[ref.ObjectNumber]
		if dict, ok := pageObj.(model.Dict); ok {
			out = append(out, model.Page{Dict: dict})
		}
	}
	return out, nil
}

func (d *builtDocument) Save(w io.Writer) error {
	pw := writer.NewPDFWriter()
	return pw.Write(w, d)
}

func (d *builtDocument) SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error {
	pw := writer.NewPDFWriter()
	return pw.WriteWithPassword(w, d, userPassword, ownerPassword)
}

// SaveWithAES256Password writes the document encrypted with AES-256. This is
// a stronger alternative to SaveWithPassword while keeping a similar API.
func (d *builtDocument) SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error {
	pw := writer.NewPDFWriter()
	return pw.WriteWithAES256Password(w, d, userPassword, ownerPassword)
}

func (d *builtDocument) SaveLinearized(ws writer.WriteSeeker) error {
	pw := writer.NewPDFWriter()
	return pw.WriteLinearized(ws, d)
}

func (d *builtDocument) ReadContent() (string, error) {
	return reader.ExtractText(d)
}

func (d *builtDocument) Search(keywords ...string) ([]model.SearchResult, error) {
	perPage, err := reader.ExtractTextPerPage(d)
	if err != nil {
		return nil, err
	}
	return reader.SearchPages(perPage, keywords...), nil
}

func (d *builtDocument) Replace(old, new string) error {
	return reader.ReplaceContent(d, old, new)
}

func (d *builtDocument) Replaces(replacements map[string]string) error {
	return reader.ReplacesContent(d, replacements)
}

func (d *builtDocument) Close() error { return nil }

// Ensure builtDocument implements writer.Document (Trailer, Resolve, ObjectNumbers).
var _ interface {
	Trailer() model.Trailer
	Resolve(model.Ref) (model.Object, error)
	ObjectNumbers() []int
} = (*builtDocument)(nil)

func (d *builtDocument) Trailer() model.Trailer { return d.trailer }

func (d *builtDocument) Resolve(ref model.Ref) (model.Object, error) {
	obj, ok := d.objects[ref.ObjectNumber]
	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (d *builtDocument) ObjectNumbers() []int {
	nums := make([]int, 0, d.size-1)
	for i := 1; i < d.size; i++ {
		nums = append(nums, i)
	}
	return nums
}

func (d *builtDocument) Info() (model.Dict, error) {
	infoRef := d.trailer.Info()
	if infoRef == nil {
		return nil, nil
	}
	obj, ok := d.objects[infoRef.ObjectNumber]
	if !ok {
		return nil, nil
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, nil
	}
	return dict, nil
}

func (d *builtDocument) MetadataStream() ([]byte, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.MetadataRef()
	if ref == nil {
		return nil, nil
	}
	obj, ok := d.objects[ref.ObjectNumber]
	if !ok {
		return nil, nil
	}
	stream, ok := obj.(*model.Stream)
	if !ok || stream == nil {
		return nil, nil
	}
	return stream.Content, nil
}

func (d *builtDocument) StartXRefOffset() int64 {
	return 0
}

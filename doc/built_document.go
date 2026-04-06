package doc

import (
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/writer"
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
		return nil, fmt.Errorf("catalog: missing /Root in trailer")
	}
	obj := d.objects[root.ObjectNumber]
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("catalog: root object %d is not a dict", root.ObjectNumber)
	}
	return &model.Catalog{Dict: dict}, nil
}

func (d *builtDocument) Pages() ([]model.Page, error) {
	root := d.trailer.Root()
	if root == nil {
		return nil, fmt.Errorf("pages: missing /Root in trailer")
	}
	// Find Pages ref from catalog
	catObj := d.objects[root.ObjectNumber]
	cat, ok := catObj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("pages: root object %d is not a dict", root.ObjectNumber)
	}
	pagesRef, ok := cat[model.Name("Pages")].(model.Ref)
	if !ok {
		return nil, fmt.Errorf("pages: missing /Pages in catalog")
	}
	pagesObj := d.objects[pagesRef.ObjectNumber]
	pagesDict, ok := pagesObj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("pages: object %d is not a dict", pagesRef.ObjectNumber)
	}
	kids, ok := pagesDict[model.Name("Kids")].(model.Array)
	if !ok {
		return nil, fmt.Errorf("pages: missing /Kids in Pages dict")
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

func (d *builtDocument) ReadContentPerPage() ([]string, error) {
	return reader.ExtractTextPerPage(d)
}

func (d *builtDocument) ReadImages() ([]reader.ImageInfo, error) {
	return reader.ExtractImages(d)
}

func (d *builtDocument) ReadImagesPerPage() ([][]reader.ImageInfo, error) {
	return reader.ExtractImagesPerPage(d)
}

func (d *builtDocument) ReadLayout() ([]reader.PageLayout, error) {
	return reader.ExtractLayout(d)
}

func (d *builtDocument) ReadTables() ([][]reader.Table, error) {
	layouts, err := reader.ExtractLayout(d)
	if err != nil {
		return nil, err
	}
	return reader.DetectTables(layouts), nil
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
		return nil, fmt.Errorf("object %d not found", ref.ObjectNumber)
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
		return nil, fmt.Errorf("info: missing /Info in trailer")
	}
	obj, ok := d.objects[infoRef.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("info: object %d not found", infoRef.ObjectNumber)
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("info: object %d is not a dict", infoRef.ObjectNumber)
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
		return nil, fmt.Errorf("metadata: no /Metadata ref in catalog")
	}
	obj, ok := d.objects[ref.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("metadata: object %d not found", ref.ObjectNumber)
	}
	stream, ok := obj.(*model.Stream)
	if !ok || stream == nil {
		return nil, fmt.Errorf("metadata: object %d is not a stream", ref.ObjectNumber)
	}
	return stream.Content, nil
}

func (d *builtDocument) StartXRefOffset() int64 {
	return 0
}

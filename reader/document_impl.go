package reader

import (
	"fmt"

	"gpdf/model"
)

// ── MetadataProvider ────────────────────────────────────────────────────────

func (d *pdfDocument) XMPMetadata() ([]byte, error) {
	return d.MetadataStream()
}

func (d *pdfDocument) AssociatedFiles() (model.Array, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	if arr := cat.AssociatedFilesArray(); arr != nil {
		return arr, nil
	}
	ref := cat.AFRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	arr, _ := obj.(model.Array)
	return arr, nil
}

func (d *pdfDocument) CatalogLang() (model.String, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return "", err
	}
	return cat.Lang(), nil
}

// ── StructureProvider ────────────────────────────────────────────────────────

func (d *pdfDocument) StructTree() (*model.StructTreeRoot, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.StructTreeRootRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("StructTreeRoot is not a dictionary")
	}
	return &model.StructTreeRoot{Dict: dict}, nil
}

func (d *pdfDocument) IsTagged() (bool, error) {
	markInfo, err := d.MarkInfo()
	if err != nil || markInfo == nil {
		return false, err
	}
	marked, _ := markInfo[model.Name("Marked")].(model.Boolean)
	return bool(marked), nil
}

func (d *pdfDocument) MarkInfo() (model.Dict, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.MarkInfoRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	dict, _ := obj.(model.Dict)
	return dict, nil
}

// ── OutlineProvider ──────────────────────────────────────────────────────────

func (d *pdfDocument) Outlines() (*model.Outlines, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.OutlinesRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("Outlines is not a dictionary")
	}
	return &model.Outlines{Dict: dict}, nil
}

func (d *pdfDocument) OutlineItems() ([]*model.OutlineItem, error) {
	outlines, err := d.Outlines()
	if err != nil || outlines == nil {
		return nil, err
	}
	firstRef := outlines.First()
	if firstRef == nil {
		return nil, nil
	}
	return d.collectOutlineItems(*firstRef)
}

func (d *pdfDocument) collectOutlineItems(ref model.Ref) ([]*model.OutlineItem, error) {
	var items []*model.OutlineItem
	current := ref
	for {
		obj, err := d.Resolve(current)
		if err != nil {
			return nil, err
		}
		dict, ok := obj.(model.Dict)
		if !ok {
			break
		}
		item := &model.OutlineItem{Dict: dict}
		items = append(items, item)
		if firstRef := outlineItemFirstRef(dict); firstRef != nil {
			children, err := d.collectOutlineItems(*firstRef)
			if err != nil {
				return nil, err
			}
			items = append(items, children...)
		}
		nextRef := item.Next()
		if nextRef == nil {
			break
		}
		current = *nextRef
	}
	return items, nil
}

func outlineItemFirstRef(dict model.Dict) *model.Ref {
	if v, ok := dict[model.Name("First")].(model.Ref); ok {
		return &v
	}
	return nil
}

// ── FormProvider ─────────────────────────────────────────────────────────────

func (d *pdfDocument) AcroForm() (*model.AcroForm, error) {
	cat, err := d.Catalog()
	if err != nil || cat == nil {
		return nil, err
	}
	ref := cat.AcroFormRef()
	if ref == nil {
		return nil, nil
	}
	obj, err := d.Resolve(*ref)
	if err != nil {
		return nil, err
	}
	dict, ok := obj.(model.Dict)
	if !ok {
		return nil, fmt.Errorf("AcroForm is not a dictionary")
	}
	return &model.AcroForm{Dict: dict}, nil
}

func (d *pdfDocument) FormFields() ([]model.Dict, error) {
	form, err := d.AcroForm()
	if err != nil || form == nil {
		return nil, err
	}
	return d.resolveFieldRefs(form.Fields())
}

func (d *pdfDocument) resolveFieldRefs(fields model.Array) ([]model.Dict, error) {
	if fields == nil {
		return nil, nil
	}
	dicts := make([]model.Dict, 0, len(fields))
	for _, entry := range fields {
		ref, ok := entry.(model.Ref)
		if !ok {
			continue
		}
		obj, err := d.Resolve(ref)
		if err != nil {
			return nil, err
		}
		dict, ok := obj.(model.Dict)
		if !ok {
			continue
		}
		dicts = append(dicts, dict)
	}
	return dicts, nil
}

// pdfDocument is a thin facade over documentCore that adds page-tree traversal
// and delegates content extraction to package-level helpers.
type pdfDocument struct {
	documentCore
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

func (d *pdfDocument) PDFVersion() PDFVersion { return d.documentCore.Version() }

func (d *pdfDocument) Content() (string, error)              { return ExtractText(d) }
func (d *pdfDocument) ContentPerPage() ([]string, error)     { return ExtractTextPerPage(d) }
func (d *pdfDocument) Images() ([]ImageInfo, error)          { return ExtractImages(d) }
func (d *pdfDocument) ImagesPerPage() ([][]ImageInfo, error) { return ExtractImagesPerPage(d) }
func (d *pdfDocument) Layout() ([]PageLayout, error)         { return ExtractLayout(d) }
func (d *pdfDocument) GenerateCode(opts CodeGenOptions) (GeneratedCode, error) {
	return GenerateCode(d, opts)
}
func (d *pdfDocument) Replace(old, new string) error { return ReplaceContent(d, old, new) }
func (d *pdfDocument) Replaces(replacements map[string]string) error {
	return ReplacesContent(d, replacements)
}

func (d *pdfDocument) Search(keywords ...string) ([]model.SearchResult, error) {
	perPage, err := ExtractTextPerPage(d)
	if err != nil {
		return nil, err
	}
	return SearchPages(perPage, keywords...), nil
}

func (d *pdfDocument) Tables() ([][]Table, error) {
	layouts, err := ExtractLayout(d)
	if err != nil {
		return nil, err
	}
	return DetectTables(layouts), nil
}

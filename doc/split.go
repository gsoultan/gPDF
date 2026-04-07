package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Split splits a document into multiple documents based on the provided page ranges.
// Each range is a slice of 0-based page indices.
func Split(src Document, ranges ...[]int) ([]Document, error) {
	if len(ranges) == 0 {
		return nil, fmt.Errorf("split: no ranges provided")
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	var docs []Document
	for i, r := range ranges {
		doc, err := splitOne(src, allPages, r)
		if err != nil {
			return nil, fmt.Errorf("split: failed at range %d: %w", i, err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// SplitEvery splits a document into multiple documents, each containing at most n pages.
func SplitEvery(src Document, n int) ([]Document, error) {
	if n <= 0 {
		return nil, fmt.Errorf("split: n must be positive")
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	numPages := len(allPages)
	var ranges [][]int
	for i := 0; i < numPages; i += n {
		end := i + n
		if end > numPages {
			end = numPages
		}
		r := make([]int, end-i)
		for j := range r {
			r[j] = i + j
		}
		ranges = append(ranges, r)
	}

	return Split(src, ranges...)
}

// Extract extracts a contiguous range of pages from the document.
// The range is inclusive of start and exclusive of end (start <= page < end).
func Extract(src Document, start, end int) (Document, error) {
	if start < 0 || end < start {
		return nil, fmt.Errorf("split: invalid range [%d, %d)", start, end)
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	if start >= len(allPages) {
		return nil, fmt.Errorf("split: start index %d out of range", start)
	}
	if end > len(allPages) {
		end = len(allPages)
	}

	r := make([]int, end-start)
	for i := range r {
		r[i] = start + i
	}

	return splitOne(src, allPages, r)
}

func splitOne(src Document, allPages []model.Page, indices []int) (Document, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("split: no indices provided")
	}

	dst := &builtDocument{
		objects: make(map[int]model.Object),
		size:    1,
	}

	// 1. Copy Info from the source document (if available)
	if info, err := src.Info(); err == nil && len(info) > 0 {
		cloner := newObjectCloner(src, dst)
		newInfo, err := cloner.clone(info)
		if err == nil {
			infoRef := dst.addObject(newInfo)
			dst.trailer.Dict = model.Dict{
				model.Name("Info"): infoRef,
			}
		}
	} else {
		dst.trailer.Dict = make(model.Dict)
	}

	var clonedPageRefs []model.Ref
	cloner := newObjectCloner(src, dst)

	for _, idx := range indices {
		if idx < 0 || idx >= len(allPages) {
			return nil, fmt.Errorf("split: page index %d out of range (0-%d)", idx, len(allPages)-1)
		}
		p := allPages[idx]

		clonedPageDict, err := cloner.clone(p.Dict)
		if err != nil {
			return nil, fmt.Errorf("split: failed to clone page %d: %w", idx, err)
		}
		pageDict := clonedPageDict.(model.Dict)

		// Remove Parent reference, we'll set it later
		delete(pageDict, model.Name("Parent"))

		pageRef := dst.addObject(pageDict)
		clonedPageRefs = append(clonedPageRefs, pageRef)
	}

	// 2. Create the new Pages tree
	pagesDict := model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Count"): model.Integer(int64(len(clonedPageRefs))),
		model.Name("Kids"):  model.Array(make([]model.Object, len(clonedPageRefs))),
	}
	for i, ref := range clonedPageRefs {
		pagesDict[model.Name("Kids")].(model.Array)[i] = ref
	}
	pagesRef := dst.addObject(pagesDict)

	// Update Parent for each page
	for _, ref := range clonedPageRefs {
		pageDict := dst.objects[ref.ObjectNumber].(model.Dict)
		pageDict[model.Name("Parent")] = pagesRef
	}

	// 3. Create the new Catalog
	catalogDict := model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): pagesRef,
	}
	catalogRef := dst.addObject(catalogDict)

	// 4. Finalize the trailer
	dst.trailer.Dict[model.Name("Root")] = catalogRef
	dst.trailer.Dict[model.Name("Size")] = model.Integer(int64(dst.size))

	return dst, nil
}

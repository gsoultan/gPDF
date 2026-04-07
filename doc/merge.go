package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Merge combines multiple documents into a single new document.
// Pages from the documents are appended in the order they are provided.
func Merge(docs ...Document) (Document, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("merge: no documents provided")
	}

	merged := &builtDocument{
		objects: make(map[int]model.Object),
		size:    1, // Start object numbers at 1
	}

	// For the merged document, we'll create a new Catalog and Pages tree.
	// We'll also copy the Info dictionary from the first document as a base.

	// 1. Copy Info from the first document (if available)
	if info, err := docs[0].Info(); err == nil && len(info) > 0 {
		cloner := newObjectCloner(docs[0], merged)
		newInfo, err := cloner.clone(info)
		if err == nil {
			infoRef := merged.addObject(newInfo)
			merged.trailer.Dict = model.Dict{
				model.Name("Info"): infoRef,
			}
		}
	} else {
		merged.trailer.Dict = make(model.Dict)
	}

	var allPages []model.Ref

	// 2. Clone all pages from all documents
	for _, srcDoc := range docs {
		pages, err := srcDoc.Pages()
		if err != nil {
			return nil, fmt.Errorf("merge: failed to get pages: %w", err)
		}

		cloner := newObjectCloner(srcDoc, merged)
		for _, p := range pages {
			// Clone the page dictionary
			clonedPageDict, err := cloner.clone(p.Dict)
			if err != nil {
				return nil, fmt.Errorf("merge: failed to clone page: %w", err)
			}
			pageDict := clonedPageDict.(model.Dict)

			// Remove Parent reference, we'll set it later
			delete(pageDict, model.Name("Parent"))

			pageRef := merged.addObject(pageDict)
			allPages = append(allPages, pageRef)
		}
	}

	// 3. Create the new Pages tree
	pagesDict := model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Count"): model.Integer(int64(len(allPages))),
		model.Name("Kids"):  model.Array(make([]model.Object, len(allPages))),
	}
	for i, ref := range allPages {
		pagesDict[model.Name("Kids")].(model.Array)[i] = ref
	}
	pagesRef := merged.addObject(pagesDict)

	// Update Parent for each page
	for _, ref := range allPages {
		pageDict := merged.objects[ref.ObjectNumber].(model.Dict)
		pageDict[model.Name("Parent")] = pagesRef
	}

	// 4. Create the new Catalog
	catalogDict := model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): pagesRef,
	}
	catalogRef := merged.addObject(catalogDict)

	// 5. Finalize the trailer
	merged.trailer.Dict[model.Name("Root")] = catalogRef
	merged.trailer.Dict[model.Name("Size")] = model.Integer(int64(merged.size))

	return merged, nil
}

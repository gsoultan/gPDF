package writer

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

const maxValidationDepth = 256

func validateDocumentGraph(doc Document) error {
	return validateDocumentGraphWithOptions(doc, false)
}

func validateIncrementalDocumentGraph(doc Document) error {
	return validateDocumentGraphWithOptions(doc, true)
}

func validateDocumentGraphWithOptions(doc Document, allowExternalRefs bool) error {
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return fmt.Errorf("%w: document has no objects", ErrInvalidDocumentGraph)
	}

	known := make(map[int]struct{}, len(objNums))
	for _, num := range objNums {
		if num <= 0 {
			return fmt.Errorf("%w: invalid object number %d", ErrInvalidDocumentGraph, num)
		}
		if _, exists := known[num]; exists {
			return fmt.Errorf("%w: duplicate object number %d", ErrInvalidDocumentGraph, num)
		}
		known[num] = struct{}{}
	}

	if err := validateObjectRefs(doc.Trailer().Dict, known, 0, allowExternalRefs); err != nil {
		return err
	}

	for _, num := range objNums {
		ref := model.Ref{ObjectNumber: num, Generation: 0}
		obj, err := doc.Resolve(ref)
		if err != nil {
			return fmt.Errorf("%w: failed to resolve %d 0 R: %v", ErrInvalidDocumentGraph, num, err)
		}
		if err := validateObjectRefs(obj, known, 0, allowExternalRefs); err != nil {
			return err
		}
	}

	return nil
}

func validateObjectRefs(obj model.Object, known map[int]struct{}, depth int, allowExternalRefs bool) error {
	if depth > maxValidationDepth {
		return fmt.Errorf("%w: object graph nesting exceeds %d", ErrInvalidDocumentGraph, maxValidationDepth)
	}

	switch v := obj.(type) {
	case model.Ref:
		if _, ok := known[v.ObjectNumber]; !ok && !allowExternalRefs {
			return fmt.Errorf("%w: dangling reference %d %d R", ErrInvalidDocumentGraph, v.ObjectNumber, v.Generation)
		}
	case model.Array:
		for _, item := range v {
			if err := validateObjectRefs(item, known, depth+1, allowExternalRefs); err != nil {
				return err
			}
		}
	case model.Dict:
		for _, item := range v {
			if err := validateObjectRefs(item, known, depth+1, allowExternalRefs); err != nil {
				return err
			}
		}
	case *model.Stream:
		if v == nil {
			return nil
		}
		if err := validateObjectRefs(v.Dict, known, depth+1, allowExternalRefs); err != nil {
			return err
		}
	}

	return nil
}

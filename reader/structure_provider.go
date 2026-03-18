package reader

import "gpdf/model"

// StructureProvider exposes the tagged PDF structure tree (PDF 1.4+, enhanced in PDF 2.0).
type StructureProvider interface {
	// StructTree returns the parsed structure tree root from the catalog /StructTreeRoot, or nil.
	StructTree() (*model.StructTreeRoot, error)
	// IsTagged reports whether the document is a tagged PDF (/MarkInfo /Marked true).
	IsTagged() (bool, error)
	// MarkInfo returns the /MarkInfo dictionary from the catalog, or nil.
	MarkInfo() (model.Dict, error)
}

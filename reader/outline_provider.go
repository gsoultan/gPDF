package reader

import "gpdf/model"

// OutlineProvider exposes the document outline (bookmarks) tree.
type OutlineProvider interface {
	// Outlines returns the document outline root from the catalog /Outlines, or nil.
	Outlines() (*model.Outlines, error)
	// OutlineItems returns all outline items as a flat list in document order (depth-first).
	OutlineItems() ([]*model.OutlineItem, error)
}

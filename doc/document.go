package doc

import "github.com/gsoultan/gpdf/model"

// Document is the main interface for a PDF document (opened or built).
// It composes focused sub-interfaces for catalog access, content reading,
// search/replace, image extraction, layout analysis, and saving.
type Document interface {
	CatalogReader
	ContentReader
	ContentSearcher
	ImageReader
	LayoutReader
	Saver
	Resolve(ref model.Ref) (model.Object, error)
	Close() error
}

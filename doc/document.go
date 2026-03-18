package doc

// Document is the main interface for a PDF document (opened or built).
// It composes focused sub-interfaces for catalog access, content reading,
// search/replace, image extraction, layout analysis, and saving.
type Document interface {
	CatalogReader
	ContentReader
	ContentSearcher
	ImageReader
	LayoutReader
	CodeGenerator
	Saver
	Close() error
}

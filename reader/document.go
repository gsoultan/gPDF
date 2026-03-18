package reader

// Document is the result of reading a PDF. It composes focused sub-interfaces
// for catalog access, object resolution, content extraction, image extraction,
// and layout analysis.
type Document interface {
	CatalogProvider
	ObjectResolver
	ContentExtractor
	ImageExtractor
	LayoutExtractor
}

package reader

// Document is the result of reading a PDF. It composes focused sub-interfaces
// covering catalog access, object resolution, content/image/layout extraction,
// metadata, structure tree, outlines, and forms — providing full PDF 1.0–2.0 support.
type Document interface {
	CatalogProvider
	ObjectResolver
	ContentExtractor
	ImageExtractor
	LayoutExtractor
	MetadataProvider
	StructureProvider
	OutlineProvider
	FormProvider
}

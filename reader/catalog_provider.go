package reader

import "gpdf/model"

// CatalogProvider gives access to the document catalog, page tree, info dict,
// metadata, trailer, PDF version, and linearization information.
type CatalogProvider interface {
	Catalog() (*model.Catalog, error)
	Pages() ([]model.Page, error)
	Info() (model.Dict, error)
	MetadataStream() ([]byte, error)
	StartXRefOffset() int64
	Trailer() model.Trailer
	// PDFVersion returns the version parsed from the file header (e.g. {2, 0} for PDF 2.0).
	PDFVersion() PDFVersion
	// Linearization returns linearization metadata when the file is linearized, or nil.
	Linearization() *LinearizationInfo
}

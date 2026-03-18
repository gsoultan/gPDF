package doc

import "gpdf/model"

// CatalogReader provides access to the document catalog, page tree, info dict, and metadata.
type CatalogReader interface {
	Catalog() (*model.Catalog, error)
	Pages() ([]model.Page, error)
	Info() (model.Dict, error)
	MetadataStream() ([]byte, error)
	StartXRefOffset() int64
	Trailer() model.Trailer
}

package reader

import "gpdf/model"

// CatalogProvider gives access to the document catalog, page tree, info dict, metadata, and trailer.
type CatalogProvider interface {
	Catalog() (*model.Catalog, error)
	Pages() ([]model.Page, error)
	Info() (model.Dict, error)
	MetadataStream() ([]byte, error)
	StartXRefOffset() int64
	Trailer() model.Trailer
}

package reader

import "gpdf/model"

// MetadataProvider exposes document-level metadata streams and associated file collections.
type MetadataProvider interface {
	// XMPMetadata returns the raw XMP metadata XML from the catalog /Metadata stream, or nil.
	XMPMetadata() ([]byte, error)
	// AssociatedFiles returns the /AF array entries from the document catalog (PDF 2.0).
	AssociatedFiles() (model.Array, error)
	// CatalogLang returns the /Lang natural language string from the catalog (PDF 1.4+).
	CatalogLang() (model.String, error)
}

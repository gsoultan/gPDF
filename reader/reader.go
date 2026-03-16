package reader

import "gpdf/model"

// Document is the result of reading a PDF: access to catalog, pages, metadata, and resolved objects.
type Document interface {
	// Catalog returns the document catalog (root), if available.
	Catalog() (*model.Catalog, error)
	// Pages returns the list of page dictionaries in order (flattened from page tree).
	Pages() ([]model.Page, error)
	// Info returns the document Info dictionary (Title, Author, Subject, etc.), or nil if absent.
	Info() (model.Dict, error)
	// MetadataStream returns the raw XMP metadata stream bytes from Catalog /Metadata, or nil if absent.
	MetadataStream() ([]byte, error)
	// StartXRefOffset returns the file offset of the xref table used to read this document (for incremental update). 0 if not from file.
	StartXRefOffset() int64
	// Resolve returns the indirect object at the given reference.
	Resolve(ref model.Ref) (model.Object, error)
	// Trailer returns the trailer dictionary.
	Trailer() model.Trailer
	// ObjectNumbers returns all indirect object numbers (for writing).
	ObjectNumbers() []int
}

// Reader reads a PDF from a random-access source into a Document.
type Reader interface {
	// ReadDocument parses the PDF from r (size bytes) and returns a Document.
	ReadDocument(r interface{ ReadAt(p []byte, off int64) (n int, err error) }, size int64) (Document, error)
	// ReadDocumentWithPassword parses the PDF and decrypts content using the user password (empty for none).
	ReadDocumentWithPassword(r interface{ ReadAt(p []byte, off int64) (n int, err error) }, size int64, userPassword string) (Document, error)
}

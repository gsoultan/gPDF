package doc

import (
	"io"
	"os"

	"gpdf/model"
	"gpdf/reader"
	"gpdf/reader/file"
)

// Document is the main interface for a PDF document (opened or built).
// Callers can read structure (Catalog, Pages), metadata (Info, MetadataStream), then Save or Close.
type Document interface {
	// Catalog returns the document catalog, if available.
	Catalog() (*model.Catalog, error)
	// Pages returns the list of pages in order.
	Pages() ([]model.Page, error)
	// Info returns the document Info dictionary (Title, Author, Subject, etc.), or nil if absent.
	Info() (model.Dict, error)
	// MetadataStream returns the raw XMP metadata stream bytes from Catalog /Metadata, or nil if absent.
	MetadataStream() ([]byte, error)
	// StartXRefOffset returns the file offset of the xref used to read this document (0 if built, not from file). Used for incremental save.
	StartXRefOffset() int64
	// Trailer returns the document trailer (Root, Size, Info, etc.). Used for incremental save to build the patch trailer.
	Trailer() model.Trailer
	// Save writes the document to w in PDF format.
	Save(w io.Writer) error
	// SaveWithPassword writes the document encrypted with the given user and owner passwords (Standard, R=2).
	SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error
	// Close releases resources (e.g. file handle). Idempotent.
	Close() error
}

// Open opens an existing PDF from path and returns a Document.
func Open(path string) (Document, error) {
	return OpenWithPassword(path, "")
}

// OpenWithPassword opens an existing PDF from path and decrypts it with the user password if encrypted.
// Use empty password for unencrypted PDFs or to open without decryption.
func OpenWithPassword(path string, userPassword string) (Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	size := info.Size()
	r := reader.NewPDFReader()
	var doc reader.Document
	if userPassword != "" {
		doc, err = r.ReadDocumentWithPassword(f, size, userPassword)
	} else {
		doc, err = r.ReadDocument(f, size)
	}
	if err != nil {
		f.Close()
		return nil, err
	}
	return file.NewDocument(f, doc), nil
}

// New returns a new DocumentBuilder for constructing a PDF from scratch.
func New() *DocumentBuilder {
	return &DocumentBuilder{}
}

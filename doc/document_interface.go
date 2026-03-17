package doc

import (
	"io"

	"gpdf/model"
	"gpdf/writer"
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
	// SaveWithAES256Password writes the document encrypted with AES-256 (V=5/R=6-style handler).
	SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error
	// SaveLinearized writes a linearized (fast web view) PDF to ws; ws must support Seek (e.g. *os.File).
	SaveLinearized(ws writer.WriteSeeker) error
	// ReadContent returns all extracted text from the document (from all pages).
	ReadContent() (string, error)
	// Search returns, for each keyword, the 0-based page indices where it was found.
	Search(keywords ...string) ([]model.SearchResult, error)
	// Replace replaces all occurrences of old with new in content streams. Call Save to persist.
	Replace(old, new string) error
	// Replaces applies multiple replacements (old -> new) to content streams. Call Save to persist.
	Replaces(replacements map[string]string) error
	// Close releases resources (e.g. file handle). Idempotent.
	Close() error
}

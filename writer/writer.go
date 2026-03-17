package writer

import (
	"io"

	"gpdf/model"
)

// Document is the input to the writer: object graph and trailer (root, size).
type Document interface {
	// Trailer returns the trailer dictionary (e.g. /Root, /Size).
	Trailer() model.Trailer
	// Resolve returns the indirect object at the given reference.
	Resolve(ref model.Ref) (model.Object, error)
	// ObjectNumbers returns all indirect object numbers to write (in order).
	ObjectNumbers() []int
}

// WriteSeeker is the subset of io.Writer required for linearized write (must support Seek).
type WriteSeeker interface {
	io.Writer
	io.Seeker
}

// Writer serializes a PDF document to bytes.
type Writer interface {
	Write(w io.Writer, doc Document) error
	WriteIncremental(w io.Writer, appendOffset int64, baseStartXRef int64, doc Document) error
	// WriteWithPassword writes the document encrypted with Standard (R=2) using user and owner passwords.
	WriteWithPassword(w io.Writer, doc Document, userPassword, ownerPassword string) error
	// WriteWithAES256Password writes the document encrypted with AES-256.
	WriteWithAES256Password(w io.Writer, doc Document, userPassword, ownerPassword string) error
	// WriteLinearized writes a linearized (fast web view) PDF; w must be seekable.
	WriteLinearized(w WriteSeeker, doc Document) error
}

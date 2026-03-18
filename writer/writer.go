package writer

import "io"

// Writer serializes a PDF document to bytes.
type Writer interface {
	Write(w io.Writer, doc Document) error
	WriteIncremental(w io.Writer, appendOffset int64, baseStartXRef int64, doc Document) error
	WriteWithPassword(w io.Writer, doc Document, userPassword, ownerPassword string) error
	WriteWithAES256Password(w io.Writer, doc Document, userPassword, ownerPassword string) error
	WriteLinearized(w WriteSeeker, doc Document) error
}

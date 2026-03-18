package reader

import "io"

// Reader reads a PDF from a random-access source into a Document.
type Reader interface {
	ReadDocument(r io.ReaderAt, size int64) (Document, error)
	ReadDocumentWithPassword(r io.ReaderAt, size int64, userPassword string) (Document, error)
}

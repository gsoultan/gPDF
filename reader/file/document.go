package file

import (
	"io"
	"os"

	"gpdf/model"
	"gpdf/reader"
	"gpdf/writer"
)

// Document wraps a reader.Document and an open file, implementing doc.Document (Save, Close).
type Document struct {
	file *os.File
	doc  reader.Document
	w    writer.Writer
}

// NewDocument returns a Document that delegates read operations to doc and closes file on Close.
func NewDocument(f *os.File, doc reader.Document) *Document {
	return &Document{
		file: f,
		doc:  doc,
		w:    writer.NewPDFWriter(),
	}
}

// Catalog returns the document catalog.
func (d *Document) Catalog() (*model.Catalog, error) {
	return d.doc.Catalog()
}

// Pages returns the list of pages.
func (d *Document) Pages() ([]model.Page, error) {
	return d.doc.Pages()
}

// Info returns the document Info dictionary.
func (d *Document) Info() (model.Dict, error) {
	return d.doc.Info()
}

// MetadataStream returns the XMP metadata stream bytes.
func (d *Document) MetadataStream() ([]byte, error) {
	return d.doc.MetadataStream()
}

// StartXRefOffset returns the file offset of the xref used to read this document (0 if not from file).
func (d *Document) StartXRefOffset() int64 {
	return d.doc.StartXRefOffset()
}

// Trailer returns the document trailer.
func (d *Document) Trailer() model.Trailer {
	return d.doc.Trailer()
}

// Save writes the document to w.
func (d *Document) Save(w io.Writer) error {
	return d.w.Write(w, d.doc)
}

// SaveWithPassword writes the document encrypted with the given passwords.
func (d *Document) SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error {
	return d.w.WriteWithPassword(w, d.doc, userPassword, ownerPassword)
}

// Close closes the underlying file. Idempotent.
func (d *Document) Close() error {
	if d.file == nil {
		return nil
	}
	err := d.file.Close()
	d.file = nil
	return err
}

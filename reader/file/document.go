package file

import (
	"io"
	"os"
	"runtime"

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
	d := &Document{
		file: f,
		doc:  doc,
		w:    writer.NewPDFWriter(),
	}
	runtime.SetFinalizer(d, (*Document).finalize)
	return d
}

func (d *Document) finalize() {
	_ = d.Close()
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

// SaveWithAES256Password writes the document encrypted with AES-256.
func (d *Document) SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error {
	return d.w.WriteWithAES256Password(w, d.doc, userPassword, ownerPassword)
}

// SaveLinearized writes a linearized (fast web view) PDF to ws.
func (d *Document) SaveLinearized(ws writer.WriteSeeker) error {
	return d.w.WriteLinearized(ws, d.doc)
}

// ReadContent returns all extracted text from the document.
func (d *Document) ReadContent() (string, error) {
	return d.doc.Content()
}

// Search returns, for each keyword, the 0-based page indices where it was found.
func (d *Document) Search(keywords ...string) ([]model.SearchResult, error) {
	perPage, err := reader.ExtractTextPerPage(d.doc)
	if err != nil {
		return nil, err
	}
	return reader.SearchPages(perPage, keywords...), nil
}

// Replace replaces all occurrences of old with new in content streams.
func (d *Document) Replace(old, new string) error {
	return reader.ReplaceContent(d.doc, old, new)
}

// Replaces applies multiple replacements to content streams.
func (d *Document) Replaces(replacements map[string]string) error {
	return reader.ReplacesContent(d.doc, replacements)
}

// Resolve returns the indirect object at the given reference.
func (d *Document) Resolve(ref model.Ref) (model.Object, error) {
	return d.doc.Resolve(ref)
}

// ObjectNumbers returns all indirect object numbers (for writing).
func (d *Document) ObjectNumbers() []int {
	return d.doc.ObjectNumbers()
}

// Close closes the underlying file. Idempotent.
func (d *Document) Close() error {
	runtime.SetFinalizer(d, nil)
	if d.file == nil {
		return nil
	}
	err := d.file.Close()
	d.file = nil
	return err
}

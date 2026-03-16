package reader

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"gpdf/model"
	"gpdf/security"
	"gpdf/stream"
	"gpdf/stream/flate"
	"gpdf/syntax/impl"
	"gpdf/xref"
)

// PDFReader implements Reader using the syntax parser and xref.
type PDFReader struct {
	filters stream.FilterRegistry
}

// NewPDFReader returns a PDF reader with default stream filters (e.g. FlateDecode) registered.
func NewPDFReader() *PDFReader {
	reg := stream.NewRegistry()
	reg.Register("FlateDecode", flate.NewFilter())
	return &PDFReader{filters: reg}
}

// NewPDFReaderWithFilters returns a PDF reader using the given filter registry.
func NewPDFReaderWithFilters(filters stream.FilterRegistry) *PDFReader {
	return &PDFReader{filters: filters}
}

// ReadDocument parses the PDF from r and returns a Document.
func (pr *PDFReader) ReadDocument(r io.ReaderAt, size int64) (Document, error) {
	return pr.readDocument(r, size, "")
}

// ReadDocumentWithPassword parses the PDF and decrypts strings/streams using the user password.
// Use empty password for unencrypted PDFs.
func (pr *PDFReader) ReadDocumentWithPassword(r io.ReaderAt, size int64, userPassword string) (Document, error) {
	return pr.readDocument(r, size, userPassword)
}

func (pr *PDFReader) readDocument(r io.ReaderAt, size int64, userPassword string) (Document, error) {
	startXRef, err := findStartXRef(r, size)
	if err != nil {
		return nil, err
	}
	parser := impl.NewParser(r, size)
	if err := parser.SetPosition(startXRef); err != nil {
		return nil, err
	}
	entries, err := parser.ParseXRefTable()
	if err != nil {
		return nil, err
	}
	trailerDict, err := parser.ParseTrailer()
	if err != nil {
		return nil, err
	}
	tbl := xref.NewTable()
	for num, e := range entries {
		tbl.Set(num, xref.Entry{Offset: e.Offset, Generation: e.Generation, InUse: e.InUse})
	}
	trailer := model.Trailer{Dict: trailerDict}
	doc := &pdfDocument{
		r:               r,
		size:            size,
		xref:            tbl,
		trailer:         trailer,
		startXRefOffset: startXRef,
		parser:          impl.NewParser(r, size),
		objects:         make(map[model.Ref]model.Object),
		filters:         pr.filters,
	}
	if userPassword != "" && trailer.Encrypt() != nil {
		encRef := trailer.Encrypt()
		encObj, err := doc.resolveRaw(*encRef)
		if err != nil {
			return nil, fmt.Errorf("encrypt: %w", err)
		}
		encDict, ok := encObj.(model.Dict)
		if !ok {
			return nil, fmt.Errorf("encrypt object is not a dict")
		}
		dec, err := security.NewStandardDecryptor(encDict, trailer.ID(), userPassword)
		if err != nil {
			return nil, fmt.Errorf("decrypt: %w", err)
		}
		doc.decryptor = dec
	}
	return doc, nil
}

// findStartXRef reads from the end of the file to find "startxref" and the following offset.
func findStartXRef(r io.ReaderAt, size int64) (int64, error) {
	const tail = 2048
	if size < 10 {
		return 0, fmt.Errorf("file too small")
	}
	start := size - tail
	if start < 0 {
		start = 0
	}
	buf := make([]byte, size-start)
	_, err := r.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		return 0, err
	}
	// Find last "startxref" in the tail
	idx := bytes.LastIndex(buf, []byte("startxref"))
	if idx < 0 {
		return 0, fmt.Errorf("startxref not found")
	}
	rest := buf[idx+len("startxref"):]
	re := regexp.MustCompile(`\s*(\d+)\s*`)
	m := re.FindSubmatch(rest)
	if m == nil {
		return 0, fmt.Errorf("startxref offset not found")
	}
	off, _ := strconv.ParseInt(string(m[1]), 10, 64)
	return off, nil
}

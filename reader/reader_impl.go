package reader

import (
	"bytes"
	"fmt"
	"io"

	"gpdf/model"
	"gpdf/security"
	"gpdf/stream"
	"gpdf/stream/asciihex"
	"gpdf/stream/ccitt"
	"gpdf/stream/crypt"
	"gpdf/stream/dct"
	"gpdf/stream/flate"
	"gpdf/stream/jbig2"
	"gpdf/stream/jpx"
	"gpdf/stream/runlength"
	"gpdf/syntax"
	"gpdf/syntax/impl"
	"gpdf/xref"

	ascii85filter "gpdf/stream/ascii85"
	lzwfilter "gpdf/stream/lzw"
)

var (
	xrefKeyword     = []byte("xref")
	startXRefMarker = []byte("startxref")
)

// PDFReader implements Reader using the syntax parser and xref.
type PDFReader struct {
	filters stream.FilterRegistry
}

// NewPDFReader returns a PDF reader with default stream filters registered.
func NewPDFReader() *PDFReader {
	reg := stream.NewRegistry()
	reg.Register("FlateDecode", flate.NewFilter())
	reg.Register("DCTDecode", dct.NewFilter())
	reg.Register("LZWDecode", lzwfilter.NewFilter())
	reg.Register("ASCII85Decode", ascii85filter.NewFilter())
	reg.Register("ASCIIHexDecode", asciihex.NewFilter())
	reg.Register("JPXDecode", jpx.NewFilter())
	reg.Register("JBIG2Decode", jbig2.NewFilter())
	reg.Register("CCITTFaxDecode", ccitt.NewFilter())
	reg.Register("RunLengthDecode", runlength.NewFilter())
	reg.Register("Crypt", crypt.NewFilter())
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
	version, err := detectPDFVersion(r)
	if err != nil {
		return nil, err
	}
	linearization, err := detectLinearization(r, size)
	if err != nil {
		return nil, err
	}
	startXRef, err := findStartXRef(r, size)
	if err != nil {
		return nil, err
	}
	entries, trailerDict, err := pr.parseXRefChain(r, size, startXRef)
	if err != nil {
		return nil, err
	}
	tbl := xref.NewTable()
	for num, e := range entries {
		entry := xref.Entry{
			Offset:     e.Offset,
			Generation: e.Generation,
			InUse:      e.InUse,
		}
		if e.Compressed {
			entry.Compressed = true
			entry.StreamObjectNumber = e.StreamObjectNumber
			entry.IndexInStream = e.IndexInStream
		}
		tbl.Set(num, entry)
	}
	trailer := model.Trailer{Dict: trailerDict}
	doc := &pdfDocument{
		documentCore: documentCore{
			r:               r,
			size:            size,
			xref:            tbl,
			trailer:         trailer,
			startXRefOffset: startXRef,
			objects:         make(map[model.Ref]model.Object),
			filters:         pr.filters,
			version:         version,
			linearization:   linearization,
		},
	}
	if encRef := trailer.Encrypt(); encRef != nil {
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

// parseXRefChain follows the /Prev chain to merge all xref sections (table or stream).
func (pr *PDFReader) parseXRefChain(r io.ReaderAt, size, offset int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	allEntries := make(map[int]syntax.XRefEntry)
	var mainTrailer model.Dict
	visited := make(map[int64]struct{})

	currentOffset := offset
	for currentOffset >= 0 {
		if _, seen := visited[currentOffset]; seen {
			return nil, nil, fmt.Errorf("xref /Prev cycle detected at offset %d", currentOffset)
		}
		visited[currentOffset] = struct{}{}

		entries, dict, err := pr.parseXRef(r, size, currentOffset)
		if err != nil {
			return nil, nil, err
		}
		if mainTrailer == nil {
			mainTrailer = dict
		}
		for num, entry := range entries {
			if _, exists := allEntries[num]; !exists {
				allEntries[num] = entry
			}
		}
		if prev, ok := dict[model.Name("Prev")].(model.Integer); ok {
			currentOffset = int64(prev)
		} else {
			break
		}
	}
	return allEntries, mainTrailer, nil
}

// parseXRef dispatches to table or stream parsing based on file content at offset.
// If both approaches fail, it falls back to a full-file scan to rebuild the XRef.
func (pr *PDFReader) parseXRef(r io.ReaderAt, size, offset int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	entries, dict, err := pr.tryParseXRef(r, size, offset)
	if err != nil {
		return pr.repairXRef(r, size, err)
	}
	return entries, dict, nil
}

// tryParseXRef attempts normal XRef parsing (table or stream) at the given offset.
func (pr *PDFReader) tryParseXRef(r io.ReaderAt, size, offset int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	var buf [4]byte
	n, err := r.ReadAt(buf[:], offset)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}
	if n == len(xrefKeyword) && bytes.Equal(buf[:], xrefKeyword) {
		return pr.parseXRefTable(r, size, offset)
	}
	return pr.parseXRefStream(r, size, offset)
}

// repairXRef rebuilds the XRef by scanning the file when normal parsing fails.
func (pr *PDFReader) repairXRef(r io.ReaderAt, size int64, originalErr error) (map[int]syntax.XRefEntry, model.Dict, error) {
	repaired, scanErr := rebuildXRefByScan(r, size)
	if scanErr != nil {
		return nil, nil, fmt.Errorf("xref parse failed (%w); repair scan also failed: %v", originalErr, scanErr)
	}
	return repaired, model.Dict{}, nil
}

func (pr *PDFReader) parseXRefTable(r io.ReaderAt, size, offset int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	parser := impl.NewParser(r, size)
	if err := parser.SetPosition(offset); err != nil {
		return nil, nil, err
	}
	entries, err := parser.ParseXRefTable()
	if err != nil {
		return nil, nil, err
	}
	trailerDict, err := parser.ParseTrailer()
	if err != nil {
		return nil, nil, err
	}
	return entries, trailerDict, nil
}

// parseXRefStream reads an xref stream object (PDF 1.5+) and returns entries + trailer dict.
func (pr *PDFReader) parseXRefStream(r io.ReaderAt, size, offset int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	parser := impl.NewParser(r, size)
	if err := parser.SetPosition(offset); err != nil {
		return nil, nil, err
	}
	indirect, _, err := parser.ParseObject()
	if err != nil {
		return nil, nil, fmt.Errorf("xref stream at %d: %w", offset, err)
	}
	if indirect == nil {
		return nil, nil, fmt.Errorf("xref stream at %d: expected indirect object", offset)
	}
	s, ok := indirect.Value.(*model.Stream)
	if !ok {
		return nil, nil, fmt.Errorf("xref stream at %d: expected stream object", offset)
	}
	decoded, err := decodeStreamData(s, pr.filters)
	if err != nil {
		return nil, nil, fmt.Errorf("xref stream decode: %w", err)
	}

	wArr, ok := s.Dict[model.Name("W")].(model.Array)
	if !ok || len(wArr) < 3 {
		return nil, nil, fmt.Errorf("xref stream: missing or invalid /W")
	}
	w1 := intFromModelObj(wArr[0])
	w2 := intFromModelObj(wArr[1])
	w3 := intFromModelObj(wArr[2])
	entrySize := w1 + w2 + w3
	if entrySize <= 0 {
		return nil, nil, fmt.Errorf("xref stream: entry size is 0")
	}

	sizeVal := intFromModelObj(s.Dict[model.Name("Size")])
	var subsections [][2]int
	if idxArr, ok := s.Dict[model.Name("Index")].(model.Array); ok && len(idxArr) >= 2 {
		for i := 0; i+1 < len(idxArr); i += 2 {
			subsections = append(subsections, [2]int{
				intFromModelObj(idxArr[i]),
				intFromModelObj(idxArr[i+1]),
			})
		}
	} else {
		subsections = [][2]int{{0, sizeVal}}
	}

	entries := make(map[int]syntax.XRefEntry)
	expectedEntries := 0
	for _, sub := range subsections {
		expectedEntries += sub[1]
	}
	if expectedEntries > 0 && len(decoded) < expectedEntries*entrySize {
		return nil, nil, fmt.Errorf("xref stream truncated: have %d bytes, need %d", len(decoded), expectedEntries*entrySize)
	}

	pos := 0
	for _, sub := range subsections {
		start, count := sub[0], sub[1]
		for i := range count {
			if pos+entrySize > len(decoded) {
				break
			}
			entry := decodeXRefStreamEntry(decoded[pos:], w1, w2, w3)
			entries[start+i] = entry
			pos += entrySize
		}
	}
	return entries, s.Dict, nil
}

// decodeXRefStreamEntry decodes a single entry from an xref stream at the given data offset.
func decodeXRefStreamEntry(data []byte, w1, w2, w3 int) syntax.XRefEntry {
	field1 := readBEField(data, w1)
	field2 := readBEField(data[w1:], w2)
	field3 := readBEField(data[w1+w2:], w3)

	typ := field1
	if w1 == 0 {
		typ = 1
	}
	switch typ {
	case 0:
		return syntax.XRefEntry{Offset: int64(field2), Generation: field3, InUse: false}
	case 1:
		return syntax.XRefEntry{Offset: int64(field2), Generation: field3, InUse: true}
	case 2:
		return syntax.XRefEntry{
			InUse:              true,
			Compressed:         true,
			StreamObjectNumber: field2,
			IndexInStream:      field3,
		}
	default:
		return syntax.XRefEntry{}
	}
}

// applyFilters applies the named filter pipeline to data using the given registry.
func applyFilters(data []byte, filterObj model.Object, reg stream.FilterRegistry) ([]byte, error) {
	if filterObj == nil || reg == nil {
		return data, nil
	}
	decode := func(filterName string, in []byte) ([]byte, error) {
		f := reg.Get(filterName)
		if f == nil {
			return nil, fmt.Errorf("unsupported stream filter: %s", filterName)
		}
		out := bytes.NewBuffer(make([]byte, 0, len(in)))
		if err := f.Decode(out, bytes.NewReader(in), filterName); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	}

	switch v := filterObj.(type) {
	case model.Name:
		return decode(string(v), data)
	case model.Array:
		for _, o := range v {
			name, ok := o.(model.Name)
			if !ok {
				return nil, fmt.Errorf("filter array contains non-name element: %T", o)
			}
			decoded, err := decode(string(name), data)
			if err != nil {
				return nil, err
			}
			data = decoded
		}
		return data, nil
	}
	return data, nil
}

// decodeStreamData applies the filter(s) from the stream dict to the raw content.
func decodeStreamData(s *model.Stream, filters stream.FilterRegistry) ([]byte, error) {
	return applyFilters(s.Content, s.Dict[model.Name("Filter")], filters)
}

// readBEField reads a big-endian unsigned integer of the given width (0–4 bytes) from data.
func readBEField(data []byte, width int) int {
	val := 0
	for i := range width {
		if i >= len(data) {
			break
		}
		val = val<<8 | int(data[i])
	}
	return val
}

func intFromModelObj(obj model.Object) int {
	switch v := obj.(type) {
	case model.Integer:
		return int(v)
	case model.Real:
		return int(v)
	}
	return 0
}

// findStartXRef reads from the end of the file to find "startxref" and the following offset.
func findStartXRef(r io.ReaderAt, size int64) (int64, error) {
	const scanChunk = 4096
	const overlap = 64

	if size < 10 {
		return 0, fmt.Errorf("file too small")
	}

	for end := size; ; {
		start := end - scanChunk
		if start < 0 {
			start = 0
		}

		readEnd := end
		if readEnd < size {
			readEnd += overlap
			if readEnd > size {
				readEnd = size
			}
		}

		buf := make([]byte, readEnd-start)
		_, err := r.ReadAt(buf, start)
		if err != nil && err != io.EOF {
			return 0, err
		}

		idx := bytes.LastIndex(buf, startXRefMarker)
		if idx >= 0 {
			rest := buf[idx+len(startXRefMarker):]
			off, err := parseStartXRefOffset(rest, size)
			if err != nil {
				return 0, err
			}
			return off, nil
		}

		if start == 0 {
			break
		}
		end = start
	}

	return 0, fmt.Errorf("startxref not found")
}

func parseStartXRefOffset(rest []byte, size int64) (int64, error) {
	i := 0
	for i < len(rest) && isPDFWhitespace(rest[i]) {
		i++
	}
	if i >= len(rest) || rest[i] < '0' || rest[i] > '9' {
		return 0, fmt.Errorf("startxref offset not found")
	}

	var off int64
	for i < len(rest) {
		b := rest[i]
		if b < '0' || b > '9' {
			break
		}
		next := off*10 + int64(b-'0')
		if next > size {
			return 0, fmt.Errorf("startxref offset %d out of bounds", next)
		}
		off = next
		i++
	}
	if off >= size {
		return 0, fmt.Errorf("startxref offset %d out of bounds", off)
	}
	return off, nil
}

func isPDFWhitespace(b byte) bool {
	switch b {
	case 0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20:
		return true
	default:
		return false
	}
}

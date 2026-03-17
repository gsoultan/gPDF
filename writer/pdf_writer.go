package writer

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sort"

	"gpdf/model"
	"gpdf/security"
	"gpdf/stream"
	"gpdf/stream/asciihex"
	"gpdf/stream/dct"
	"gpdf/stream/flate"

	ascii85filter "gpdf/stream/ascii85"
	lzwfilter "gpdf/stream/lzw"
)

// PDFWriter implements Writer for PDF 2.0 output.
type PDFWriter struct {
	filters stream.FilterRegistry
}

// NewPDFWriter returns a PDF writer with default stream filters registered.
func NewPDFWriter() *PDFWriter {
	reg := stream.NewRegistry()
	reg.Register("FlateDecode", flate.NewFilter())
	reg.Register("DCTDecode", dct.NewFilter())
	reg.Register("LZWDecode", lzwfilter.NewFilter())
	reg.Register("ASCII85Decode", ascii85filter.NewFilter())
	reg.Register("ASCIIHexDecode", asciihex.NewFilter())
	return &PDFWriter{filters: reg}
}

// NewPDFWriterWithFilters returns a PDF writer using the given filter registry.
func NewPDFWriterWithFilters(filters stream.FilterRegistry) *PDFWriter {
	return &PDFWriter{filters: filters}
}

// Write writes the document to w.
func (pw *PDFWriter) Write(w io.Writer, doc Document) error {
	const header = "%PDF-2.0\n%\xE2\xE3\xCF\xD3\n"
	objNums := doc.ObjectNumbers()
	sort.Ints(objNums)
	offsets := make(map[int]int64)
	var body bytes.Buffer
	for _, num := range objNums {
		ref := model.Ref{ObjectNumber: num, Generation: 0}
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos := int64(len(header)) + int64(body.Len())
		offsets[num] = pos
		if err := pw.writeIndirectObject(&body, num, 0, obj); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	xrefStart := int64(len(header)) + int64(body.Len())
	if err := pw.writeXRefTable(w, doc, objNums, offsets, -1); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "trailer\n"); err != nil {
		return err
	}
	if err := pw.writeDict(w, doc.Trailer().Dict); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nstartxref\n%d\n%%%%EOF\n", xrefStart); err != nil {
		return err
	}
	return nil
}

// WriteWithPassword writes the document encrypted with Standard handler (R=2) using user and owner passwords.
func (pw *PDFWriter) WriteWithPassword(w io.Writer, doc Document, userPassword, ownerPassword string) error {
	const header = "%PDF-2.0\n%\xE2\xE3\xCF\xD3\n"
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return fmt.Errorf("document has no objects")
	}
	sort.Ints(objNums)
	maxNum := objNums[len(objNums)-1]
	encryptNum := maxNum + 1
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return err
	}
	encDict, enc, err := security.BuildEncryptDictForWrite(userPassword, ownerPassword, id, -4)
	if err != nil {
		return err
	}
	trailerDict := copyTrailerWithEncrypt(doc.Trailer().Dict, encryptNum, id)
	offsets := make(map[int]int64)
	var body bytes.Buffer
	ref := model.Ref{ObjectNumber: 0, Generation: 0}
	for _, num := range objNums {
		ref.ObjectNumber = num
		ref.Generation = 0
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos := int64(len(header)) + int64(body.Len())
		offsets[num] = pos
		if err := pw.writeIndirectObjectEnc(&body, num, 0, obj, enc); err != nil {
			return err
		}
	}
	pos := int64(len(header)) + int64(body.Len())
	offsets[encryptNum] = pos
	if err := pw.writeIndirectObjectEnc(&body, encryptNum, 0, encDict, nil); err != nil {
		return err
	}
	allNums := append([]int{}, objNums...)
	allNums = append(allNums, encryptNum)
	sort.Ints(allNums)
	maxNum = allNums[len(allNums)-1]
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	xrefStart := int64(len(header)) + int64(body.Len())
	if err := pw.writeXRefTable(w, doc, nil, offsets, maxNum); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "trailer\n"); err != nil {
		return err
	}
	if err := pw.writeDict(w, trailerDict); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nstartxref\n%d\n%%%%EOF\n", xrefStart); err != nil {
		return err
	}
	return nil
}

// WriteWithAES256Password writes the document encrypted with AES-256 using a
// simplified Standard handler (V=5, R=6). This is intended as a stronger
// alternative to WriteWithPassword while keeping the API similar.
func (pw *PDFWriter) WriteWithAES256Password(w io.Writer, doc Document, userPassword, ownerPassword string) error {
	const header = "%PDF-2.0\n%\xE2\xE3\xCF\xD3\n"
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return fmt.Errorf("document has no objects")
	}
	sort.Ints(objNums)
	maxNum := objNums[len(objNums)-1]
	encryptNum := maxNum + 1
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return err
	}
	encDict, enc, err := security.BuildAES256EncryptDictForWrite(userPassword, ownerPassword, id, -4)
	if err != nil {
		return err
	}
	trailerDict := copyTrailerWithEncrypt(doc.Trailer().Dict, encryptNum, id)
	offsets := make(map[int]int64)
	var body bytes.Buffer
	ref := model.Ref{ObjectNumber: 0, Generation: 0}
	for _, num := range objNums {
		ref.ObjectNumber = num
		ref.Generation = 0
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos := int64(len(header)) + int64(body.Len())
		offsets[num] = pos
		if err := pw.writeIndirectObjectEnc(&body, num, 0, obj, enc); err != nil {
			return err
		}
	}
	pos := int64(len(header)) + int64(body.Len())
	offsets[encryptNum] = pos
	if err := pw.writeIndirectObjectEnc(&body, encryptNum, 0, encDict, nil); err != nil {
		return err
	}
	allNums := append([]int{}, objNums...)
	allNums = append(allNums, encryptNum)
	sort.Ints(allNums)
	maxNum = allNums[len(allNums)-1]
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	xrefStart := int64(len(header)) + int64(body.Len())
	if err := pw.writeXRefTable(w, doc, nil, offsets, maxNum); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "trailer\n"); err != nil {
		return err
	}
	if err := pw.writeDict(w, trailerDict); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nstartxref\n%d\n%%%%EOF\n", xrefStart); err != nil {
		return err
	}
	return nil
}

func copyTrailerWithEncrypt(d model.Dict, encryptNum int, id []byte) model.Dict {
	out := make(model.Dict, len(d)+2)
	for k, v := range d {
		out[k] = v
	}
	out[model.Name("Encrypt")] = model.Ref{ObjectNumber: encryptNum, Generation: 0}
	out[model.Name("ID")] = model.Array{model.String(string(id))}
	out[model.Name("Size")] = model.Integer(int64(encryptNum + 1))
	return out
}

func (pw *PDFWriter) writeIndirectObjectEnc(w io.Writer, num, gen int, obj model.Object, enc security.Encryptor) error {
	if _, err := fmt.Fprintf(w, "%d %d obj\n", num, gen); err != nil {
		return err
	}
	ref := model.Ref{ObjectNumber: num, Generation: gen}
	if err := pw.writeObjectEnc(w, obj, &ref, enc); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\nendobj\n"); err != nil {
		return err
	}
	return nil
}

func (pw *PDFWriter) writeObjectEnc(w io.Writer, obj model.Object, ref *model.Ref, enc security.Encryptor) error {
	if enc != nil && ref != nil {
		switch v := obj.(type) {
		case model.String:
			cipher, err := enc.EncryptString(*ref, []byte(v))
			if err != nil {
				return err
			}
			pw.writeHexString(w, cipher)
			return nil
		case *model.Stream:
			content := v.Content
			if pw.filters != nil {
				if filterObj := v.Dict[model.Name("Filter")]; filterObj != nil {
					if name, ok := filterObj.(model.Name); ok {
						if f := pw.filters.Get(string(name)); f != nil {
							var buf bytes.Buffer
							if err := f.Encode(&buf, bytes.NewReader(v.Content), string(name)); err == nil {
								content = buf.Bytes()
							}
						}
					}
				}
			}
			streamDict := make(model.Dict, len(v.Dict))
			for k, val := range v.Dict {
				streamDict[k] = val
			}
			streamDict[model.Name("Length")] = model.Integer(int64(len(content)))
			if err := pw.writeDictEnc(w, streamDict, ref, enc); err != nil {
				return err
			}
			encrypted, err := enc.EncryptStream(*ref, content)
			if err != nil {
				return err
			}
			io.WriteString(w, "\nstream\n")
			w.Write(encrypted)
			if len(encrypted) > 0 && encrypted[len(encrypted)-1] != '\n' {
				io.WriteString(w, "\n")
			}
			io.WriteString(w, "endstream")
			return nil
		case model.Dict:
			return pw.writeDictEnc(w, v, ref, enc)
		case model.Array:
			io.WriteString(w, "[")
			for i, e := range v {
				if i > 0 {
					io.WriteString(w, " ")
				}
				if err := pw.writeObjectEnc(w, e, ref, enc); err != nil {
					return err
				}
			}
			io.WriteString(w, "]")
			return nil
		}
	}
	return pw.writeObject(w, obj)
}

func (pw *PDFWriter) writeDictEnc(w io.Writer, d model.Dict, ref *model.Ref, enc security.Encryptor) error {
	io.WriteString(w, "<<")
	for k, v := range d {
		io.WriteString(w, "\n/")
		io.WriteString(w, escapeName(string(k)))
		io.WriteString(w, " ")
		if err := pw.writeObjectEnc(w, v, ref, enc); err != nil {
			return err
		}
	}
	io.WriteString(w, "\n>>")
	return nil
}

func (pw *PDFWriter) writeHexString(w io.Writer, b []byte) {
	io.WriteString(w, "<")
	io.WriteString(w, hex.EncodeToString(b))
	io.WriteString(w, ">")
}

func (pw *PDFWriter) writeXRefTable(w io.Writer, doc Document, objNums []int, offsets map[int]int64, maxNum int) error {
	if _, err := io.WriteString(w, "xref\n"); err != nil {
		return err
	}
	if maxNum < 0 && len(objNums) > 0 {
		maxNum = objNums[len(objNums)-1]
	}
	if maxNum < 0 {
		io.WriteString(w, "0 0\n")
		return nil
	}
	if _, err := fmt.Fprintf(w, "0 %d\n", maxNum+1); err != nil {
		return err
	}
	for i := 0; i <= maxNum; i++ {
		if off, ok := offsets[i]; ok {
			fmt.Fprintf(w, "%010d %05d n \n", off, 0)
		} else {
			fmt.Fprintf(w, "%010d %05d f \n", 0, 65535)
		}
	}
	return nil
}

// WriteIncremental appends an incremental update (new objects + xref + trailer with /Prev + startxref + %%EOF) to w.
func (pw *PDFWriter) WriteIncremental(w io.Writer, appendOffset int64, baseStartXRef int64, doc Document) error {
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return fmt.Errorf("incremental update requires at least one object")
	}
	sort.Ints(objNums)
	offsets := make(map[int]int64)
	var body bytes.Buffer
	for _, num := range objNums {
		ref := model.Ref{ObjectNumber: num, Generation: 0}
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos := appendOffset + int64(body.Len())
		offsets[num] = pos
		if err := pw.writeIndirectObject(&body, num, 0, obj); err != nil {
			return err
		}
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	xrefStart := appendOffset + int64(body.Len())
	if err := pw.writeXRefSubsection(w, objNums, offsets); err != nil {
		return err
	}
	trailerDict := copyTrailerWithPrev(doc.Trailer().Dict, baseStartXRef, objNums)
	if _, err := io.WriteString(w, "trailer\n"); err != nil {
		return err
	}
	if err := pw.writeDict(w, trailerDict); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nstartxref\n%d\n%%%%EOF\n", xrefStart); err != nil {
		return err
	}
	return nil
}

func copyTrailerWithPrev(d model.Dict, prev int64, objNums []int) model.Dict {
	out := make(model.Dict, len(d)+2)
	for k, v := range d {
		out[k] = v
	}
	out[model.Name("Prev")] = model.Integer(prev)
	max := 0
	for _, n := range objNums {
		if n > max {
			max = n
		}
	}
	out[model.Name("Size")] = model.Integer(int64(max + 1))
	return out
}

func (pw *PDFWriter) writeXRefSubsection(w io.Writer, objNums []int, offsets map[int]int64) error {
	if len(objNums) == 0 {
		return nil
	}
	first := objNums[0]
	max := objNums[len(objNums)-1]
	count := max - first + 1
	if _, err := io.WriteString(w, "xref\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d %d\n", first, count); err != nil {
		return err
	}
	for i := first; i <= max; i++ {
		if off, ok := offsets[i]; ok {
			fmt.Fprintf(w, "%010d %05d n \n", off, 0)
		} else {
			fmt.Fprintf(w, "%010d %05d f \n", 0, 65535)
		}
	}
	return nil
}

func (pw *PDFWriter) writeIndirectObject(w io.Writer, num, gen int, obj model.Object) error {
	if _, err := fmt.Fprintf(w, "%d %d obj\n", num, gen); err != nil {
		return err
	}
	if err := pw.writeObject(w, obj); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\nendobj\n"); err != nil {
		return err
	}
	return nil
}

func (pw *PDFWriter) writeObject(w io.Writer, obj model.Object) error {
	switch v := obj.(type) {
	case model.Null:
		io.WriteString(w, "null")
	case model.Boolean:
		if v {
			io.WriteString(w, "true")
		} else {
			io.WriteString(w, "false")
		}
	case model.Integer:
		fmt.Fprintf(w, "%d", v)
	case model.Real:
		fmt.Fprintf(w, "%g", v)
	case model.String:
		pw.writeString(w, string(v))
	case model.Name:
		fmt.Fprintf(w, "/%s", escapeName(string(v)))
	case model.Ref:
		fmt.Fprintf(w, "%d %d R", v.ObjectNumber, v.Generation)
	case model.Array:
		io.WriteString(w, "[")
		for i, e := range v {
			if i > 0 {
				io.WriteString(w, " ")
			}
			pw.writeObject(w, e)
		}
		io.WriteString(w, "]")
	case model.Dict:
		return pw.writeDict(w, v)
	case *model.Stream:
		content := v.Content
		if pw.filters != nil {
			if filterObj := v.Dict[model.Name("Filter")]; filterObj != nil {
				if name, ok := filterObj.(model.Name); ok {
					if f := pw.filters.Get(string(name)); f != nil {
						var enc bytes.Buffer
						if err := f.Encode(&enc, bytes.NewReader(v.Content), string(name)); err == nil {
							content = enc.Bytes()
						}
					}
				}
			}
		}
		streamDict := make(model.Dict, len(v.Dict))
		for k, val := range v.Dict {
			streamDict[k] = val
		}
		streamDict[model.Name("Length")] = model.Integer(int64(len(content)))
		if err := pw.writeDict(w, streamDict); err != nil {
			return err
		}
		io.WriteString(w, "\nstream\n")
		w.Write(content)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			io.WriteString(w, "\n")
		}
		io.WriteString(w, "endstream")
	default:
		return fmt.Errorf("unknown object type %T", obj)
	}
	return nil
}

func (pw *PDFWriter) writeString(w io.Writer, s string) {
	io.WriteString(w, "(")
	for i := range len(s) {
		b := s[i]
		switch b {
		case '\\', '(', ')':
			fmt.Fprintf(w, "\\%c", b)
		case '\n':
			io.WriteString(w, "\\n")
		case '\r':
			io.WriteString(w, "\\r")
		case '\t':
			io.WriteString(w, "\\t")
		case '\b':
			io.WriteString(w, "\\b")
		case '\f':
			io.WriteString(w, "\\f")
		default:
			if b < 0x20 || b > 0x7e {
				fmt.Fprintf(w, "\\%03o", b)
				continue
			}
			_, _ = w.Write([]byte{b})
		}
	}
	io.WriteString(w, ")")
}

func escapeName(s string) string {
	var b bytes.Buffer
	for _, c := range s {
		if c <= ' ' || c >= 127 || c == '#' || c == '/' || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '%' {
			fmt.Fprintf(&b, "#%02x", c)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func (pw *PDFWriter) writeDict(w io.Writer, d model.Dict) error {
	io.WriteString(w, "<<")
	for k, v := range d {
		io.WriteString(w, "\n/")
		io.WriteString(w, escapeName(string(k)))
		io.WriteString(w, " ")
		pw.writeObject(w, v)
	}
	io.WriteString(w, "\n>>")
	return nil
}

package reader

import (
	"bytes"
	"fmt"
	"testing"

	"gpdf/model"
	"gpdf/syntax/impl"
)

func minimalPDFBytes() []byte {
	header := []byte("%PDF-2.0\n")
	body := []byte(`1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /MediaBox [0 0 595 842] >>
endobj
`)
	// Object starts: 1 at len(header), 2 and 3 follow
	off1 := int64(len(header))
	off2 := int64(len(header)) + int64(bytes.Index(body, []byte("2 0 obj")))
	off3 := int64(len(header)) + int64(bytes.Index(body, []byte("3 0 obj")))
	xrefLines := fmt.Sprintf("xref\n0 4\n0000000000 65535 f \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n",
		off1, off2, off3)
	trailer := "<< /Size 4 /Root 1 0 R >>\nstartxref\n"
	xrefStart := int64(len(header)) + int64(len(body)) + int64(bytes.Index([]byte(xrefLines), []byte("xref")))
	trailer += fmt.Sprintf("%d\n%%%%EOF\n", xrefStart)
	return []byte(string(header) + string(body) + xrefLines + "trailer\n" + trailer)
}

// TestParseObject1FromOffset9 verifies that parsing from offset 9 of the minimal PDF yields object 1 (catalog).
func TestParseObject1FromOffset9(t *testing.T) {
	minimalPDF := minimalPDFBytes()
	if minimalPDF[9] != '1' {
		t.Fatalf("expected '1' at offset 9, got %q", minimalPDF[9])
	}
	r := bytes.NewReader(minimalPDF)
	p := impl.NewParser(r, int64(len(minimalPDF)))
	if err := p.SetPosition(9); err != nil {
		t.Fatal(err)
	}
	indirect, inline, err := p.ParseObject()
	if err != nil {
		t.Fatal(err)
	}
	if indirect == nil {
		t.Fatalf("expected indirect object, got inline %v", inline)
	}
	if indirect.ObjectNumber != 1 {
		t.Errorf("expected object 1, got %d", indirect.ObjectNumber)
	}
	dict, ok := indirect.Value.(model.Dict)
	if !ok {
		t.Fatalf("expected catalog dict, got %T", indirect.Value)
	}
	if dict[model.Name("Type")] != model.Name("Catalog") {
		t.Errorf("expected /Type /Catalog, got %v", dict[model.Name("Type")])
	}
}

// TestParseObject3FromOff3 verifies that parsing object 3 from its computed offset yields the page dict.
func TestParseObject3FromOff3(t *testing.T) {
	minimalPDF := minimalPDFBytes()
	header := []byte("%PDF-2.0\n")
	body := minimalPDF[len(header):][:bytes.Index(minimalPDF[len(header):], []byte("xref"))]
	off3 := int64(len(header)) + int64(bytes.Index(body, []byte("3 0 obj")))
	if off3 >= int64(len(minimalPDF)) || minimalPDF[off3] != '3' {
		t.Fatalf("off3=%d byte=%q", off3, safeByte(minimalPDF, off3))
	}
	r := bytes.NewReader(minimalPDF)
	p := impl.NewParser(r, int64(len(minimalPDF)))
	if err := p.SetPosition(off3); err != nil {
		t.Fatal(err)
	}
	indirect, _, err := p.ParseObject()
	if err != nil {
		t.Fatalf("ParseObject at off3=%d: %v", off3, err)
	}
	if indirect == nil || indirect.ObjectNumber != 3 {
		t.Fatalf("expected indirect object 3, got %+v", indirect)
	}
	dict, ok := indirect.Value.(model.Dict)
	if !ok {
		t.Fatalf("expected page dict, got %T", indirect.Value)
	}
	if dict[model.Name("Type")] != model.Name("Page") {
		t.Fatalf("expected /Type /Page, got %v", dict[model.Name("Type")])
	}
}

// TestMinimalPDFOffsets verifies that object offsets in the minimal PDF point to "N 0 obj".
func TestMinimalPDFOffsets(t *testing.T) {
	minimalPDF := minimalPDFBytes()
	header := []byte("%PDF-2.0\n")
	body := minimalPDF[len(header) : len(header)+bytes.Index(minimalPDF[len(header):], []byte("xref"))]
	off1 := int64(len(header))
	off2 := off1 + int64(bytes.Index(body, []byte("2 0 obj")))
	off3 := off1 + int64(bytes.Index(body, []byte("3 0 obj")))
	if off1 >= int64(len(minimalPDF)) || minimalPDF[off1] != '1' {
		t.Fatalf("off1=%d byte=%q", off1, safeByte(minimalPDF, off1))
	}
	if off2 >= int64(len(minimalPDF)) || minimalPDF[off2] != '2' {
		t.Fatalf("off2=%d byte=%q", off2, safeByte(minimalPDF, off2))
	}
	if off3 >= int64(len(minimalPDF)) || minimalPDF[off3] != '3' {
		t.Fatalf("off3=%d byte=%q (at 115: %q)", off3, safeByte(minimalPDF, off3), safeByte(minimalPDF, 115))
	}
}

func safeByte(b []byte, i int64) byte {
	if i < 0 || int(i) >= len(b) {
		return 0
	}
	return b[i]
}

func TestReadDocument_Minimal(t *testing.T) {
	minimalPDF := minimalPDFBytes()
	// Sanity: object 1 must start at offset 9
	if minimalPDF[9] != '1' {
		t.Fatalf("expected '1' at offset 9, got %q", minimalPDF[9])
	}
	r := bytes.NewReader(minimalPDF)
	rr := NewPDFReader()
	doc, err := rr.ReadDocument(r, int64(len(minimalPDF)))
	if err != nil {
		t.Fatal(err)
	}
	// Verify xref entry 1 has offset 9 before resolving
	pdfDoc, ok := doc.(*pdfDocument)
	if !ok {
		t.Fatal("expected *pdfDocument")
	}
	e, ok := pdfDoc.xref.Get(1)
	if !ok || !e.InUse || e.Offset != 9 {
		t.Fatalf("xref entry 1: ok=%v InUse=%v Offset=%d (want 9)", ok, e.InUse, e.Offset)
	}
	cat, err := doc.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if cat == nil {
		t.Fatal("expected catalog")
	}
	pages, err := doc.Pages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
}

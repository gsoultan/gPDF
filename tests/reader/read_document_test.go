package reader_test

import (
	"bytes"
	"fmt"
	"testing"

	"gpdf/doc"
	"gpdf/model"
	"gpdf/reader"
)

func TestReadDocumentWithPasswordDecryptsInfoStrings(t *testing.T) {
	document, err := doc.New().
		Title("Encrypted Title").
		Author("Alice").
		Subject("Secret").
		AddPage().
		Build()
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	defer document.Close()

	var buf bytes.Buffer
	if err := document.SaveWithPassword(&buf, "user", "owner"); err != nil {
		t.Fatalf("SaveWithPassword returned error: %v", err)
	}

	parsed, err := reader.NewPDFReader().ReadDocumentWithPassword(
		bytes.NewReader(buf.Bytes()),
		int64(buf.Len()),
		"user",
	)
	if err != nil {
		t.Fatalf("ReadDocumentWithPassword returned error: %v", err)
	}

	info, err := parsed.Info()
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil Info dictionary")
	}
	if got := stringFromInfo(info, "Title"); got != "Encrypted Title" {
		t.Fatalf("Title mismatch: got %q", got)
	}
	if got := stringFromInfo(info, "Author"); got != "Alice" {
		t.Fatalf("Author mismatch: got %q", got)
	}
	if got := stringFromInfo(info, "Subject"); got != "Secret" {
		t.Fatalf("Subject mismatch: got %q", got)
	}
}

func TestReadDocumentRejectsMismatchedXRefObject(t *testing.T) {
	pdf, offsets := minimalPDFBytes(t)
	original := fmt.Sprintf("%010d 00000 n ", offsets[1])
	corrupt := fmt.Sprintf("%010d 00000 n ", offsets[3])
	pdf = bytes.Replace(pdf, []byte(original), []byte(corrupt), 1)

	parsed, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("ReadDocument returned error: %v", err)
	}

	if _, err := parsed.Catalog(); err == nil {
		t.Fatal("expected Catalog to fail when xref points at the wrong indirect object")
	}
}

func TestReadDocumentFindsStartXRefBeyondTailWindow(t *testing.T) {
	pdf, _ := minimalPDFBytes(t)
	pdf = append(pdf, bytes.Repeat([]byte(" "), 5000)...)

	parsed, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("ReadDocument returned error: %v", err)
	}

	pages, err := parsed.Pages()
	if err != nil {
		t.Fatalf("Pages returned error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
}

func minimalPDFBytes(t *testing.T) ([]byte, map[int]int64) {
	t.Helper()

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

	offsets := map[int]int64{
		1: int64(len(header)),
		2: int64(len(header)) + int64(bytes.Index(body, []byte("2 0 obj"))),
		3: int64(len(header)) + int64(bytes.Index(body, []byte("3 0 obj"))),
	}

	xrefLines := fmt.Sprintf(
		"xref\n0 4\n0000000000 65535 f \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n",
		offsets[1],
		offsets[2],
		offsets[3],
	)
	xrefStart := int64(len(header)) + int64(len(body)) + int64(bytes.Index([]byte(xrefLines), []byte("xref")))
	trailer := fmt.Sprintf("trailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefStart)

	return append(append(append([]byte{}, header...), body...), []byte(xrefLines+trailer)...), offsets
}

func stringFromInfo(info model.Dict, key string) string {
	value, ok := info[model.Name(key)].(model.String)
	if !ok {
		return ""
	}
	return string(value)
}

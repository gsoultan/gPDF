package doc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

func TestReadContent_OpenReadContentClose(t *testing.T) {
	// Build a PDF with known text
	d, err := doc.New().
		Title("ReadContent Test").
		AddPage().
		DrawText("Hello, PDF world!", 72, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	d.Close()

	// Write to temp file so we can use doc.Open
	tmp := filepath.Join(t.TempDir(), "readcontent.pdf")
	if err := os.WriteFile(tmp, buf.Bytes(), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Open -> ReadContent -> Close
	doc, err := doc.Open(tmp)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	text, err := doc.ReadContent()
	if err != nil {
		t.Fatalf("ReadContent: %v", err)
	}

	if !strings.Contains(text, "Hello, PDF world!") {
		t.Errorf("ReadContent: got %q, want to contain %q", text, "Hello, PDF world!")
	}
}

func TestReadContent_BuiltDocument(t *testing.T) {
	d, err := doc.New().
		Title("Built").
		AddPage().
		DrawText("Built doc text", 72, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer d.Close()

	text, err := d.ReadContent()
	if err != nil {
		t.Fatalf("ReadContent: %v", err)
	}

	if !strings.Contains(text, "Built doc text") {
		t.Errorf("ReadContent: got %q, want to contain %q", text, "Built doc text")
	}
}

func TestSearch_OpenSearchClose(t *testing.T) {
	d, err := doc.New().
		Title("Search Test").
		AddPage().
		DrawText("Alpha and Beta", 72, 700, "Helvetica", 12).
		AddPage().
		DrawText("Beta and Gamma", 72, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	d.Close()

	tmp := filepath.Join(t.TempDir(), "search.pdf")
	if err := os.WriteFile(tmp, buf.Bytes(), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	pdf, err := doc.Open(tmp)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pdf.Close()

	results, err := pdf.Search("Alpha", "Beta", "Gamma", "Missing")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("Search: expected 4 results, got %d", len(results))
	}
	if results[0].Keyword != "Alpha" || len(results[0].Pages) != 1 || results[0].Pages[0] != 0 {
		t.Errorf("Alpha: got %+v", results[0])
	}
	if idx := results[0].Indices[0]; len(idx) != 1 || idx[0] != 0 {
		t.Errorf("Alpha indices on page 0: got %v, want [0]", results[0].Indices[0])
	}
	if results[1].Keyword != "Beta" || len(results[1].Pages) != 2 {
		t.Errorf("Beta: got %+v, want pages [0,1]", results[1])
	}
	if idx := results[1].Indices[0]; len(idx) != 1 || idx[0] != 10 {
		t.Errorf("Beta indices on page 0: got %v, want [10] (after 'Alpha and ')", results[1].Indices[0])
	}
	if idx := results[1].Indices[1]; len(idx) != 1 || idx[0] != 0 {
		t.Errorf("Beta indices on page 1: got %v, want [0]", results[1].Indices[1])
	}
	if results[2].Keyword != "Gamma" || len(results[2].Pages) != 1 || results[2].Pages[0] != 1 {
		t.Errorf("Gamma: got %+v", results[2])
	}
	if results[3].Keyword != "Missing" || len(results[3].Pages) != 0 || results[3].Indices == nil {
		t.Errorf("Missing: got %+v, want no pages", results[3])
	}
}

func TestReplace_OpenReplaceSaveClose(t *testing.T) {
	d, err := doc.New().
		Title("Replace Test").
		AddPage().
		DrawText("Hello world", 72, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	d.Close()

	tmp := filepath.Join(t.TempDir(), "replace.pdf")
	if err := os.WriteFile(tmp, buf.Bytes(), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	pdf, err := doc.Open(tmp)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pdf.Close()

	if err := pdf.Replace("world", "PDF"); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	text, err := pdf.ReadContent()
	if err != nil {
		t.Fatalf("ReadContent: %v", err)
	}
	if !strings.Contains(text, "Hello PDF") || strings.Contains(text, "world") {
		t.Errorf("Replace: got %q", text)
	}

	outPath := filepath.Join(t.TempDir(), "replaced.pdf")
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := pdf.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out.Close()

	pdf2, err := doc.Open(outPath)
	if err != nil {
		t.Fatalf("Open replaced: %v", err)
	}
	defer pdf2.Close()
	text2, _ := pdf2.ReadContent()
	if !strings.Contains(text2, "Hello PDF") {
		t.Errorf("After save/reopen: got %q", text2)
	}
}

func TestReplaces_Map(t *testing.T) {
	d, err := doc.New().
		Title("Replaces Test").
		AddPage().
		DrawText("Alpha Beta Gamma", 72, 700, "Helvetica", 12).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer d.Close()

	if err := d.Replaces(map[string]string{"Alpha": "A", "Gamma": "C"}); err != nil {
		t.Fatalf("Replaces: %v", err)
	}

	text, err := d.ReadContent()
	if err != nil {
		t.Fatalf("ReadContent: %v", err)
	}
	expected := "A Beta C"
	if !strings.Contains(text, "A Beta C") {
		t.Errorf("Replaces: got %q, want to contain %q", text, expected)
	}
}

package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

func TestMerge(t *testing.T) {
	// Create two simple PDFs to merge
	pdf1Path := "test1.pdf"
	pdf2Path := "test2.pdf"
	mergedPath := "merged.pdf"

	defer os.Remove(pdf1Path)
	defer os.Remove(pdf2Path)
	defer os.Remove(mergedPath)

	// Create PDF 1
	b1 := doc.New().Title("PDF 1").AddPage()
	// Add some content to b1 if possible, but let's keep it simple first
	d1, err := b1.Build()
	if err != nil {
		t.Fatalf("Build PDF 1: %v", err)
	}
	f1, _ := os.Create(pdf1Path)
	d1.Save(f1)
	f1.Close()
	d1.Close()

	// Create PDF 2
	b2 := doc.New().Title("PDF 2").AddPage().AddPage()
	d2, err := b2.Build()
	if err != nil {
		t.Fatalf("Build PDF 2: %v", err)
	}
	f2, _ := os.Create(pdf2Path)
	d2.Save(f2)
	f2.Close()
	d2.Close()

	// Re-open them for merging
	doc1, err := doc.Open(pdf1Path)
	if err != nil {
		t.Fatalf("Open PDF 1: %v", err)
	}
	defer doc1.Close()

	doc2, err := doc.Open(pdf2Path)
	if err != nil {
		t.Fatalf("Open PDF 2: %v", err)
	}
	defer doc2.Close()

	// Attempt to merge
	// Note: doc.Merge doesn't exist yet, this should fail to compile
	mergedDoc, err := doc.Merge(doc1, doc2)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	defer mergedDoc.Close()

	// Verify merged document
	pages, err := mergedDoc.Pages()
	if err != nil {
		t.Fatalf("Pages: %v", err)
	}

	if len(pages) != 3 {
		t.Errorf("Expected 3 pages, got %d", len(pages))
	}

	// Save merged document
	var buf bytes.Buffer
	if err := mergedDoc.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Saved PDF is empty")
	}
}

func TestMergeThree(t *testing.T) {
	// Create three documents
	b1 := doc.New().Title("PDF 1").AddPage()
	d1, _ := b1.Build()
	defer d1.Close()

	b2 := doc.New().Title("PDF 2").AddPage().AddPage()
	d2, _ := b2.Build()
	defer d2.Close()

	b3 := doc.New().Title("PDF 3").AddPage()
	d3, _ := b3.Build()
	defer d3.Close()

	// Merge all three
	merged, err := doc.Merge(d1, d2, d3)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	defer merged.Close()

	pages, _ := merged.Pages()
	if len(pages) != 4 {
		t.Errorf("Expected 4 pages, got %d", len(pages))
	}
}

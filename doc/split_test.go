package doc_test

import (
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

func TestSplit(t *testing.T) {
	// 1. Create a 3-page document
	builder := doc.New()
	builder.AddPage() // Page 1
	builder.AddPage() // Page 2
	builder.AddPage() // Page 3
	src, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build source document: %v", err)
	}

	pages, err := src.Pages()
	if err != nil || len(pages) != 3 {
		t.Fatalf("Source document should have 3 pages, got %d (err: %v)", len(pages), err)
	}

	// 2. Test Split by ranges
	// Range 1: [0, 2] (Pages 1 and 3)
	// Range 2: [1] (Page 2)
	ranges := [][]int{{0, 2}, {1}}
	splitDocs, err := doc.Split(src, ranges...)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(splitDocs) != 2 {
		t.Fatalf("Expected 2 split documents, got %d", len(splitDocs))
	}

	// Check first split doc (Pages 1 and 3)
	pages1, err := splitDocs[0].Pages()
	if err != nil || len(pages1) != 2 {
		t.Errorf("First split doc should have 2 pages, got %d (err: %v)", len(pages1), err)
	}

	// Check second split doc (Page 2)
	pages2, err := splitDocs[1].Pages()
	if err != nil || len(pages2) != 1 {
		t.Errorf("Second split doc should have 1 page, got %d (err: %v)", len(pages2), err)
	}

	// 3. Test SplitEvery(1)
	splitEvery1, err := doc.SplitEvery(src, 1)
	if err != nil {
		t.Fatalf("SplitEvery(1) failed: %v", err)
	}
	if len(splitEvery1) != 3 {
		t.Errorf("SplitEvery(1) should return 3 documents, got %d", len(splitEvery1))
	}

	// 4. Test Extract(1, 3) -> Pages 2 and 3
	extracted, err := doc.Extract(src, 1, 3)
	if err != nil {
		t.Fatalf("Extract(1, 3) failed: %v", err)
	}
	pagesExt, err := extracted.Pages()
	if err != nil || len(pagesExt) != 2 {
		t.Errorf("Extracted doc should have 2 pages, got %d (err: %v)", len(pagesExt), err)
	}
}

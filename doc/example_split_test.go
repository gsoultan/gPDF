package doc_test

import (
	"fmt"
	"log"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

func ExampleSplit() {
	// 1. Open a document to split
	src, err := doc.Open("input.pdf")
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer src.Close()

	// 2. Define page ranges for the new documents
	// Each range is a slice of 0-based page indices.
	// Here, we split into two documents:
	// - doc1: pages 0 and 2 (first and third page)
	// - doc2: pages 1 (second page)
	ranges := [][]int{
		{0, 2},
		{1},
	}

	// 3. Perform the split
	splitDocs, err := doc.Split(src, ranges...)
	if err != nil {
		log.Fatalf("Split failed: %v", err)
	}

	// 4. Save each split document
	for i, splitDoc := range splitDocs {
		defer splitDoc.Close()

		outputPath := fmt.Sprintf("split_%d.pdf", i)
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}

		if err := splitDoc.Save(f); err != nil {
			f.Close()
			log.Fatalf("Failed to save split PDF: %v", err)
		}
		f.Close()
	}
}

func ExampleSplitEvery() {
	// 1. Open a document to split
	src, err := doc.Open("input.pdf")
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer src.Close()

	// 2. Split into multiple documents, each containing at most 2 pages
	splitDocs, err := doc.SplitEvery(src, 2)
	if err != nil {
		log.Fatalf("SplitEvery failed: %v", err)
	}

	// 3. Save each split document
	for i, splitDoc := range splitDocs {
		defer splitDoc.Close()

		outputPath := fmt.Sprintf("chunk_%d.pdf", i)
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}

		if err := splitDoc.Save(f); err != nil {
			f.Close()
			log.Fatalf("Failed to save split PDF: %v", err)
		}
		f.Close()
	}
}

func ExampleExtract() {
	// 1. Open a document
	src, err := doc.Open("input.pdf")
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer src.Close()

	// 2. Extract pages 1-3 (inclusive-exclusive range: [1, 3))
	// This will extract the second and third pages of the document.
	extracted, err := doc.Extract(src, 1, 3)
	if err != nil {
		log.Fatalf("Extract failed: %v", err)
	}
	defer extracted.Close()

	// 3. Save the extracted pages to a new document
	f, err := os.Create("extracted.pdf")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := extracted.Save(f); err != nil {
		log.Fatalf("Failed to save extracted PDF: %v", err)
	}
}

package doc_test

import (
	"log"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

func ExampleMerge() {
	// Open the first document
	doc1, err := doc.Open("file1.pdf")
	if err != nil {
		log.Fatalf("Failed to open file1.pdf: %v", err)
	}
	defer doc1.Close()

	// Open the second document
	doc2, err := doc.Open("file2.pdf")
	if err != nil {
		log.Fatalf("Failed to open file2.pdf: %v", err)
	}
	defer doc2.Close()

	// Merge the documents
	merged, err := doc.Merge(doc1, doc2)
	if err != nil {
		log.Fatalf("Failed to merge documents: %v", err)
	}
	defer merged.Close()

	// Save the merged document to a new file
	f, err := os.Create("merged.pdf")
	if err != nil {
		log.Fatalf("Failed to create merged.pdf: %v", err)
	}
	defer f.Close()

	if err := merged.Save(f); err != nil {
		log.Fatalf("Failed to save merged document: %v", err)
	}
}

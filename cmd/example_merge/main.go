// Example merge demonstrates how to merge multiple PDF files into one.
//
// Usage:
//
//	go run ./cmd/example_merge output.pdf input1.pdf input2.pdf ...
package main

import (
	"fmt"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run ./cmd/example_merge output.pdf input1.pdf input2.pdf ...")
		fmt.Println("\nAt least two input files are required.")
		os.Exit(1)
	}

	outputPath := os.Args[1]
	inputPaths := os.Args[2:]

	// 1. Open all input documents
	var docs []doc.Document
	for _, path := range inputPaths {
		d, err := doc.Open(path)
		if err != nil {
			fmt.Printf("Failed to open %s: %v\n", path, err)
			cleanup(docs)
			os.Exit(1)
		}
		docs = append(docs, d)
	}
	defer cleanup(docs)

	fmt.Printf("Merging %d files...\n", len(docs))

	// 2. Perform the merge
	merged, err := doc.Merge(docs...)
	if err != nil {
		fmt.Printf("Merge failed: %v\n", err)
		os.Exit(1)
	}
	defer merged.Close()

	// 3. Save the result
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := merged.Save(f); err != nil {
		fmt.Printf("Failed to save merged PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully merged into %s\n", outputPath)
}

func cleanup(docs []doc.Document) {
	for _, d := range docs {
		d.Close()
	}
}

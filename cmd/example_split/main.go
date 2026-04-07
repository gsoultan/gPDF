// Example split demonstrates how to split a PDF file into multiple smaller files.
//
// Usage:
//
//	go run ./cmd/example_split input.pdf output_prefix pages_per_file
//
// Example:
//
//	go run ./cmd/example_split sample.pdf split_ 1
//
// This will split sample.pdf into files with 1 page each (split_0.pdf, split_1.pdf, ...).
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gsoultan/gpdf/doc"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run ./cmd/example_split input.pdf output_prefix pages_per_file")
		fmt.Println("\nExample: go run ./cmd/example_split sample.pdf split_ 1")
		fmt.Println("This will split sample.pdf into files with 1 page each (split_0.pdf, split_1.pdf, ...)")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPrefix := os.Args[2]
	pagesPerFile, err := strconv.Atoi(os.Args[3])
	if err != nil || pagesPerFile <= 0 {
		fmt.Printf("Invalid pages_per_file: %s. Must be a positive integer.\n", os.Args[3])
		os.Exit(1)
	}

	// 1. Open the source document
	src, err := doc.Open(inputPath)
	if err != nil {
		fmt.Printf("Failed to open %s: %v\n", inputPath, err)
		os.Exit(1)
	}
	defer src.Close()

	fmt.Printf("Splitting %s into files with %d pages each...\n", inputPath, pagesPerFile)

	// 2. Perform the split
	// SplitEvery splits the document into multiple documents, each containing up to 'n' pages.
	splitDocs, err := doc.SplitEvery(src, pagesPerFile)
	if err != nil {
		fmt.Printf("Split failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Save each split document to a new file
	for i, splitDoc := range splitDocs {
		defer splitDoc.Close()

		outputPath := fmt.Sprintf("%s%d.pdf", outputPrefix, i)
		f, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Failed to create output file %s: %v\n", outputPath, err)
			continue
		}

		if err := splitDoc.Save(f); err != nil {
			fmt.Printf("Failed to save split PDF %s: %v\n", outputPath, err)
			f.Close()
			continue
		}
		f.Close()
		fmt.Printf("Saved: %s\n", outputPath)
	}

	fmt.Printf("Successfully split into %d files\n", len(splitDocs))
}

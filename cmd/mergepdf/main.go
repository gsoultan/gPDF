package main

import (
	"fmt"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <output.pdf> <input1.pdf> <input2.pdf> [input3.pdf ...]\n", os.Args[0])
		os.Exit(1)
	}

	outputPath := os.Args[1]
	inputPaths := os.Args[2:]

	var docs []doc.Document
	for _, path := range inputPaths {
		d, err := doc.Open(path)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", path, err)
			cleanup(docs)
			os.Exit(1)
		}
		docs = append(docs, d)
	}
	defer cleanup(docs)

	fmt.Printf("Merging %d files into %s...\n", len(docs), outputPath)
	merged, err := doc.Merge(docs...)
	if err != nil {
		fmt.Printf("Error merging: %v\n", err)
		os.Exit(1)
	}
	defer merged.Close()

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := merged.Save(f); err != nil {
		fmt.Printf("Error saving merged PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success!")
}

func cleanup(docs []doc.Document) {
	for _, d := range docs {
		d.Close()
	}
}

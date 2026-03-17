// readcontent demonstrates Open → ReadContent → Close using a sample PDF.
//
// Usage (run from repo root so "CV CONTOH.pdf" is found):
//
//	go run ./cmd/readcontent
//	go run ./cmd/readcontent "CV CONTOH.pdf"
//	go run ./cmd/readcontent ./path/to/file.pdf
package main

import (
	"fmt"
	"os"

	"gpdf/doc"
)

func main() {
	path := "CV CONTOH.pdf"
	if len(os.Args) >= 2 {
		path = os.Args[1]
	}

	d, err := doc.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open %q: %v\n", path, err)
		os.Exit(1)
	}
	defer d.Close()

	// ReadContent: all text from the document
	text, err := d.ReadContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadContent: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== ReadContent ===")
	if text == "" {
		fmt.Println("(no text extracted)")
	} else {
		// Print first 2000 runes so the output is bounded
		const maxPreview = 2000
		if len(text) > maxPreview {
			fmt.Println(text[:maxPreview])
			fmt.Printf("\n... (%d more characters)\n", len(text)-maxPreview)
		} else {
			fmt.Println(text)
		}
	}

	// Example: search for keywords
	keywords := []string{"email", "experience", "skill"}
	if len(keywords) > 0 {
		results, err := d.Search(keywords...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Search: %v\n", err)
			return
		}
		fmt.Println("\n=== Search ===")
		for _, r := range results {
			if len(r.Pages) == 0 {
				fmt.Printf("%q: not found\n", r.Keyword)
				continue
			}
			fmt.Printf("%q: pages %v", r.Keyword, r.Pages)
			if len(r.Indices) > 0 {
				fmt.Printf(" indices %v", r.Indices)
			}
			fmt.Println()
		}
	}
}

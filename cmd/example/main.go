// Example demonstrates gPDF: open, read, new, and save.
//
// Usage:
//
//	go run ./cmd/example [output.pdf]
//
// If output.pdf is given, creates a new PDF and saves it there.
// Otherwise prints a short usage message.
package main

import (
	"fmt"
	"os"

	"gpdf/doc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/example <output.pdf>")
		fmt.Println("Creates a new PDF with title, author, and two pages, then saves to output.pdf.")
		os.Exit(1)
	}
	path := os.Args[1]

	// New: construct a PDF with the fluent API (optional DrawText on first page)
	d, err := doc.New().
		Title("gPDF Example").
		Author("gPDF").
		PageSize(595, 842).
		AddPage().
		DrawText("Hello, PDF 2.0!", 100, 700, "Helvetica", 14).
		AddPage().
		Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer d.Close()

	// Read: inspect before saving
	catalog, _ := d.Catalog()
	if catalog != nil {
		fmt.Println("Catalog OK")
	}
	pages, _ := d.Pages()
	fmt.Printf("Pages: %d\n", len(pages))

	// Save: write to file
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := d.Save(f); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Saved: %s\n", path)
}

package main

import (
	"fmt"
	"os"

	"gpdf/doc"
)

// This example builds a "quotation" style PDF document
// demonstrating headings, a simple table layout, and an embedded image.
//
// Usage:
//
//	go run ./cmd/quotation <output.pdf>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/quotation <output.pdf>")
		fmt.Println("Creates a quotation-style PDF with headings, image, and tables, then saves to output.pdf.")
		os.Exit(1)
	}
	path := os.Args[1]

	builder := doc.New().
		Title("gPDF Quotation Example").
		Author("gPDF").
		Subject("Quotation with image and table-style layout").
		PageSize(595, 842).
		AddPage()

	// First page: heading and introductory text (using simple text runs).
	builder = builder.
		DrawText("Quotation", 72, 780, "Helvetica-Bold", 20).
		DrawText("Prepared for: Example Customer Ltd.", 72, 750, "Helvetica", 12).
		DrawText("Prepared by: gPDF Solutions", 72, 730, "Helvetica", 12).
		DrawText("Date: 2026-03-16", 72, 710, "Helvetica", 12)

	// Simple "table-style" layout using positioned text.
	builder = builder.
		// Header row
		DrawText("Item", 72, 640, "Helvetica-Bold", 12).
		DrawText("Description", 140, 640, "Helvetica-Bold", 12).
		DrawText("Qty", 400, 640, "Helvetica-Bold", 12).
		DrawText("Unit / Total", 440, 640, "Helvetica-Bold", 12).

		// Row 1
		DrawText("Q-001", 72, 620, "Helvetica", 11).
		DrawText("Implementation support (up to 40h).", 140, 620, "Helvetica", 11).
		DrawText("1", 400, 620, "Helvetica", 11).
		DrawText("5,000.00 / 5,000.00", 440, 620, "Helvetica", 11).

		// Row 2
		DrawText("Q-002", 72, 600, "Helvetica", 11).
		DrawText("Premium support for 12 months.", 140, 600, "Helvetica", 11).
		DrawText("1", 400, 600, "Helvetica", 11).
		DrawText("2,400.00 / 2,400.00", 440, 600, "Helvetica", 11).

		// Row 3
		DrawText("Q-003", 72, 580, "Helvetica", 11).
		DrawText("On-site training workshop (1 day).", 140, 580, "Helvetica", 11).
		DrawText("1", 400, 580, "Helvetica", 11).
		DrawText("3,000.00 / 3,000.00", 440, 580, "Helvetica", 11).

		// Totals
		DrawText("Net: 10,400.00", 380, 550, "Helvetica-Bold", 11).
		DrawText("Tax (10%): 1,040.00", 380, 535, "Helvetica-Bold", 11).
		DrawText("Total: 11,440.00", 380, 520, "Helvetica-Bold", 11).

		// Footer note.
		DrawText("This quotation is valid for 30 days from the date above.", 72, 480, "Helvetica-Oblique", 10).
		DrawText("Please contact sales@example.com for any questions or adjustments.", 72, 465, "Helvetica-Oblique", 10)

	document, err := builder.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer document.Close()

	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := document.Save(f); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Saved quotation example: %s\n", path)
}

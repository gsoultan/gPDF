package main

import (
	"fmt"
	"os"

	"gpdf/doc"
)

// pdf20demo builds a PDF 2.0–oriented demo document showcasing:
// - Tagged structure (sections, headings, paragraphs, figures, tables, lists)
// - Optional content groups (layers)
// - AcroForm fields (text field, checkbox, submit button)
// - AES-256 encrypted save
//
// Usage:
//
//	go run ./cmd/pdf20demo <output.pdf>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/pdf20demo <output.pdf>")
		fmt.Println("Creates a PDF 2.0 demo with tagged content, layers, forms, and AES-256 encryption, then saves to output.pdf.")
		os.Exit(1)
	}
	path := os.Args[1]

	builder := doc.New().
		Title("gPDF PDF 2.0 Demo").
		Author("gPDF").
		Subject("Tagged PDF, layers, forms, and AES-256 encryption").
		PageSize(595, 842).
		SetLanguage("en-US").
		SetTagged().
		SetAcroFormSigFlags(3)
	builder.AddPage()

	// Section 1: tagged heading and paragraph.
	builder = builder.BeginSection().
		DrawHeading(0, 1, "PDF 2.0 Demo", 72, 780, "", 0).
		DrawTaggedParagraphBox(0,
			"This document demonstrates tagged content, optional layers, AcroForm fields, and AES-256 encryption produced by gPDF.",
			72, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 451},
		).
		EndSection()

	// Section 2: tagged table with header row.
	tb := builder.BeginTable(0, 72, 620, 451, 140, 3).
		AllowPageBreak().
		HeaderRow(
			doc.TableCellSpec{Text: "Item", IsHeader: true},
			doc.TableCellSpec{Text: "Description", IsHeader: true},
			doc.TableCellSpec{Text: "Total", IsHeader: true},
		).
		Row(
			doc.TableCellSpec{Text: "T-001"},
			doc.TableCellSpec{Text: "Tagged content (headings, paragraphs, lists, tables)."},
			doc.TableCellSpec{Text: "$1,000.00"},
		).
		Row(
			doc.TableCellSpec{Text: "T-002"},
			doc.TableCellSpec{Text: "Optional content groups (layers) for overlays and conditional content."},
			doc.TableCellSpec{Text: "$500.00"},
		).
		Row(
			doc.TableCellSpec{Text: "T-003"},
			doc.TableCellSpec{Text: "Interactive forms (text fields, checkboxes, submit buttons)."},
			doc.TableCellSpec{Text: "$750.00"},
		).
		EndTable()
	if tb == nil {
		fmt.Fprintln(os.Stderr, "failed to create table")
		os.Exit(1)
	}

	// Section 3: optional content group (layer) with overlay note.
	layer := builder.BeginLayer("Overlay", true)
	builder = builder.DrawInLayer(layer, 0, func(db *doc.DocumentBuilder) {
		db.DrawTaggedQuoteBox(0,
			"This note belongs to the 'Overlay' layer. In viewers that support optional content groups, it can be toggled.",
			72, 460, "Helvetica-Oblique", 10,
			doc.TextLayoutOptions{Width: 451},
		)
	})

	// Section 4: simple tagged list.
	builder = builder.BeginSection().
		DrawHeading(0, 2, "Features", 72, 420, "", 0)
	builder = builder.DrawList(0,
		[]string{
			"Tagged headings, paragraphs, tables, and lists for accessibility.",
			"Optional content layers for conditional display.",
			"AcroForm fields for interactive data entry.",
			"AES-256 encryption for modern password protection.",
		},
		72, 400, 14, true, "Helvetica", 11,
	).EndSection()

	// Section 5: AcroForm fields on the first page.
	builder = builder.
		AddTextField(0, 72, 340, 320, 360, "name", "", "Your name", true).
		AddTextField(0, 72, 310, 320, 330, "email", "", "Your email address", true).
		AddCheckBox(0, 72, 280, 84, 292, "accept_terms", false, "I accept the terms", true).
		DrawText("Name:", 72, 365, "Helvetica", 11).
		DrawText("Email:", 72, 335, "Helvetica", 11).
		DrawText("I accept the terms", 90, 285, "Helvetica", 11).
		AddSubmitButton(0, 380, 280, 520, 300, "submit", "Submit demo", "https://example.com/submit", "Submit the demo form")

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
	// Encrypt with AES-256 using a demo password.
	if err := document.SaveWithAES256Password(f, "demo-user", "demo-owner"); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Saved PDF 2.0 demo: %s\n", path)
}

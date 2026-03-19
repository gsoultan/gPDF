package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gpdf/doc"
)

// This example builds a more complex "quotation" style PDF document
// demonstrating headings, professional table layout with alternating colors,
// an embedded logo, and various text styling features.
//
// Usage:
//
//	go run ./cmd/quotation <output.pdf>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/quotation <output.pdf>")
		fmt.Println("Creates a professional quotation PDF and saves to output.pdf.")
		os.Exit(1)
	}
	path := os.Args[1]

	// ── Palette ──────────────────────────────────────────────────────────────
	brandColor := doc.Color{R: 0.1, G: 0.3, B: 0.6} // Deep Blue
	headerGray := doc.Color{R: 0.85, G: 0.85, B: 0.85}
	altRowGray := doc.Color{R: 0.94, G: 0.96, B: 1.0} // Light Blue-ish Gray
	lightText := doc.Color{R: 0.4, G: 0.4, B: 0.4}

	builder := doc.New().
		Title("Professional Quotation - gPDF Solutions").
		Author("gPDF Engineering").
		Subject("Business Quotation Example").
		PageSize(595, 842)
	builder.AddPage()

	// 1. Header with Logo and Title
	// Create a simple logo in-memory (raw RGB)
	logoData := createLogo()
	builder.DrawImage(480, 750, 60, 60, logoData, 100, 100, 8, "DeviceRGB")

	builder.
		DrawHeading(0, 1, "QUOTATION", 72, 780, "Helvetica-Bold", 24).
		DrawLine(0, 72, 770, 523, 770, doc.LineStyle{Width: 2, Color: brandColor})

	// 2. Contact Information
	builder.
		DrawTextColored("Prepared for:", 72, 740, "Helvetica-Bold", 10, lightText).
		DrawTextBox(0, "Example Customer Ltd.\n123 Business Avenue\nCity, State 54321", 72, 725, "Helvetica", 11, doc.TextLayoutOptions{Width: 200, LineHeight: 14}).
		DrawTextColored("Prepared by:", 300, 740, "Helvetica-Bold", 10, lightText).
		DrawTextBox(0, "gPDF Solutions Corp.\n456 Tech Park\nInnovation Way", 300, 725, "Helvetica", 11, doc.TextLayoutOptions{Width: 200, LineHeight: 14})

	builder.
		DrawTextColored("Date:", 72, 660, "Helvetica-Bold", 10, lightText).
		DrawText("2026-03-19", 110, 660, "Helvetica", 10).
		DrawTextColored("Quote #:", 300, 660, "Helvetica-Bold", 10, lightText).
		DrawText("QT-2026-001", 350, 660, "Helvetica", 10)

	// 3. Items Table
	table := builder.BeginTable(0, 72, 620, 451, 0, 4). // cols=4
								WithHeaderFillColor(headerGray).
								WithAlternateRowColor(altRowGray).
								HeaderRow(
			doc.TableCellSpec{Text: "Item ID", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Description"},
			doc.TableCellSpec{Text: "Qty", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "Unit Price", Style: doc.CellStyle{PaddingRight: 8}},
		).
		Row(
			doc.TableCellSpec{Text: "PROD-001", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Enterprise License - Core Module with multi-region high availability support and 24/7 priority response"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$5,000.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		Row(
			doc.TableCellSpec{Text: "SERV-042", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Implementation & Setup (On-site)"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$2,500.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		Row(
			doc.TableCellSpec{Text: "SUPP-ANN", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Annual Platinum Support Package"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$1,200.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		Row(
			doc.TableCellSpec{Text: "TRAIN-01", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "User Training Workshop (Up to 10 pax)"},
			doc.TableCellSpec{Text: "2", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$750.00", Style: doc.CellStyle{PaddingRight: 8}},
		)

	currentY := table.CurrentY() - 30 // add some margin after table
	builder = table.EndTable()

	// 4. Totals Area
	totalX := 400.0
	builder.
		DrawTextRight("Subtotal:", totalX, currentY, "Helvetica", 11).
		DrawTextRight("$10,200.00", 523, currentY, "Helvetica", 11)

	currentY -= 20
	builder.
		DrawTextRight("Tax (10%):", totalX, currentY, "Helvetica", 11).
		DrawTextRight("$1,020.00", 523, currentY, "Helvetica", 11)

	currentY -= 25
	builder.
		DrawLine(0, totalX-50, currentY+15, 523, currentY+15, doc.LineStyle{Width: 1, Color: brandColor}).
		DrawTextRightColored("Grand Total:", totalX, currentY, "Helvetica-Bold", 14, brandColor).
		DrawTextRightColored("$11,220.00", 523, currentY, "Helvetica-Bold", 14, brandColor)

	// 5. Terms and Conditions
	currentY -= 60
	builder.DrawHeading(0, 3, "Terms and Conditions", 72, currentY, "Helvetica-Bold", 12)
	currentY -= 20
	builder.DrawList(0, []string{
		"Quotation is valid for 30 days from the date of issue.",
		"Payment terms: 50% upfront, 50% upon completion.",
		"Implementation schedule to be agreed upon signing.",
		"All prices are in USD.",
	}, 72, currentY, 15, false, "Helvetica", 10)

	// 6. Footer
	builder.
		DrawLine(0, 72, 60, 523, 60, doc.LineStyle{Width: 0.5, Color: lightText}).
		DrawTextCenteredColored("Thank you for your business!", 297, 45, "Helvetica-Oblique", 10, lightText).
		DrawTextCenteredColored("gPDF Solutions Corp | www.gpdf-example.com | +1 (555) 012-3456", 297, 30, "Helvetica", 8, lightText)

	// Build and Save
	document, err := builder.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error building document:", err)
		os.Exit(1)
	}
	defer document.Close()

	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating file:", err)
		os.Exit(1)
	}
	if err := document.Save(f); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, "Error saving document:", err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "Error closing file:", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated complex quotation: %s\n", path)
}

func createLogo() []byte {
	const size = 100
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill background with a nice color
	brand := color.RGBA{R: 25, G: 77, B: 153, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, brand)
		}
	}

	// Draw a simple "G" shape in white
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 20; y < 80; y++ {
		for x := 20; x < 30; x++ {
			img.Set(x, y, white)
		} // Left bar
	}
	for x := 20; x < 80; x++ {
		for y := 20; y < 30; y++ {
			img.Set(x, y, white)
		} // Top bar
		for y := 70; y < 80; y++ {
			img.Set(x, y, white)
		} // Bottom bar
	}
	for y := 50; y < 80; y++ {
		for x := 70; x < 80; x++ {
			img.Set(x, y, white)
		} // Right bottom bar
	}
	for x := 50; x < 80; x++ {
		for y := 50; y < 60; y++ {
			img.Set(x, y, white)
		} // Middle bar
	}

	// Just return the raw RGB bytes since DrawImage expects it for DeviceRGB
	res := make([]byte, size*size*3)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			c := img.At(x, y).(color.RGBA)
			idx := (y*size + x) * 3
			res[idx] = c.R
			res[idx+1] = c.G
			res[idx+2] = c.B
		}
	}
	return res
}

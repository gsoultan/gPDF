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
		A4()
	page := builder.AddPage().CurrentPage()

	// 1. Header with Logo and Title
	// Create a simple logo in-memory (raw RGB)
	logoData := createLogo()
	builder.DrawImage(480, 750, 60, 60, logoData, 100, 100, 8, "DeviceRGB")

	page.Heading("QUOTATION", 1).At(72, 780).Draw()
	page.Line(doc.Pt{X: 72, Y: 770}, doc.Pt{X: 523, Y: 770}).
		Style(doc.LineStyle{Width: 2, Color: brandColor}).
		Draw()

	// 2. Contact Information
	page.Text("Prepared for:").At(72, 740).Font("Helvetica-Bold").Size(10).Color(lightText).Draw()
	page.TextBox("Example Customer Ltd.\n123 Business Avenue\nCity, State 54321").
		At(72, 725).
		Font("Helvetica").
		Size(11).
		Width(200).
		LineHeight(14).
		Draw()

	page.Text("Prepared by:").At(300, 740).Font("Helvetica-Bold").Size(10).Color(lightText).Draw()
	page.TextBox("gPDF Solutions Corp.\n456 Tech Park\nInnovation Way").
		At(300, 725).
		Font("Helvetica").
		Size(11).
		Width(200).
		LineHeight(14).
		Draw()

	page.Text("Date:").At(72, 660).Font("Helvetica-Bold").Size(10).Color(lightText).Draw()
	page.Text("2026-03-19").At(110, 660).Font("Helvetica").Size(10).Draw()

	page.Text("Quote #:").At(300, 660).Font("Helvetica-Bold").Size(10).Color(lightText).Draw()
	page.Text("QT-2026-001").At(350, 660).Font("Helvetica").Size(10).Draw()

	// 3. Items Table
	table := page.Table(4).
		At(72, 620).
		Width(451).
		WithHeaderFillColor(headerGray).
		WithAlternateRowColor(altRowGray).
		HeaderSpec(
			doc.TableCellSpec{Text: "Item ID", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Description"},
			doc.TableCellSpec{Text: "Qty", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "Unit Price", Style: doc.CellStyle{PaddingRight: 8}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "PROD-001", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Enterprise License - Core Module with multi-region high availability support and 24/7 priority response"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$5,000.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "SERV-042", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Implementation & Setup (On-site)"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$2,500.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "SUPP-ANN", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Annual Platinum Support Package"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$1,200.00", Style: doc.CellStyle{PaddingRight: 8}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "TRAIN-01", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "User Training Workshop (Up to 10 pax)"},
			doc.TableCellSpec{Text: "2", Style: doc.CellStyle{PaddingRight: 8}},
			doc.TableCellSpec{Text: "$750.00", Style: doc.CellStyle{PaddingRight: 8}},
		)

	currentY := table.CurrentY() - 30 // add some margin after table
	builder = table.EndTable()

	// 4. Totals Area
	totalX := 400.0
	page.Text("Subtotal:").At(totalX, currentY).Font("Helvetica").Size(11).Align(doc.AlignRight).Draw()
	page.Text("$10,200.00").At(523, currentY).Font("Helvetica").Size(11).Align(doc.AlignRight).Draw()

	currentY -= 20
	page.Text("Tax (10%):").At(totalX, currentY).Font("Helvetica").Size(11).Align(doc.AlignRight).Draw()
	page.Text("$1,020.00").At(523, currentY).Font("Helvetica").Size(11).Align(doc.AlignRight).Draw()

	currentY -= 25
	page.Line(doc.Pt{X: totalX - 50, Y: currentY + 15}, doc.Pt{X: 523, Y: currentY + 15}).
		Style(doc.LineStyle{Width: 1, Color: brandColor}).
		Draw()

	page.Text("Grand Total:").At(totalX, currentY).Font("Helvetica-Bold").Size(14).Color(brandColor).Align(doc.AlignRight).Draw()
	page.Text("$11,220.00").At(523, currentY).Font("Helvetica-Bold").Size(14).Color(brandColor).Align(doc.AlignRight).Draw()

	// 5. Terms and Conditions
	currentY -= 60
	page.Heading("Terms and Conditions", 3).At(72, currentY).Draw()
	currentY -= 20
	page.List([]string{
		"Quotation is valid for 30 days from the date of issue.",
		"Payment terms: 50% upfront, 50% upon completion.",
		"Implementation schedule to be agreed upon signing.",
		"All prices are in USD.",
	}).At(72, currentY).LineHeight(15).Font("Helvetica").Size(10).Draw()

	// 6. Footer
	page.Line(doc.Pt{X: 72, Y: 60}, doc.Pt{X: 523, Y: 60}).
		Style(doc.LineStyle{Width: 0.5, Color: lightText}).
		Draw()

	page.Text("Thank you for your business!").At(297, 45).Font("Helvetica-Oblique").Size(10).Color(lightText).Align(doc.AlignCenter).Draw()
	page.Text("gPDF Solutions Corp | www.gpdf-example.com | +1 (555) 012-3456").At(297, 30).Font("Helvetica").Size(8).Color(lightText).Align(doc.AlignCenter).Draw()

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

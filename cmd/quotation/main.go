package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/gsoultan/gpdf/doc"
)

// This example builds a more complex "quotation" style PDF document
// demonstrating a "full fluent" based flow layout API.
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

	logoData := createLogo()

	err := doc.New().
		Title("Professional Quotation - gPDF Solutions").
		Author("gPDF Engineering").
		Subject("Business Quotation Example").
		A4().
		AddPage().
		Flow(doc.FlowOptions{Margin: 72}).
		// 1. Header with Logo (Using a table for side-by-side layout in flow)
		Table(2).
		WithColumnWidths(300, 151). // Adjusted for A4 (595 - 72*2 = 451)
		RowSpec(
			doc.TableCellSpec{Text: "QUOTATION", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 24, TextColor: brandColor, HasTextColor: true}},
			doc.TableCellSpec{Image: &doc.TableCellImageSpec{Raw: logoData, WidthPt: 60, HeightPt: 60}, Style: doc.CellStyle{HAlign: doc.CellHAlignRight}},
		).
		Done().
		Line(2, brandColor).
		Space(20).
		// 2. Contact Information & Meta Info
		Table(2).
		RowSpec(
			doc.TableCellSpec{Text: "Prepared for:", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 10, TextColor: lightText, HasTextColor: true}},
			doc.TableCellSpec{Text: "Prepared by:", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 10, TextColor: lightText, HasTextColor: true, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "Example Customer Ltd.\n123 Business Avenue\nCity, State 54321", Style: doc.CellStyle{FontSize: 11}},
			doc.TableCellSpec{Text: "gPDF Solutions Corp.\n456 Tech Park\nInnovation Way", Style: doc.CellStyle{FontSize: 11, HAlign: doc.CellHAlignRight}},
		).
		Done().
		Space(20).
		Table(2).
		RowSpec(
			doc.TableCellSpec{Text: "Date:", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 10, TextColor: lightText, HasTextColor: true}},
			doc.TableCellSpec{Text: "Quote #:", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 10, TextColor: lightText, HasTextColor: true, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "2026-03-19", Style: doc.CellStyle{FontSize: 10}},
			doc.TableCellSpec{Text: "QT-2026-001", Style: doc.CellStyle{FontSize: 10, HAlign: doc.CellHAlignRight}},
		).
		Done().
		Space(30).
		// 3. Items Table
		Table(4).
		WithColumnWidths(1, 4, 1, 1).
		WithHeaderFillColor(headerGray).
		WithAlternateRowColor(altRowGray).
		HeaderSpec(
			doc.TableCellSpec{Text: "Item ID", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Description"},
			doc.TableCellSpec{Text: "Qty", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "Unit Price", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "PROD-001", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Enterprise License - Core Module with multi-region high availability support and 24/7 priority response"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$5,000.00", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "SERV-042", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Implementation & Setup (On-site)"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$2,500.00", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "SUPP-ANN", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "Annual Platinum Support Package"},
			doc.TableCellSpec{Text: "1", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$1,200.00", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "TRAIN-01", Style: doc.CellStyle{PaddingLeft: 8}},
			doc.TableCellSpec{Text: "User Training Workshop (Up to 10 pax)"},
			doc.TableCellSpec{Text: "2", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$750.00", Style: doc.CellStyle{PaddingRight: 8, HAlign: doc.CellHAlignRight}},
		).
		Done().
		Space(30).
		// 4. Totals Area (Now fully part of the flow using a borderless table)
		Table(2).
		WithColumnWidths(350, 101).
		RowSpec(
			doc.TableCellSpec{Text: "Subtotal:", Style: doc.CellStyle{HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$10,200.00", Style: doc.CellStyle{HAlign: doc.CellHAlignRight}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "Tax (10%):", Style: doc.CellStyle{HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$1,020.00", Style: doc.CellStyle{HAlign: doc.CellHAlignRight}},
		).
		Done().
		Space(5).
		Line(1, brandColor).
		Space(5).
		Table(2).
		WithColumnWidths(350, 101).
		RowSpec(
			doc.TableCellSpec{Text: "Grand Total:", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 14, TextColor: brandColor, HasTextColor: true, HAlign: doc.CellHAlignRight}},
			doc.TableCellSpec{Text: "$11,220.00", Style: doc.CellStyle{FontName: "Helvetica-Bold", FontSize: 14, TextColor: brandColor, HasTextColor: true, HAlign: doc.CellHAlignRight}},
		).
		Done().
		Space(60).
		// 5. Terms and Conditions
		Heading("Terms and Conditions", 3).
		Space(10).
		List([]string{
			"Quotation is valid for 30 days from the date of issue.",
			"Payment terms: 50% upfront, 50% upon completion.",
			"Implementation schedule to be agreed upon signing.",
			"All prices are in USD.",
		}, false).
		// 6. Footer (Back to absolute positioning at the bottom)
		End().
		CurrentPage().
		Line(doc.Pt{X: 72, Y: 60}, doc.Pt{X: 523, Y: 60}).Width(0.5).Color(lightText).Draw().
		Text("Thank you for your business!").At(297, 45).Font("Helvetica-Oblique").Size(10).Color(lightText).Align(doc.AlignCenter).Draw().
		Text("gPDF Solutions Corp | www.gpdf-example.com | +1 (555) 012-3456").At(297, 30).Font("Helvetica").Size(8).Color(lightText).Align(doc.AlignCenter).Draw().
		End().
		BuildAndSave(path)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating quotation:", err)
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

	// Encode to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}

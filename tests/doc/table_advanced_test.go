package doc_test

import (
	"bytes"
	"testing"

	"gpdf/doc"
)

func TestTableAdvancedFeatures(t *testing.T) {
	b := doc.New().PageSize(595, 842)
	b.AddPage()

	// 1. Weighted column widths (1:2:1)
	// 2. repeating headers and footers
	// 3. custom cell styles (font, size)
	// 4. row splitting (one very tall row)

	tbl := b.BeginTable(0, 50, 750, 500, 0, 3).
		AllowPageBreak().
		WithMargins(50, 50).
		WithColumnWidths(1, 2, 1).
		WithHeaderFillColor(doc.ColorNavy).
		WithAlternateRowColor(doc.ColorLightGray)

	// Header
	tbl.HeaderSpec(
		doc.TableCellSpec{Text: "Key", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}, FontName: "Helvetica-Bold", FontSize: 12}},
		doc.TableCellSpec{Text: "Value / Long Description", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}, FontName: "Helvetica-Bold", FontSize: 12}},
		doc.TableCellSpec{Text: "Status", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}, FontName: "Helvetica-Bold", FontSize: 12}},
	)

	// Footer
	tbl.FooterRow(
		doc.TableCellSpec{Text: "Page Summary", ColSpan: 2, Style: doc.CellStyle{FontName: "Helvetica-Oblique", FontSize: 8}},
		doc.TableCellSpec{Text: "Continued...", Style: doc.CellStyle{FontName: "Helvetica-Oblique", FontSize: 8}},
	)

	// Some regular rows
	for i := 1; i <= 5; i++ {
		tbl.RowSpec(
			doc.TableCellSpec{Text: "Item"},
			doc.TableCellSpec{Text: "Normal description text for item."},
			doc.TableCellSpec{Text: "OK"},
		)
	}

	// A very tall row that MUST split
	longText := "This is a very long text that should span multiple pages. "
	for range 20 {
		longText += "We are testing the automatic row splitting feature of the gPDF library. "
	}

	tbl.RowSpec(
		doc.TableCellSpec{Text: "Long Item"},
		doc.TableCellSpec{Text: longText, Style: doc.CellStyle{FontSize: 11}},
		doc.TableCellSpec{Text: "SPLIT"},
	)

	// More rows after the split
	for i := 1; i <= 5; i++ {
		tbl.RowSpec(
			doc.TableCellSpec{Text: "Post-Split"},
			doc.TableCellSpec{Text: "Text after the split."},
			doc.TableCellSpec{Text: "DONE"},
		)
	}

	tbl.EndTable()

	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("Empty PDF generated")
	}

	// Check if we have multiple pages (since it should have split)
	pages, err := d.Pages()
	if err != nil {
		t.Fatalf("Pages failed: %v", err)
	}
	if len(pages) < 2 {
		t.Errorf("Expected at least 2 pages due to splitting, got %d", len(pages))
	}
}

func TestTableWeightedWidths(t *testing.T) {
	b := doc.New().PageSize(500, 500)
	b.AddPage()

	// Test 1:3:1 weights
	tbl := b.BeginTable(0, 50, 450, 400, 0, 3).
		WithColumnWidths(1, 3, 1)

	tbl.RowSpec(
		doc.TableCellSpec{Text: "W1"},
		doc.TableCellSpec{Text: "W3 (Wide)"},
		doc.TableCellSpec{Text: "W1"},
	).EndTable()

	_, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
}

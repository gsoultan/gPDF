package doc_test

import (
	"gpdf/doc"
	"testing"
)

func TestNewImprovedAPI(t *testing.T) {
	b := doc.New().
		Title("New API Test").
		PageSize(595, 842)

	// 1. Page-based API
	p := b.AddPage().CurrentPage()
	p.Heading("Welcome to gPDF New API", 1).At(72, 750).Draw()
	p.Text("This is property chaining").
		At(72, 720).
		Font("Helvetica-Oblique").
		Size(14).
		Color(doc.ColorNavy).
		Draw()

	// 2. Simplified Geometry
	p.Line(doc.At(72, 715), doc.At(523, 715)).
		Width(2).
		Color(doc.ColorRed).
		Draw()

	p.Rect(doc.R(72, 600, 100, 50)).
		Fill(doc.ColorLightGray).
		Stroke(doc.LineStyle{Width: 1, Color: doc.ColorBlack}).
		Draw()

	// 3. Declarative Table
	p.Table(3).
		At(72, 500).
		Width(400).
		Header("ID", "Item", "Price").
		Row("1", "Widget", "$10.00").
		Row("2", "Gadget", "$20.00").
		Draw()

	// 4. Flow API
	b.Flow(doc.FlowOptions{Margin: 72}).
		Heading("Automatic Flow", 2).
		Paragraph("This text automatically wraps and can span multiple lines. We don't need to track Y coordinates manually here.").
		Space(20).
		Heading("Another Section", 3).
		Paragraph("More flowing content below the first paragraph.").
		End()

	_, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
}

package doc_test

import (
	"gpdf/doc"
	"testing"
)

func TestFlowExtended(t *testing.T) {
	b := doc.New().Title("Extended Flow Test").A4()

	// 1. Test Unified Cursor on Page
	p := b.AddPage().CurrentPage()
	p.Heading("Unified Cursor Test", 1).Draw() // Should be at top
	p.Paragraph("This is the first paragraph using unified cursor.").Draw()
	p.Paragraph("This is the second paragraph, it should be below the first one.").Draw()

	y1 := p.CurrentY()
	p.Text("Simple text line").Draw()
	y2 := p.CurrentY()
	if y2 >= y1 {
		t.Errorf("Expected Y to decrease after Text.Draw, got %v -> %v", y1, y2)
	}

	// 2. Test Flow with new elements
	logoData := make([]byte, 100*100*3) // dummy RGB data
	b.Flow(doc.FlowOptions{Margin: 72}).
		Heading("Flow with New Elements", 2).
		Paragraph("Testing image, list, line, and rect in flow.").
		Image(logoData, 50, 50).
		Space(10).
		List([]string{"Item 1", "Item 2", "Item 3"}, true).
		Line(2, doc.ColorNavy).
		Rect(30, doc.LineStyle{Width: 1, Color: doc.ColorBlack}, doc.ColorLightGray, true).
		End()

	// 3. Test Table-to-Flow integration
	flow := b.Flow(doc.FlowOptions{Margin: 72})
	flow.Heading("Table Integration", 2).
		Paragraph("The following table is part of the flow.")

	flow.Table(2).
		Header("Key", "Value").
		Row("Flow", "Working").
		Row("Integration", "Seamless").
		EndTable()

	flow.Paragraph("This paragraph should be below the table.")
	flow.End()

	_, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
}

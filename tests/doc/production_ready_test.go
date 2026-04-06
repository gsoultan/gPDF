package doc_test

import (
	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/doc/style"
	"testing"
)

func TestProductionReadyFeatures(t *testing.T) {
	b := doc.New()
	b.AddPage()
	p := b.CurrentPage()

	// 1. Advanced List Styling
	p.List([]string{"Item 1", "Item 2", "Item 3"}).
		Marker(style.ListMarkerSquare).
		Color(doc.ColorBlue).
		At(72, 700).
		Draw()

	p.List([]string{"Step I", "Step II", "Step III"}).
		Marker(style.ListMarkerRomanUpper).
		At(72, 600).
		Draw()

	// 2. Full Justification
	p.TextBox("This is a long paragraph that should be fully justified. It should spread across the entire width of the box, ensuring that both the left and right margins are straight.").
		At(72, 500).
		Width(200).
		Align(doc.AlignJustify).
		Draw()

	// 3. Synthetic Styling (Simulated via API if available, or just verify it compiles)
	// Currently we don't have a fluent API for synthetic bold on TextObject,
	// but it's used internally by some components or can be added.
	// We'll just verify the builder methods for now.

	// 4. Bidi Support (Basic)
	p.Text("Arabic text: \u0627\u0644\u0633\u0644\u0627\u0645 \u0639\u0644\u064a\u0643\u0645").
		At(72, 400).
		Draw()

	// 5. Vertical Text
	p.TextBox("VERTICAL").
		At(400, 700).
		Width(20).
		Style(doc.TextStyle{IsVertical: true}).
		Draw()

	if b.Err() != nil {
		t.Fatalf("Document build failed: %v", b.Err())
	}
}

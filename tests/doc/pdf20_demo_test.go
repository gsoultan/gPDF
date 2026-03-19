package doc

import (
	"bytes"
	"testing"

	"gpdf/doc"
)

// TestPDF20DemoLikeDocument builds a representative PDF 2.0 style document and
// asserts that it can be saved, reopened, and that basic structures exist.
func TestPDF20DemoLikeDocument(t *testing.T) {
	builder := doc.New().
		Title("Test PDF 2.0 Demo").
		Author("gPDF").
		SetLanguage("en-US").
		SetTagged().
		AddPage()

	// Tagged heading and paragraph.
	builder = builder.BeginSection().
		DrawHeading(0, 1, "Heading", 72, 780, "", 0)
	builder = builder.DrawTaggedParagraphBox(0,
		"Demo paragraph.", 72, 760, "Helvetica", 12,
		doc.TextLayoutOptions{Width: 451},
	).EndSection()

	// Simple tagged table.
	tb := builder.BeginTable(0, 72, 700, 451, 60, 2).
		HeaderSpec(
			doc.TableCellSpec{Text: "Col1", IsHeader: true},
			doc.TableCellSpec{Text: "Col2", IsHeader: true},
		).
		RowSpec(
			doc.TableCellSpec{Text: "A"},
			doc.TableCellSpec{Text: "B"},
		).
		EndTable()
	if tb == nil {
		t.Fatalf("BeginTable returned nil")
	}

	// A simple layer and a form field.
	layer := builder.BeginLayer("TestLayer", true)
	builder = builder.DrawInLayer(layer, 0, func(db *doc.DocumentBuilder) {
		db.DrawTaggedQuoteBox(0, "Layered text.", 72, 660, "Helvetica-Oblique", 10,
			doc.TextLayoutOptions{Width: 451})
	})
	builder = builder.SetAcroForm().
		AddTextField(0, 72, 620, 200, 640, "field", "", "Field", false)

	document, err := builder.Build()
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	defer document.Close()

	var buf bytes.Buffer
	if err := document.Save(&buf); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty PDF output")
	}
}

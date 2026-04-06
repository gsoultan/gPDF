package doc_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

func TestDrawTextBox_WrapsIntoMultipleLines(t *testing.T) {
	buf := new(bytes.Buffer)
	builder := doc.New().
		NoCompressContent().
		Title("Wrap").
		PageSize(200, 200).
		AddPage()

	text := "This is a long line that should wrap into multiple lines within the text box."
	opts := doc.TextLayoutOptions{
		Width:      80, // narrow box to force wrapping
		LineHeight: 12,
	}
	d, err := builder.DrawTextBox(0, text, 20, 150, "Helvetica", 10, opts).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer d.Close()

	if err := d.Save(buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	pdf := buf.String()
	// Expect multiple Tj operators from wrapped lines.
	if strings.Count(pdf, " Tj") < 2 {
		t.Fatalf("expected at least 2 Tj operators for wrapped lines, got %d", strings.Count(pdf, " Tj"))
	}
}

func TestDrawTextBox_AlignmentAffectsXCoordinate(t *testing.T) {
	text := "Align me"
	opts := doc.TextLayoutOptions{
		Width:      120,
		LineHeight: 12,
	}

	leftBuf := new(bytes.Buffer)
	leftDoc, err := doc.New().
		NoCompressContent().
		Title("Left").
		PageSize(200, 200).
		AddPage().
		DrawTextBox(0, text, 40, 150, "Helvetica", 10, opts).
		Build()
	if err != nil {
		t.Fatalf("left Build: %v", err)
	}
	defer leftDoc.Close()
	if err := leftDoc.Save(leftBuf); err != nil {
		t.Fatalf("left Save: %v", err)
	}

	centerBuf := new(bytes.Buffer)
	opts.Align = doc.TextAlignCenter
	centerDoc, err := doc.New().
		NoCompressContent().
		Title("Center").
		PageSize(200, 200).
		AddPage().
		DrawTextBox(0, text, 40, 150, "Helvetica", 10, opts).
		Build()
	if err != nil {
		t.Fatalf("center Build: %v", err)
	}
	defer centerDoc.Close()
	if err := centerDoc.Save(centerBuf); err != nil {
		t.Fatalf("center Save: %v", err)
	}

	leftPDF := leftBuf.String()
	centerPDF := centerBuf.String()

	// Look for first Td occurrence as a simple proxy for X coordinate positioning.
	leftIdx := strings.Index(leftPDF, " Td")
	centerIdx := strings.Index(centerPDF, " Td")
	if leftIdx == -1 || centerIdx == -1 {
		t.Fatalf("expected Td operators in both PDFs (leftIdx=%d, centerIdx=%d)", leftIdx, centerIdx)
	}
	if leftIdx == centerIdx {
		t.Fatalf("expected different Td positioning between left and center aligned text")
	}
}

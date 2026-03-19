package doc_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"gpdf/doc"
)

// ── Color ────────────────────────────────────────────────────────────────────

func TestColorFromHex(t *testing.T) {
	tests := []struct {
		input   string
		want    doc.Color
		wantErr bool
	}{
		{"#FF0000", doc.Color{R: 1, G: 0, B: 0}, false},
		{"FF0000", doc.Color{R: 1, G: 0, B: 0}, false},
		{"#000000", doc.Color{R: 0, G: 0, B: 0}, false},
		{"#ffffff", doc.Color{R: 1, G: 1, B: 1}, false},
		{"#80FF80", doc.Color{R: float64(0x80) / 255, G: 1, B: float64(0x80) / 255}, false},
		{"F00", doc.Color{R: 1, G: 0, B: 0}, false},
		{"#F00", doc.Color{R: 1, G: 0, B: 0}, false},
		{"invalid", doc.Color{}, true}, // bad input → error
	}
	for _, tc := range tests {
		got, err := doc.ColorFromHex(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ColorFromHex(%q) expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ColorFromHex(%q) error: %v", tc.input, err)
			continue
		}
		if abs64(got.R-tc.want.R) > 0.005 || abs64(got.G-tc.want.G) > 0.005 || abs64(got.B-tc.want.B) > 0.005 {
			t.Errorf("ColorFromHex(%q) = {%.3f %.3f %.3f}, want {%.3f %.3f %.3f}",
				tc.input, got.R, got.G, got.B, tc.want.R, tc.want.G, tc.want.B)
		}
	}
}

func TestColorGray50(t *testing.T) {
	c := doc.ColorGray50(0.5)
	if c.R != 0.5 || c.G != 0.5 || c.B != 0.5 {
		t.Errorf("ColorGray50(0.5) = %v, want {0.5 0.5 0.5}", c)
	}
}

func TestPredefinedColors(t *testing.T) {
	// Spot-check a few predefined colors are distinct and non-zero.
	if doc.ColorOrange == doc.ColorBlue {
		t.Error("ColorOrange should differ from ColorBlue")
	}
	if doc.ColorTeal == (doc.Color{}) {
		t.Error("ColorTeal should be non-zero")
	}
}

func abs64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// ── Drawing ──────────────────────────────────────────────────────────────────

func buildOnePagePDF(t *testing.T) *doc.DocumentBuilder {
	t.Helper()
	b := doc.New().PageSize(595, 842)
	b.AddPage()
	return b
}

func mustBuild(t *testing.T, b *doc.DocumentBuilder) {
	t.Helper()
	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	buf := new(bytes.Buffer)
	if err := d.Save(buf); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF")
	}
	_ = d.Close()
}

func TestDrawRoundedRect(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawRoundedRect(0, 50, 50, 200, 100, 10, doc.LineStyle{Width: 1, Color: doc.ColorBlack}).
		FillRoundedRect(0, 50, 200, 200, 100, 15, doc.ColorTeal).
		FillStrokeRoundedRect(0, 50, 350, 200, 100, 8, doc.ColorOrange, doc.LineStyle{Width: 2, Color: doc.ColorBlack})
	mustBuild(t, b)
}

func TestDrawEllipse(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawEllipse(0, 150, 600, 80, 40, doc.LineStyle{Width: 1, Color: doc.ColorBlue}).
		FillEllipse(0, 150, 500, 60, 30, doc.ColorPurple).
		FillStrokeEllipse(0, 150, 400, 70, 35, doc.ColorYellow, doc.LineStyle{Width: 1, Color: doc.ColorBlack})
	mustBuild(t, b)
}

func TestDrawPolygon(t *testing.T) {
	// Triangle.
	triangle := []float64{100, 700, 200, 700, 150, 650}
	b := buildOnePagePDF(t).
		DrawPolygon(0, triangle, doc.LineStyle{Width: 1, Color: doc.ColorRed}).
		FillPolygon(0, triangle, doc.ColorGreen).
		FillStrokePolygon(0, triangle, doc.ColorCyan, doc.LineStyle{Width: 2, Color: doc.ColorNavy})
	mustBuild(t, b)
}

func TestPathBuilderArc(t *testing.T) {
	b := buildOnePagePDF(t)
	b = b.BeginPath(0).
		Arc(200, 400, 80, 50, 0, 270).
		ClosePath().
		Stroke(doc.LineStyle{Width: 1, Color: doc.ColorBlue})
	mustBuild(t, b)
}

func TestPathBuilderRoundedRect(t *testing.T) {
	b := buildOnePagePDF(t)
	b = b.BeginPath(0).
		RoundedRect(100, 300, 200, 80, 12).
		Fill(doc.ColorLightGray)
	mustBuild(t, b)
}

// ── Image opacity and rotation ────────────────────────────────────────────────

func makeTestJPEG(t *testing.T) ([]byte, int, int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := range 10 {
		for x := range 10 {
			img.Set(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes(), 10, 10
}

func TestDrawJPEGWithOpacity(t *testing.T) {
	data, w, h := makeTestJPEG(t)
	b := buildOnePagePDF(t).
		DrawJPEGWithOpacity(50, 600, 100, 100, data, w, h, "DeviceRGB", 0.5)
	mustBuild(t, b)
}

func TestDrawJPEGRotated(t *testing.T) {
	data, w, h := makeTestJPEG(t)
	b := buildOnePagePDF(t).
		DrawJPEGRotated(50, 600, 100, 100, data, w, h, "DeviceRGB", 45)
	mustBuild(t, b)
}

// ── Text ────────────────────────────────────────────────────────────────────

func TestDrawTextCentered(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawTextCentered("Hello Center", 297, 700, "Helvetica", 14).
		DrawTextCenteredColored("Colored Center", 297, 680, "Helvetica", 12, doc.ColorRed, 0)
	mustBuild(t, b)
}

func TestDrawTextRight(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawTextRight("Right aligned", 580, 700, "Helvetica", 12).
		DrawTextRightColored("Right colored", 580, 680, "Helvetica-Bold", 12, doc.ColorBlue, 0)
	mustBuild(t, b)
}

func TestDrawTextWithUnderline(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawTextWithUnderline("Underlined text", 50, 700, "Helvetica", 14, doc.ColorBlack)
	mustBuild(t, b)
}

func TestDrawTextWithStrikethrough(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawTextWithStrikethrough("Strike me", 50, 700, "Helvetica", 14, doc.ColorRed)
	mustBuild(t, b)
}

func TestTextLayoutOptionsLetterSpacing(t *testing.T) {
	b := buildOnePagePDF(t).
		DrawTextBoxColored(0, "Wide spaced text here.", 50, 750,
			"Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 18, LetterSpacing: 1.5},
			doc.ColorBlack)
	mustBuild(t, b)
}

func TestDrawTextBoxAlignCenter(t *testing.T) {
	const para = "The quick brown fox jumps over the lazy dog near the river bank."
	b := buildOnePagePDF(t).
		DrawTextBox(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignCenter})
	mustBuild(t, b)
}

func TestDrawTextBoxAlignRight(t *testing.T) {
	const para = "The quick brown fox jumps over the lazy dog near the river bank."
	b := buildOnePagePDF(t).
		DrawTextBox(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignRight})
	mustBuild(t, b)
}

func TestDrawTextBoxAlignJustify(t *testing.T) {
	const para = "The quick brown fox jumps over the lazy dog. A second sentence follows to ensure multiple wrapped lines are produced for justify testing."
	b := buildOnePagePDF(t).
		DrawTextBox(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignJustify})
	mustBuild(t, b)
}

func TestDrawTextBoxColoredAlignCenter(t *testing.T) {
	const para = "Centered colored text spanning multiple lines for layout verification."
	b := buildOnePagePDF(t).
		DrawTextBoxColored(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignCenter},
			doc.ColorBlue)
	mustBuild(t, b)
}

func TestDrawTextBoxColoredAlignRight(t *testing.T) {
	const para = "Right-aligned colored text spanning multiple lines for layout verification."
	b := buildOnePagePDF(t).
		DrawTextBoxColored(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignRight},
			doc.ColorRed)
	mustBuild(t, b)
}

func TestDrawTextBoxColoredAlignJustify(t *testing.T) {
	const para = "Justified colored text. The quick brown fox jumps over the lazy dog near the river bank on a sunny afternoon."
	b := buildOnePagePDF(t).
		DrawTextBoxColored(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignJustify},
			doc.ColorDarkGray)
	mustBuild(t, b)
}

func TestDrawTextBoxJustifyMultiParagraph(t *testing.T) {
	// Two paragraphs separated by \n — last line of each paragraph must not be stretched.
	const para = "First paragraph with enough words to wrap.\nSecond paragraph also has enough words to wrap across lines here."
	b := buildOnePagePDF(t).
		DrawTextBox(0, para, 50, 750, "Helvetica", 12,
			doc.TextLayoutOptions{Width: 300, LineHeight: 16, Align: doc.TextAlignJustify})
	mustBuild(t, b)
}

// ── Table ───────────────────────────────────────────────────────────────────

func TestTableWithFillColors(t *testing.T) {
	b := buildOnePagePDF(t)
	b = b.BeginTable(0, 50, 750, 450, 200, 3).
		WithHeaderFillColor(doc.ColorNavy).
		WithAlternateRowColor(doc.ColorLightGray).
		HeaderSpec(
			doc.TableCellSpec{Text: "Name", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}}},
			doc.TableCellSpec{Text: "Age", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}}},
			doc.TableCellSpec{Text: "City", Style: doc.CellStyle{TextColorRGB: [3]float64{1, 1, 1}}},
		).
		RowSpec(
			doc.TableCellSpec{Text: "Alice"},
			doc.TableCellSpec{Text: "30"},
			doc.TableCellSpec{Text: "Jakarta"},
		).
		RowSpec(
			doc.TableCellSpec{Text: "Bob"},
			doc.TableCellSpec{Text: "25"},
			doc.TableCellSpec{Text: "Bandung"},
		).
		RowSpec(
			doc.TableCellSpec{
				Text:    "Merged col",
				Style:   doc.CellStyle{FillColor: doc.ColorOrange, HasFillColor: true},
				ColSpan: 2,
			},
			doc.TableCellSpec{Text: "Bali"},
		).
		EndTable()
	mustBuild(t, b)
}

func TestTableColSpanWidth(t *testing.T) {
	// Verify that a ColSpan:3 cell across a 3-col table builds without error.
	b := buildOnePagePDF(t)
	b = b.BeginTable(0, 50, 750, 450, 100, 3).
		RowSpec(
			doc.TableCellSpec{Text: "Full width span", ColSpan: 3},
		).
		RowSpec(
			doc.TableCellSpec{Text: "A"},
			doc.TableCellSpec{Text: "B"},
			doc.TableCellSpec{Text: "C"},
		).
		EndTable()
	mustBuild(t, b)
}

package doc_test

import (
	"strings"
	"testing"

	"gpdf/doc"
	"gpdf/font"
)

type mockFont struct {
	name      string
	supported map[rune]bool
	kerns     map[uint32]int
}

func (f *mockFont) PostScriptName() string { return f.name }
func (f *mockFont) UnitsPerEm() int        { return 1000 }
func (f *mockFont) Ascent() int            { return 800 }
func (f *mockFont) Descent() int           { return -200 }
func (f *mockFont) LineGap() int           { return 0 }
func (f *mockFont) GlyphWidth(r rune) int {
	if f.supported[r] {
		return 500
	}
	return 0
}
func (f *mockFont) Contains(r rune) bool {
	return f.supported[r]
}
func (f *mockFont) Kern(r1, r2 rune) int {
	return f.kerns[uint32(r1)<<16|uint32(r2)]
}
func (f *mockFont) TextWidth(text string, fontSize float64) float64 {
	var total int
	var lastR rune
	for i, r := range text {
		total += f.GlyphWidth(r)
		if i > 0 {
			total += f.Kern(lastR, r)
		}
		lastR = r
	}
	return float64(total) * fontSize / 1000.0
}
func (f *mockFont) Metrics() font.Metrics { return font.Metrics{UnitsPerEm: 1000} }

func TestFontFallback(t *testing.T) {
	b := doc.New()

	mainFont := &mockFont{
		name:      "Main",
		supported: map[rune]bool{'a': true, 'b': true},
	}
	fallbackFont := &mockFont{
		name:      "Fallback",
		supported: map[rune]bool{'c': true},
	}

	b.RegisterFont(mainFont)
	b.RegisterFallbackFont(fallbackFont)

	// 'abc' -> 'ab' from Main, 'c' from Fallback
	b.AddPage()
	b.DrawText("abc", 50, 700, "Main", 12)

	// Verify warnings
	b.DrawText("d", 50, 680, "Main", 12) // 'd' is missing in both
	warnings := b.Warnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "no glyph for d") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for missing glyph 'd', got %v", warnings)
	}
}

func TestRichText(t *testing.T) {
	b := doc.New()
	b.AddPage()

	rt := doc.NewRichText().
		Add("Hello ", doc.DefaultTextStyle()).
		Add("Bold", doc.DefaultTextStyle().Font("Helvetica-Bold")).
		Add(" Red", doc.DefaultTextStyle().WithColor(doc.ColorRed))

	b.DrawRichText(0, rt, 50, 700)
	mustBuild(t, b)
}

func TestKerningCalculation(t *testing.T) {
	f := &mockFont{
		name:      "Kerns",
		supported: map[rune]bool{'A': true, 'V': true},
		kerns: map[uint32]int{
			uint32('A')<<16 | uint32('V'): -100,
		},
	}

	// Width of 'A' is 500, 'V' is 500. Kern is -100.
	// Total units: 500 + 500 - 100 = 900.
	// At size 10, width should be 900 * 10 / 1000 = 9.0
	w := f.TextWidth("AV", 10)
	if w != 9.0 {
		t.Errorf("expected width 9.0, got %f", w)
	}
}

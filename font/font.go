package font

// Font provides glyph metrics and encoding for text measurement and PDF font embedding.
type Font interface {
	// PostScriptName returns the font's PostScript name (e.g. "Helvetica-Bold").
	PostScriptName() string

	// UnitsPerEm returns the number of font units per em square.
	UnitsPerEm() int

	// Ascent returns the typographic ascender in font units.
	Ascent() int

	// Descent returns the typographic descender in font units (typically negative).
	Descent() int

	// LineGap returns the typographic line gap in font units.
	LineGap() int

	// GlyphWidth returns the advance width for the given rune in font units.
	// Returns 0 if the rune is not mapped in this font.
	GlyphWidth(r rune) int

	// TextWidth measures the total advance width of text in points at the given fontSize.
	TextWidth(text string, fontSize float64) float64

	// Metrics returns full font metrics.
	Metrics() Metrics
}

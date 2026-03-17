package truetype

import "gpdf/font"

// Font is a parsed TrueType font providing glyph metrics and character mapping.
type Font struct {
	raw    []byte
	tables map[string]tableRecord

	// head
	unitsPerEm       int
	bbox             [4]float64
	indexToLocFormat int

	// maxp
	numGlyphs int

	// hhea
	ascent      int
	descent     int
	lineGap     int
	numHMetrics int

	// hmtx: per-glyph advance widths indexed by glyph ID
	glyphWidths []uint16

	// cmap: rune to glyph ID mapping
	runeToGlyph map[rune]uint16

	// OS/2
	os2WeightClass int
	sTypoAscender  int
	sTypoDescender int
	sTypoLineGap   int
	sCapHeight     int
	sxHeight       int

	// name
	psName string

	// post
	italicAngle float64
}

// PostScriptName returns the font's PostScript name from the name table.
func (f *Font) PostScriptName() string { return f.psName }

// UnitsPerEm returns the number of font design units per em square.
func (f *Font) UnitsPerEm() int { return f.unitsPerEm }

// Ascent returns the typographic ascender in font design units.
// Prefers OS/2 sTypoAscender when available, falls back to hhea ascent.
func (f *Font) Ascent() int {
	if f.sTypoAscender != 0 {
		return f.sTypoAscender
	}
	return f.ascent
}

// Descent returns the typographic descender in font design units (typically negative).
func (f *Font) Descent() int {
	if f.sTypoDescender != 0 {
		return f.sTypoDescender
	}
	return f.descent
}

// LineGap returns the typographic line gap in font design units.
func (f *Font) LineGap() int {
	if f.sTypoLineGap != 0 {
		return f.sTypoLineGap
	}
	return f.lineGap
}

// GlyphWidth returns the advance width for the given rune in font design units.
// Returns 0 for unmapped runes.
func (f *Font) GlyphWidth(r rune) int {
	gid, ok := f.runeToGlyph[r]
	if !ok || int(gid) >= len(f.glyphWidths) {
		return 0
	}
	return int(f.glyphWidths[gid])
}

// TextWidth measures the total advance width of text in points at the given fontSize.
func (f *Font) TextWidth(text string, fontSize float64) float64 {
	if f.unitsPerEm == 0 {
		return 0
	}
	var total int
	for _, r := range text {
		total += f.GlyphWidth(r)
	}
	return float64(total) * fontSize / float64(f.unitsPerEm)
}

// Metrics returns full font metrics.
func (f *Font) Metrics() font.Metrics {
	flags := 0
	if f.italicAngle != 0 {
		flags |= 1 << 6 // Italic
	}
	return font.Metrics{
		UnitsPerEm:  f.unitsPerEm,
		Ascent:      f.Ascent(),
		Descent:     f.Descent(),
		LineGap:     f.LineGap(),
		CapHeight:   f.sCapHeight,
		XHeight:     f.sxHeight,
		ItalicAngle: f.italicAngle,
		Flags:       flags,
		BBox:        f.bbox,
		StemV:       80 + (f.os2WeightClass-400)/4,
	}
}

// NumGlyphs returns the total number of glyphs in the font.
func (f *Font) NumGlyphs() int { return f.numGlyphs }

// GlyphID returns the glyph index for the given rune, or 0 (.notdef) if unmapped.
func (f *Font) GlyphID(r rune) uint16 {
	return f.runeToGlyph[r]
}

// Encode converts a Unicode string to a byte sequence of 2-byte big-endian glyph IDs.
func (f *Font) Encode(text string) []byte {
	out := make([]byte, 0, len(text)*2)
	for _, r := range text {
		gid := f.runeToGlyph[r] // 0 (.notdef) for unmapped runes
		out = append(out, byte(gid>>8), byte(gid))
	}
	return out
}

// CIDWidths returns a map from glyph ID to advance width (in font units)
// for all glyph IDs in usedGlyphIDs.
func (f *Font) CIDWidths(usedGlyphIDs map[uint16]bool) map[uint16]int {
	out := make(map[uint16]int, len(usedGlyphIDs))
	for gid := range usedGlyphIDs {
		if int(gid) < len(f.glyphWidths) {
			out[gid] = int(f.glyphWidths[gid])
		}
	}
	return out
}

// Ensure *Font implements font.Font and font.EmbeddableFont at compile time.
var _ font.Font = (*Font)(nil)
var _ font.EmbeddableFont = (*Font)(nil)

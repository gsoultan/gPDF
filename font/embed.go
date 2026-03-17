package font

// EmbeddableFont extends Font with methods needed for PDF font embedding.
type EmbeddableFont interface {
	Font

	// Encode converts a Unicode text string to a byte sequence of 2-byte
	// big-endian glyph IDs suitable for a CID font content stream.
	Encode(text string) []byte

	// Subset returns a TrueType font program containing only the glyphs
	// in usedGlyphIDs (plus .notdef). Glyph IDs are preserved.
	Subset(usedGlyphIDs map[uint16]bool) ([]byte, error)

	// ToUnicodeCMap generates a CMap stream mapping glyph IDs to Unicode
	// code points for text extraction in PDF viewers.
	ToUnicodeCMap(usedRunes map[rune]uint16) []byte

	// CIDWidths returns a map of glyph ID to advance width in font units
	// for all glyph IDs present in usedGlyphIDs.
	CIDWidths(usedGlyphIDs map[uint16]bool) map[uint16]int

	// GlyphID returns the glyph index for the given rune, or 0 if unmapped.
	GlyphID(r rune) uint16
}

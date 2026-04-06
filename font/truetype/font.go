package truetype

import (
	"encoding/binary"
	"github.com/gsoultan/gpdf/font"
)

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

	// kern: packed (gid1 << 16 | gid2) to kerning adjustment in font units
	kernPairs map[uint32]int16

	// GSUB: ligature substitution
	gsub *GSUB
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

// Contains returns true if the font supports the given rune.
func (f *Font) Contains(r rune) bool {
	_, ok := f.runeToGlyph[r]
	return ok
}

// Kern returns the kerning adjustment between two runes in font design units.
func (f *Font) Kern(r1, r2 rune) int {
	if f.kernPairs == nil {
		return 0
	}
	gid1 := f.runeToGlyph[r1]
	gid2 := f.runeToGlyph[r2]
	if gid1 == 0 || gid2 == 0 {
		return 0
	}
	key := uint32(gid1)<<16 | uint32(gid2)
	return int(f.kernPairs[key])
}

// TextWidth measures the total advance width of text in points at the given fontSize.
func (f *Font) TextWidth(text string, fontSize float64) float64 {
	if f.unitsPerEm == 0 {
		return 0
	}
	gids := make([]uint16, 0, len(text))
	for _, r := range text {
		gids = append(gids, f.runeToGlyph[r])
	}
	if f.gsub != nil && len(f.gsub.Ligatures) > 0 {
		gids = f.applyLigatures(gids)
	}

	var total int
	for i, gid := range gids {
		if int(gid) < len(f.glyphWidths) {
			total += int(f.glyphWidths[gid])
		}
		if i > 0 {
			// Kerning is generally not applied to ligatures in the same way,
			// but we'll apply it between the ligature and adjacent glyphs.
			// Actually, kernPairs uses GIDs.
			key := uint32(gids[i-1])<<16 | uint32(gids[i])
			total += int(f.kernPairs[key])
		}
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
// It also applies ligature substitutions from the GSUB table if available.
func (f *Font) Encode(text string) []byte {
	gids := make([]uint16, 0, len(text))
	for _, r := range text {
		gids = append(gids, f.runeToGlyph[r])
	}

	if f.gsub != nil && len(f.gsub.Ligatures) > 0 {
		gids = f.applyLigatures(gids)
	}

	out := make([]byte, 0, len(gids)*2)
	for _, gid := range gids {
		out = append(out, byte(gid>>8), byte(gid))
	}
	return out
}

func (f *Font) applyLigatures(gids []uint16) []uint16 {
	if len(gids) < 2 {
		return gids
	}
	var res []uint16
	for i := 0; i < len(gids); {
		// Try 3-glyph ligature
		if i+2 < len(gids) {
			key := make([]byte, 6)
			binary.BigEndian.PutUint16(key[0:2], gids[i])
			binary.BigEndian.PutUint16(key[2:4], gids[i+1])
			binary.BigEndian.PutUint16(key[4:6], gids[i+2])
			if lig, ok := f.gsub.Ligatures[string(key)]; ok {
				res = append(res, lig)
				i += 3
				continue
			}
		}
		// Try 2-glyph ligature
		if i+1 < len(gids) {
			key := make([]byte, 4)
			binary.BigEndian.PutUint16(key[0:2], gids[i])
			binary.BigEndian.PutUint16(key[2:4], gids[i+1])
			if lig, ok := f.gsub.Ligatures[string(key)]; ok {
				res = append(res, lig)
				i += 2
				continue
			}
		}
		res = append(res, gids[i])
		i++
	}
	return res
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

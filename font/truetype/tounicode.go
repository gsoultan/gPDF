package truetype

import (
	"fmt"
	"sort"
	"strings"
)

// ToUnicodeCMap generates a PDF ToUnicode CMap stream that maps glyph IDs (CIDs)
// back to Unicode code points, enabling text search and copy in PDF viewers.
func (f *Font) ToUnicodeCMap(usedRunes map[rune]uint16) []byte {
	type mapping struct {
		gid  uint16
		code rune
	}
	var mappings []mapping
	for r, gid := range usedRunes {
		if gid == 0 {
			continue
		}
		mappings = append(mappings, mapping{gid: gid, code: r})
	}
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].gid < mappings[j].gid
	})

	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\n")
	b.WriteString("begincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	b.WriteString("/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n")
	b.WriteString("<0000> <FFFF>\n")
	b.WriteString("endcodespacerange\n")

	// Write bfchar entries in batches of 100 (PDF limit per beginbfchar block).
	for i := 0; i < len(mappings); {
		end := i + 100
		if end > len(mappings) {
			end = len(mappings)
		}
		fmt.Fprintf(&b, "%d beginbfchar\n", end-i)
		for _, m := range mappings[i:end] {
			if m.code <= 0xFFFF {
				fmt.Fprintf(&b, "<%04X> <%04X>\n", m.gid, m.code)
			} else {
				// Supplementary plane: encode as UTF-16 surrogate pair.
				hi, lo := surrogates(m.code)
				fmt.Fprintf(&b, "<%04X> <%04X%04X>\n", m.gid, hi, lo)
			}
		}
		b.WriteString("endbfchar\n")
		i = end
	}

	b.WriteString("endcmap\n")
	b.WriteString("CMapName currentdict /CMap defineresource pop\n")
	b.WriteString("end\nend\n")
	return []byte(b.String())
}

func surrogates(r rune) (hi, lo uint16) {
	r -= 0x10000
	hi = 0xD800 + uint16(r>>10)
	lo = 0xDC00 + uint16(r&0x3FF)
	return
}

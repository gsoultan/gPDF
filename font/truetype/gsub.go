package truetype

import (
	"encoding/binary"
	"fmt"
)

// GSUB holds simplified ligature substitution data.
type GSUB struct {
	// Key is a packed sequence of glyph IDs, value is the ligature glyph ID.
	// We only support ligatures of 2-3 glyphs for now.
	Ligatures map[string]uint16
}

func (f *Font) parseGSUB() error {
	d, err := f.tableData("GSUB")
	if err != nil {
		return nil // Optional
	}
	if len(d) < 10 {
		return fmt.Errorf("truetype: GSUB table too short")
	}

	lookupListOff := binary.BigEndian.Uint16(d[8:10])
	if int(lookupListOff)+2 > len(d) {
		return fmt.Errorf("truetype: GSUB lookup list offset out of bounds")
	}

	numLookups := int(binary.BigEndian.Uint16(d[lookupListOff : lookupListOff+2]))
	f.gsub = &GSUB{Ligatures: make(map[string]uint16)}

	for i := range numLookups {
		off := int(lookupListOff) + 2 + i*2
		lookupOff := int(binary.BigEndian.Uint16(d[off : off+2]))
		if int(lookupListOff)+lookupOff+6 > len(d) {
			continue
		}
		absLookupOff := int(lookupListOff) + lookupOff
		lookupType := binary.BigEndian.Uint16(d[absLookupOff : absLookupOff+2])
		if lookupType != 4 { // Ligature Substitution
			continue
		}

		numSubtables := int(binary.BigEndian.Uint16(d[absLookupOff+4 : absLookupOff+6]))
		for j := range numSubtables {
			subOff := int(binary.BigEndian.Uint16(d[absLookupOff+6+j*2 : absLookupOff+8+j*2]))
			f.parseGSUBType4(d[absLookupOff+subOff:])
		}
	}
	return nil
}

func (f *Font) parseGSUBType4(d []byte) {
	if len(d) < 6 {
		return
	}
	// format := binary.BigEndian.Uint16(d[0:2])
	coverageOff := binary.BigEndian.Uint16(d[2:4])
	numSets := int(binary.BigEndian.Uint16(d[4:6]))
	if len(d) < 6+numSets*2 {
		return
	}

	// Coverage table tells us which glyph IDs start a ligature
	starts := f.parseCoverage(d[coverageOff:])
	if len(starts) != numSets {
		return
	}

	for i, startGID := range starts {
		setOff := int(binary.BigEndian.Uint16(d[6+i*2 : 8+i*2]))
		if setOff+2 > len(d) {
			continue
		}
		numLigs := int(binary.BigEndian.Uint16(d[setOff : setOff+2]))
		for k := range numLigs {
			ligOff := int(binary.BigEndian.Uint16(d[setOff+2+k*2 : setOff+4+k*2]))
			f.parseLigature(d[setOff+ligOff:], startGID)
		}
	}
}

func (f *Font) parseCoverage(d []byte) []uint16 {
	if len(d) < 4 {
		return nil
	}
	format := binary.BigEndian.Uint16(d[0:2])
	if format == 1 {
		numGlyphs := int(binary.BigEndian.Uint16(d[2:4]))
		if len(d) < 4+numGlyphs*2 {
			return nil
		}
		res := make([]uint16, numGlyphs)
		for i := range numGlyphs {
			res[i] = binary.BigEndian.Uint16(d[4+i*2 : 6+i*2])
		}
		return res
	} else if format == 2 {
		numRanges := int(binary.BigEndian.Uint16(d[2:4]))
		if len(d) < 4+numRanges*6 {
			return nil
		}
		var res []uint16
		for i := range numRanges {
			start := binary.BigEndian.Uint16(d[4+i*6 : 6+i*6])
			end := binary.BigEndian.Uint16(d[6+i*6 : 8+i*6])
			// startIdx := binary.BigEndian.Uint16(d[8+i*6 : 10+i*6])
			for g := start; g <= end; g++ {
				res = append(res, g)
			}
		}
		return res
	}
	return nil
}

func (f *Font) parseLigature(d []byte, startGID uint16) {
	if len(d) < 4 {
		return
	}
	ligGlyph := binary.BigEndian.Uint16(d[0:2])
	compCount := int(binary.BigEndian.Uint16(d[2:4]))
	if len(d) < 4+(compCount-1)*2 {
		return
	}
	key := make([]byte, compCount*2)
	binary.BigEndian.PutUint16(key[0:2], startGID)
	for i := range compCount - 1 {
		gid := binary.BigEndian.Uint16(d[4+i*2 : 6+i*2])
		binary.BigEndian.PutUint16(key[2+i*2:4+i*2], gid)
	}
	f.gsub.Ligatures[string(key)] = ligGlyph
}

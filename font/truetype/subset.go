package truetype

import (
	"encoding/binary"
	"fmt"
)

// Subset returns a new TrueType font program containing only the glyphs
// referenced by usedGlyphIDs plus glyph 0 (.notdef). Glyph IDs are preserved
// (unused entries are zeroed in glyf), keeping /CIDToGIDMap /Identity valid.
func (f *Font) Subset(usedGlyphIDs map[uint16]bool) ([]byte, error) {
	glyfData, err := f.tableData("glyf")
	if err != nil {
		return nil, err
	}
	locaData, err := f.tableData("loca")
	if err != nil {
		return nil, err
	}
	offsets := f.parseLocaOffsets(locaData)
	if len(offsets) < f.numGlyphs+1 {
		return nil, fmt.Errorf("truetype: loca table has too few entries")
	}

	// Collect component glyphs for composites.
	allGlyphs := make(map[uint16]bool, len(usedGlyphIDs)+1)
	allGlyphs[0] = true
	for gid := range usedGlyphIDs {
		allGlyphs[gid] = true
	}
	for gid := range usedGlyphIDs {
		f.collectCompositeComponents(glyfData, offsets, gid, allGlyphs)
	}

	// Build new glyf: keep used glyph data, zero-out unused.
	var newGlyf []byte
	newLoca := make([]uint32, f.numGlyphs+1)
	for gid := range f.numGlyphs {
		newLoca[gid] = uint32(len(newGlyf))
		start := offsets[gid]
		end := offsets[gid+1]
		if allGlyphs[uint16(gid)] && end > start {
			newGlyf = append(newGlyf, glyfData[start:end]...)
		}
		// Pad to even boundary.
		if len(newGlyf)%2 != 0 {
			newGlyf = append(newGlyf, 0)
		}
	}
	newLoca[f.numGlyphs] = uint32(len(newGlyf))

	return f.rebuildFont(newGlyf, newLoca)
}

func (f *Font) parseLocaOffsets(data []byte) []uint32 {
	n := f.numGlyphs + 1
	offsets := make([]uint32, n)
	switch f.indexToLocFormat {
	case 0: // short format: uint16, multiply by 2
		for i := range n {
			if i*2+2 > len(data) {
				break
			}
			offsets[i] = uint32(binary.BigEndian.Uint16(data[i*2:i*2+2])) * 2
		}
	default: // long format: uint32
		for i := range n {
			if i*4+4 > len(data) {
				break
			}
			offsets[i] = binary.BigEndian.Uint32(data[i*4 : i*4+4])
		}
	}
	return offsets
}

func (f *Font) collectCompositeComponents(glyfData []byte, offsets []uint32, gid uint16, out map[uint16]bool) {
	if int(gid) >= f.numGlyphs {
		return
	}
	start := offsets[gid]
	end := offsets[gid+1]
	if end <= start || int(end) > len(glyfData) {
		return
	}
	data := glyfData[start:end]
	if len(data) < 10 {
		return
	}
	numContours := int16(binary.BigEndian.Uint16(data[0:2]))
	if numContours >= 0 {
		return // simple glyph
	}
	// Composite glyph: parse component records starting at offset 10.
	off := 10
	for off+4 <= len(data) {
		flags := binary.BigEndian.Uint16(data[off : off+2])
		componentGID := binary.BigEndian.Uint16(data[off+2 : off+4])
		if !out[componentGID] {
			out[componentGID] = true
			f.collectCompositeComponents(glyfData, offsets, componentGID, out)
		}
		off += 4
		// Argument size depends on flags.
		if flags&0x0001 != 0 { // ARG_1_AND_2_ARE_WORDS
			off += 4
		} else {
			off += 2
		}
		if flags&0x0008 != 0 { // WE_HAVE_A_SCALE
			off += 2
		} else if flags&0x0040 != 0 { // WE_HAVE_AN_X_AND_Y_SCALE
			off += 4
		} else if flags&0x0080 != 0 { // WE_HAVE_A_TWO_BY_TWO
			off += 8
		}
		if flags&0x0020 == 0 { // no MORE_COMPONENTS
			break
		}
	}
}

// rebuildFont constructs a valid TrueType file with updated glyf and loca tables.
func (f *Font) rebuildFont(newGlyf []byte, newLoca []uint32) ([]byte, error) {
	// Tables to include in the subset font.
	keepTags := []string{
		"head", "hhea", "maxp", "OS/2", "name", "cmap", "post",
		"cvt ", "fpgm", "prep",
	}
	type tableEntry struct {
		tag  string
		data []byte
	}
	var entries []tableEntry
	for _, tag := range keepTags {
		d, err := f.tableData(tag)
		if err != nil {
			continue // optional tables may be absent
		}
		cp := make([]byte, len(d))
		copy(cp, d)
		entries = append(entries, tableEntry{tag: tag, data: cp})
	}

	// Build loca table.
	locaBytes := f.buildLocaBytes(newLoca)
	entries = append(entries, tableEntry{tag: "loca", data: locaBytes})
	entries = append(entries, tableEntry{tag: "glyf", data: newGlyf})

	// Build hmtx with original widths (glyph IDs preserved).
	hmtxData, err := f.tableData("hmtx")
	if err == nil {
		cp := make([]byte, len(hmtxData))
		copy(cp, hmtxData)
		entries = append(entries, tableEntry{tag: "hmtx", data: cp})
	}

	numTables := len(entries)
	headerSize := 12 + numTables*16

	// Compute table offsets.
	offset := uint32(headerSize)
	offsets := make([]uint32, numTables)
	for i, e := range entries {
		offsets[i] = offset
		offset += uint32(len(e.data))
		if offset%4 != 0 {
			offset += 4 - (offset % 4) // pad to 4-byte boundary
		}
	}

	// Build the font file.
	out := make([]byte, 0, offset)

	// Offset table.
	var header [12]byte
	binary.BigEndian.PutUint32(header[0:4], 0x00010000) // sfVersion
	binary.BigEndian.PutUint16(header[4:6], uint16(numTables))
	// searchRange, entrySelector, rangeShift
	sr := 1
	es := 0
	for sr*2 <= numTables {
		sr *= 2
		es++
	}
	binary.BigEndian.PutUint16(header[6:8], uint16(sr*16))
	binary.BigEndian.PutUint16(header[8:10], uint16(es))
	binary.BigEndian.PutUint16(header[10:12], uint16(numTables*16-sr*16))
	out = append(out, header[:]...)

	// Table directory.
	for i, e := range entries {
		var rec [16]byte
		copy(rec[0:4], e.tag)
		binary.BigEndian.PutUint32(rec[4:8], tableChecksum(e.data))
		binary.BigEndian.PutUint32(rec[8:12], offsets[i])
		binary.BigEndian.PutUint32(rec[12:16], uint32(len(e.data)))
		out = append(out, rec[:]...)
	}

	// Table data.
	for _, e := range entries {
		out = append(out, e.data...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
	}

	// Update head checkSumAdjustment.
	for i, e := range entries {
		if e.tag == "head" && len(e.data) >= 12 {
			adj := 0xB1B0AFBA - tableChecksum(out)
			headOff := offsets[i]
			_ = i
			binary.BigEndian.PutUint32(out[headOff+8:headOff+12], adj)
			break
		}
	}

	return out, nil
}

func (f *Font) buildLocaBytes(loca []uint32) []byte {
	switch f.indexToLocFormat {
	case 0: // short
		b := make([]byte, len(loca)*2)
		for i, off := range loca {
			binary.BigEndian.PutUint16(b[i*2:i*2+2], uint16(off/2))
		}
		return b
	default: // long
		b := make([]byte, len(loca)*4)
		for i, off := range loca {
			binary.BigEndian.PutUint32(b[i*4:i*4+4], off)
		}
		return b
	}
}

func tableChecksum(data []byte) uint32 {
	var sum uint32
	nLongs := (len(data) + 3) / 4
	for i := range nLongs {
		var val uint32
		off := i * 4
		for j := range 4 {
			val <<= 8
			if off+j < len(data) {
				val |= uint32(data[off+j])
			}
		}
		sum += val
	}
	return sum
}

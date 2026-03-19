package truetype

import (
	"encoding/binary"
	"fmt"
)

// tableRecord locates a single table within the font file.
type tableRecord struct {
	tag      string
	offset   uint32
	length   uint32
	checksum uint32
}

// Parse reads a TrueType (.ttf) or OpenType (.otf) font from raw bytes and returns a Font.
// It extracts glyph metrics, character-to-glyph mappings, and naming information.
func Parse(data []byte) (*Font, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("truetype: data too short for offset table")
	}
	scaler := binary.BigEndian.Uint32(data[0:4])
	if scaler == 0x74746366 { // 'ttcf'
		return nil, fmt.Errorf("truetype: use ParseCollection for .ttc files")
	}

	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	if len(data) < 12+numTables*16 {
		return nil, fmt.Errorf("truetype: data too short for table directory")
	}
	tables := make(map[string]tableRecord, numTables)
	for i := range numTables {
		off := 12 + i*16
		tag := string(data[off : off+4])
		tables[tag] = tableRecord{
			tag:      tag,
			checksum: binary.BigEndian.Uint32(data[off+4 : off+8]),
			offset:   binary.BigEndian.Uint32(data[off+8 : off+12]),
			length:   binary.BigEndian.Uint32(data[off+12 : off+16]),
		}
	}
	f := &Font{
		raw:    data,
		tables: tables,
	}
	return f.initialize()
}

// ParseCollection reads a TrueType Collection (.ttc) font and returns all individual fonts.
func ParseCollection(data []byte) ([]*Font, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("truetype: data too short for collection header")
	}
	if string(data[0:4]) != "ttcf" {
		return nil, fmt.Errorf("truetype: not a font collection")
	}
	numFonts := int(binary.BigEndian.Uint32(data[8:12]))
	if len(data) < 12+numFonts*4 {
		return nil, fmt.Errorf("truetype: data too short for collection offset table")
	}

	var fonts []*Font
	for i := range numFonts {
		offset := binary.BigEndian.Uint32(data[12+i*4 : 16+i*4])
		f, err := parseAtOffset(data, offset)
		if err != nil {
			return nil, fmt.Errorf("truetype: font %d in collection: %w", i, err)
		}
		fonts = append(fonts, f)
	}
	return fonts, nil
}

func parseAtOffset(data []byte, start uint32) (*Font, error) {
	if int(start+12) > len(data) {
		return nil, fmt.Errorf("truetype: offset out of bounds")
	}
	numTables := int(binary.BigEndian.Uint16(data[start+4 : start+6]))
	if int(start+12+uint32(numTables)*16) > len(data) {
		return nil, fmt.Errorf("truetype: table directory out of bounds")
	}

	tables := make(map[string]tableRecord, numTables)
	for i := range numTables {
		off := start + 12 + uint32(i)*16
		tag := string(data[off : off+4])
		tables[tag] = tableRecord{
			tag:      tag,
			checksum: binary.BigEndian.Uint32(data[off+4 : off+8]),
			offset:   binary.BigEndian.Uint32(data[off+8 : off+12]),
			length:   binary.BigEndian.Uint32(data[off+12 : off+16]),
		}
	}
	f := &Font{
		raw:    data,
		tables: tables,
	}
	return f.initialize()
}

func (f *Font) initialize() (*Font, error) {
	if err := f.parseHead(); err != nil {
		return nil, err
	}
	if err := f.parseMaxp(); err != nil {
		return nil, err
	}
	if err := f.parseHhea(); err != nil {
		return nil, err
	}
	if err := f.parseHmtx(); err != nil {
		return nil, err
	}
	if err := f.parseCmap(); err != nil {
		return nil, err
	}
	if err := f.parseOS2(); err != nil {
		// OS/2 is optional; fall back to hhea values
	}
	if err := f.parseName(); err != nil {
		// name table is optional; use empty name
	}
	if err := f.parsePost(); err != nil {
		// post table is optional
	}
	if err := f.parseKern(); err != nil {
		// kern table is optional
	}
	if err := f.parseGSUB(); err != nil {
		// GSUB table is optional
	}
	return f, nil
}

func (f *Font) tableData(tag string) ([]byte, error) {
	rec, ok := f.tables[tag]
	if !ok {
		return nil, fmt.Errorf("truetype: missing '%s' table", tag)
	}
	end := rec.offset + rec.length
	if int(end) > len(f.raw) {
		return nil, fmt.Errorf("truetype: '%s' table out of bounds", tag)
	}
	return f.raw[rec.offset:end], nil
}

func (f *Font) parseHead() error {
	d, err := f.tableData("head")
	if err != nil {
		return err
	}
	if len(d) < 54 {
		return fmt.Errorf("truetype: head table too short")
	}
	f.unitsPerEm = int(binary.BigEndian.Uint16(d[18:20]))
	f.bbox = [4]float64{
		float64(int16(binary.BigEndian.Uint16(d[36:38]))),
		float64(int16(binary.BigEndian.Uint16(d[38:40]))),
		float64(int16(binary.BigEndian.Uint16(d[40:42]))),
		float64(int16(binary.BigEndian.Uint16(d[42:44]))),
	}
	f.indexToLocFormat = int(int16(binary.BigEndian.Uint16(d[50:52])))
	return nil
}

func (f *Font) parseMaxp() error {
	d, err := f.tableData("maxp")
	if err != nil {
		return err
	}
	if len(d) < 6 {
		return fmt.Errorf("truetype: maxp table too short")
	}
	f.numGlyphs = int(binary.BigEndian.Uint16(d[4:6]))
	return nil
}

func (f *Font) parseHhea() error {
	d, err := f.tableData("hhea")
	if err != nil {
		return err
	}
	if len(d) < 36 {
		return fmt.Errorf("truetype: hhea table too short")
	}
	f.ascent = int(int16(binary.BigEndian.Uint16(d[4:6])))
	f.descent = int(int16(binary.BigEndian.Uint16(d[6:8])))
	f.lineGap = int(int16(binary.BigEndian.Uint16(d[8:10])))
	f.numHMetrics = int(binary.BigEndian.Uint16(d[34:36]))
	return nil
}

func (f *Font) parseHmtx() error {
	d, err := f.tableData("hmtx")
	if err != nil {
		return err
	}
	needed := f.numHMetrics * 4
	if f.numGlyphs > f.numHMetrics {
		needed += (f.numGlyphs - f.numHMetrics) * 2
	}
	if len(d) < needed {
		return fmt.Errorf("truetype: hmtx table too short")
	}
	f.glyphWidths = make([]uint16, f.numGlyphs)
	var lastWidth uint16
	for i := range f.numHMetrics {
		w := binary.BigEndian.Uint16(d[i*4 : i*4+2])
		f.glyphWidths[i] = w
		lastWidth = w
	}
	for i := f.numHMetrics; i < f.numGlyphs; i++ {
		f.glyphWidths[i] = lastWidth
	}
	return nil
}

func (f *Font) parseCmap() error {
	d, err := f.tableData("cmap")
	if err != nil {
		return err
	}
	if len(d) < 4 {
		return fmt.Errorf("truetype: cmap table too short")
	}
	numSubtables := int(binary.BigEndian.Uint16(d[2:4]))
	if len(d) < 4+numSubtables*8 {
		return fmt.Errorf("truetype: cmap subtable records truncated")
	}
	// Prefer: platform 3 encoding 10 (Windows UCS-4, Format 12),
	// then platform 3 encoding 1 (Windows BMP, Format 4),
	// then platform 0 (Unicode).
	var bestOffset uint32
	bestPriority := -1
	for i := range numSubtables {
		off := 4 + i*8
		platformID := binary.BigEndian.Uint16(d[off : off+2])
		encodingID := binary.BigEndian.Uint16(d[off+2 : off+4])
		subtableOff := binary.BigEndian.Uint32(d[off+4 : off+8])
		priority := -1
		switch {
		case platformID == 3 && encodingID == 10:
			priority = 3
		case platformID == 3 && encodingID == 1:
			priority = 2
		case platformID == 0:
			priority = 1
		}
		if priority > bestPriority {
			bestPriority = priority
			bestOffset = subtableOff
		}
	}
	if bestPriority < 0 {
		return fmt.Errorf("truetype: no usable cmap subtable found")
	}
	return f.parseCmapSubtable(d, bestOffset)
}

func (f *Font) parseCmapSubtable(d []byte, offset uint32) error {
	if int(offset)+2 > len(d) {
		return fmt.Errorf("truetype: cmap subtable offset out of bounds")
	}
	format := binary.BigEndian.Uint16(d[offset : offset+2])
	switch format {
	case 4:
		return f.parseCmapFormat4(d, offset)
	case 12:
		return f.parseCmapFormat12(d, offset)
	default:
		return fmt.Errorf("truetype: unsupported cmap format %d", format)
	}
}

func (f *Font) parseCmapFormat4(d []byte, offset uint32) error {
	base := int(offset)
	if base+14 > len(d) {
		return fmt.Errorf("truetype: cmap format 4 header truncated")
	}
	segCountX2 := int(binary.BigEndian.Uint16(d[base+6 : base+8]))
	segCount := segCountX2 / 2
	headerSize := 14 + segCount*8 + 2
	if base+headerSize > len(d) {
		return fmt.Errorf("truetype: cmap format 4 data truncated")
	}
	endCodes := base + 14
	startCodes := endCodes + segCount*2 + 2
	idDeltas := startCodes + segCount*2
	idRangeOffsets := idDeltas + segCount*2

	f.runeToGlyph = make(map[rune]uint16, 256)
	for i := range segCount {
		endCode := int(binary.BigEndian.Uint16(d[endCodes+i*2 : endCodes+i*2+2]))
		startCode := int(binary.BigEndian.Uint16(d[startCodes+i*2 : startCodes+i*2+2]))
		idDelta := int(int16(binary.BigEndian.Uint16(d[idDeltas+i*2 : idDeltas+i*2+2])))
		rangeOff := int(binary.BigEndian.Uint16(d[idRangeOffsets+i*2 : idRangeOffsets+i*2+2]))
		if startCode == 0xFFFF {
			break
		}
		for c := startCode; c <= endCode; c++ {
			var glyphID uint16
			if rangeOff == 0 {
				glyphID = uint16((c + idDelta) & 0xFFFF)
			} else {
				glyphIndexOff := idRangeOffsets + i*2 + rangeOff + (c-startCode)*2
				if glyphIndexOff+2 > len(d) {
					continue
				}
				glyphID = binary.BigEndian.Uint16(d[glyphIndexOff : glyphIndexOff+2])
				if glyphID != 0 {
					glyphID = uint16((int(glyphID) + idDelta) & 0xFFFF)
				}
			}
			if glyphID != 0 {
				f.runeToGlyph[rune(c)] = glyphID
			}
		}
	}
	return nil
}

func (f *Font) parseCmapFormat12(d []byte, offset uint32) error {
	base := int(offset)
	if base+16 > len(d) {
		return fmt.Errorf("truetype: cmap format 12 header truncated")
	}
	numGroups := int(binary.BigEndian.Uint32(d[base+12 : base+16]))
	if base+16+numGroups*12 > len(d) {
		return fmt.Errorf("truetype: cmap format 12 data truncated")
	}
	f.runeToGlyph = make(map[rune]uint16, numGroups*4)
	for i := range numGroups {
		off := base + 16 + i*12
		startChar := binary.BigEndian.Uint32(d[off : off+4])
		endChar := binary.BigEndian.Uint32(d[off+4 : off+8])
		startGlyph := binary.BigEndian.Uint32(d[off+8 : off+12])
		for c := startChar; c <= endChar; c++ {
			gid := startGlyph + (c - startChar)
			if gid > 0 && gid < 0xFFFF {
				f.runeToGlyph[rune(c)] = uint16(gid)
			}
		}
	}
	return nil
}

func (f *Font) parseOS2() error {
	d, err := f.tableData("OS/2")
	if err != nil {
		return err
	}
	if len(d) < 78 {
		return fmt.Errorf("truetype: OS/2 table too short")
	}
	f.os2WeightClass = int(binary.BigEndian.Uint16(d[4:6]))
	f.sTypoAscender = int(int16(binary.BigEndian.Uint16(d[68:70])))
	f.sTypoDescender = int(int16(binary.BigEndian.Uint16(d[70:72])))
	f.sTypoLineGap = int(int16(binary.BigEndian.Uint16(d[72:74])))
	if len(d) >= 90 {
		f.sCapHeight = int(int16(binary.BigEndian.Uint16(d[88:90])))
	}
	if len(d) >= 92 {
		f.sxHeight = int(int16(binary.BigEndian.Uint16(d[86:88])))
	}
	return nil
}

func (f *Font) parseName() error {
	d, err := f.tableData("name")
	if err != nil {
		return err
	}
	if len(d) < 6 {
		return fmt.Errorf("truetype: name table too short")
	}
	count := int(binary.BigEndian.Uint16(d[2:4]))
	stringOffset := int(binary.BigEndian.Uint16(d[4:6]))
	if len(d) < 6+count*12 {
		return fmt.Errorf("truetype: name records truncated")
	}
	for i := range count {
		off := 6 + i*12
		nameID := binary.BigEndian.Uint16(d[off+6 : off+8])
		if nameID != 6 {
			continue
		}
		platformID := binary.BigEndian.Uint16(d[off : off+2])
		length := int(binary.BigEndian.Uint16(d[off+8 : off+10]))
		strOff := stringOffset + int(binary.BigEndian.Uint16(d[off+10:off+12]))
		if strOff+length > len(d) {
			continue
		}
		raw := d[strOff : strOff+length]
		switch platformID {
		case 3:
			// Windows: UTF-16BE
			f.psName = decodeUTF16BE(raw)
			return nil
		case 1:
			// Macintosh: MacRoman (approximate as ASCII)
			f.psName = string(raw)
			return nil
		}
	}
	return nil
}

func (f *Font) parsePost() error {
	d, err := f.tableData("post")
	if err != nil {
		return err
	}
	if len(d) < 32 {
		return fmt.Errorf("truetype: post table too short")
	}
	// italicAngle is a 16.16 fixed-point number
	whole := int16(binary.BigEndian.Uint16(d[4:6]))
	frac := binary.BigEndian.Uint16(d[6:8])
	f.italicAngle = float64(whole) + float64(frac)/65536.0
	return nil
}

func (f *Font) parseKern() error {
	d, err := f.tableData("kern")
	if err != nil {
		return nil // kern table is optional
	}
	if len(d) < 4 {
		return nil
	}
	version := binary.BigEndian.Uint16(d[0:2])
	numSubtables := int(binary.BigEndian.Uint16(d[2:4]))
	if version != 0 {
		return nil // only support version 0 kern table
	}

	off := 4
	for range numSubtables {
		if off+6 > len(d) {
			break
		}
		// subtableVersion := binary.BigEndian.Uint16(d[off:off+2])
		length := int(binary.BigEndian.Uint16(d[off+2 : off+4]))
		coverage := binary.BigEndian.Uint16(d[off+4 : off+6])

		// Format 0: Horizontal kerning
		// coverage bits: 0: horizontal, 1: minimum, 2: cross-stream, 3: override
		if (coverage&0xFF00) == 0 && (coverage&0x01 != 0) {
			if off+14 <= len(d) {
				nPairs := int(binary.BigEndian.Uint16(d[off+6 : off+8]))
				if f.kernPairs == nil {
					f.kernPairs = make(map[uint32]int16, nPairs)
				}
				pOff := off + 14
				for range nPairs {
					if pOff+6 > off+length || pOff+6 > len(d) {
						break
					}
					left := binary.BigEndian.Uint16(d[pOff : pOff+2])
					right := binary.BigEndian.Uint16(d[pOff+2 : pOff+4])
					value := int16(binary.BigEndian.Uint16(d[pOff+4 : pOff+6]))
					f.kernPairs[uint32(left)<<16|uint32(right)] = value
					pOff += 6
				}
			}
		}
		off += length
	}
	return nil
}

func decodeUTF16BE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		r := rune(binary.BigEndian.Uint16(b[i : i+2]))
		runes = append(runes, r)
	}
	return string(runes)
}

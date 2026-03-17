package doc

import (
	"crypto/rand"
	"sort"

	"gpdf/font"
	"gpdf/model"
)

// embeddedFontUsage tracks rune usage for one registered embeddable font.
type embeddedFontUsage struct {
	font      font.EmbeddableFont
	usedRunes map[rune]uint16 // rune -> glyph ID
	usedGIDs  map[uint16]bool // glyph IDs used
}

func newEmbeddedFontUsage(f font.EmbeddableFont) *embeddedFontUsage {
	return &embeddedFontUsage{
		font:      f,
		usedRunes: make(map[rune]uint16),
		usedGIDs:  make(map[uint16]bool),
	}
}

func (u *embeddedFontUsage) markText(text string) {
	for _, r := range text {
		if _, ok := u.usedRunes[r]; ok {
			continue
		}
		gid := u.font.GlyphID(r)
		u.usedRunes[r] = gid
		if gid != 0 {
			u.usedGIDs[gid] = true
		}
	}
}

// embeddedFontObjects holds the object numbers allocated for one embedded font.
type embeddedFontObjects struct {
	type0Num      int // Type0 (root) font dict
	cidFontNum    int
	descriptorNum int
	fontFileNum   int
	tounicodeNum  int
}

// buildEmbeddedFontObjects creates all PDF objects for one embedded TrueType font
// and stores them in objs. Returns the object number of the Type0 font dict.
func buildEmbeddedFontObjects(
	usage *embeddedFontUsage,
	objs map[int]model.Object,
	nextNum *int,
) int {
	ef := usage.font
	tag := randomSubsetTag()
	baseName := tag + "+" + ef.PostScriptName()
	m := ef.Metrics()

	// Allocate object numbers.
	type0Num := *nextNum
	*nextNum++
	cidFontNum := *nextNum
	*nextNum++
	descriptorNum := *nextNum
	*nextNum++
	fontFileNum := *nextNum
	*nextNum++
	tounicodeNum := *nextNum
	*nextNum++

	// Subset font program.
	subsetBytes, err := ef.Subset(usage.usedGIDs)
	if err != nil {
		// Fallback: embed full font if subsetting fails.
		subsetBytes = nil
	}

	// FontFile2 stream (subset TrueType program).
	if subsetBytes != nil {
		objs[fontFileNum] = &model.Stream{
			Dict: model.Dict{
				model.Name("Length"): model.Integer(int64(len(subsetBytes))),
				model.Name("Filter"): model.Name("FlateDecode"),
			},
			Content: subsetBytes,
		}
	}

	// ToUnicode CMap stream.
	toUnicodeBytes := ef.ToUnicodeCMap(usage.usedRunes)
	objs[tounicodeNum] = &model.Stream{
		Dict: model.Dict{
			model.Name("Length"): model.Integer(int64(len(toUnicodeBytes))),
		},
		Content: toUnicodeBytes,
	}

	// FontDescriptor.
	scale := 1000.0 / float64(m.UnitsPerEm)
	descriptorDict := model.Dict{
		model.Name("Type"):        model.Name("FontDescriptor"),
		model.Name("FontName"):    model.Name(baseName),
		model.Name("Flags"):       model.Integer(int64(pdfFontFlags(m))),
		model.Name("FontBBox"):    fontBBoxArray(m, scale),
		model.Name("ItalicAngle"): model.Real(m.ItalicAngle),
		model.Name("Ascent"):      model.Integer(int64(float64(m.Ascent) * scale)),
		model.Name("Descent"):     model.Integer(int64(float64(m.Descent) * scale)),
		model.Name("CapHeight"):   model.Integer(int64(float64(m.CapHeight) * scale)),
		model.Name("StemV"):       model.Integer(int64(m.StemV)),
	}
	if subsetBytes != nil {
		descriptorDict[model.Name("FontFile2")] = model.Ref{ObjectNumber: fontFileNum}
	}
	objs[descriptorNum] = descriptorDict

	// CIDFont (DescendantFont).
	widths := ef.CIDWidths(usage.usedGIDs)
	cidFontDict := model.Dict{
		model.Name("Type"):     model.Name("Font"),
		model.Name("Subtype"):  model.Name("CIDFontType2"),
		model.Name("BaseFont"): model.Name(baseName),
		model.Name("CIDSystemInfo"): model.Dict{
			model.Name("Registry"):   model.String("Adobe"),
			model.Name("Ordering"):   model.String("Identity"),
			model.Name("Supplement"): model.Integer(0),
		},
		model.Name("FontDescriptor"): model.Ref{ObjectNumber: descriptorNum},
		model.Name("CIDToGIDMap"):    model.Name("Identity"),
		model.Name("DW"):             model.Integer(1000),
		model.Name("W"):              buildCIDWidthArray(widths, m.UnitsPerEm),
	}
	objs[cidFontNum] = cidFontDict

	// Type0 font (root).
	type0Dict := model.Dict{
		model.Name("Type"):     model.Name("Font"),
		model.Name("Subtype"):  model.Name("Type0"),
		model.Name("BaseFont"): model.Name(baseName),
		model.Name("Encoding"): model.Name("Identity-H"),
		model.Name("DescendantFonts"): model.Array{
			model.Ref{ObjectNumber: cidFontNum},
		},
		model.Name("ToUnicode"): model.Ref{ObjectNumber: tounicodeNum},
	}
	objs[type0Num] = type0Dict

	return type0Num
}

// buildCIDWidthArray builds the /W array for a CIDFont.
// Format: [gid [width] gid [width] ...] with widths scaled to 1000-unit space.
func buildCIDWidthArray(widths map[uint16]int, unitsPerEm int) model.Array {
	if len(widths) == 0 {
		return model.Array{}
	}
	scale := 1000.0 / float64(unitsPerEm)
	gids := make([]int, 0, len(widths))
	for gid := range widths {
		gids = append(gids, int(gid))
	}
	sort.Ints(gids)

	var arr model.Array
	for _, gid := range gids {
		w := int(float64(widths[uint16(gid)]) * scale)
		arr = append(arr,
			model.Integer(int64(gid)),
			model.Array{model.Integer(int64(w))},
		)
	}
	return arr
}

func fontBBoxArray(m font.Metrics, scale float64) model.Array {
	return model.Array{
		model.Integer(int64(m.BBox[0] * scale)),
		model.Integer(int64(m.BBox[1] * scale)),
		model.Integer(int64(m.BBox[2] * scale)),
		model.Integer(int64(m.BBox[3] * scale)),
	}
}

func pdfFontFlags(m font.Metrics) int {
	flags := 0x0004 // Nonsymbolic
	if m.ItalicAngle != 0 {
		flags |= 0x0040 // Italic
	}
	return flags
}

// replaceEmbeddedFontPlaceholders walks page dicts in objs and replaces
// placeholder font entries (marked with /_embedded) with Refs to Type0 font objects.
func (b *DocumentBuilder) replaceEmbeddedFontPlaceholders(objs map[int]model.Object, fontObjNums map[string]int) {
	for _, obj := range objs {
		pageDict, ok := obj.(model.Dict)
		if !ok {
			continue
		}
		resObj, ok := pageDict[model.Name("Resources")]
		if !ok {
			continue
		}
		resDict, ok := resObj.(model.Dict)
		if !ok {
			continue
		}
		fontObj, ok := resDict[model.Name("Font")]
		if !ok {
			continue
		}
		fontDict, ok := fontObj.(model.Dict)
		if !ok {
			continue
		}
		for resName, val := range fontDict {
			d, ok := val.(model.Dict)
			if !ok {
				continue
			}
			embName, ok := d[model.Name("_embedded")].(model.Name)
			if !ok {
				continue
			}
			if objNum, found := fontObjNums[string(embName)]; found {
				fontDict[resName] = model.Ref{ObjectNumber: objNum}
			}
		}
	}
}

func randomSubsetTag() string {
	var b [3]byte
	rand.Read(b[:])
	tag := make([]byte, 6)
	for i := range 6 {
		tag[i] = 'A' + (b[i/2]>>(4*(1-uint(i%2))))&0x0F%26
	}
	return string(tag)
}

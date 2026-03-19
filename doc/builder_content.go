package doc

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"gpdf/content"
	"gpdf/model"
)

// buildGraphicsOps returns content stream ops from graphic runs (backgrounds, lines, shapes).
func (b *DocumentBuilder) buildGraphicsOps(page *pageSpec, pageHeight float64) []content.Op {
	var ops []content.Op
	for _, gr := range page.GraphicRuns {
		ops = append(ops, gr.Ops...)
	}
	return ops
}

// buildTextOps returns content stream ops from text runs and a font resource map (baseName -> resName).
// If embeddedFonts is nil, uses b.fc.embeddedFonts.
func (b *DocumentBuilder) buildTextOps(page *pageSpec, pageHeight float64, embeddedFonts map[string]*embeddedFontUsage) ([]content.Op, map[string]model.Name) {
	fontRes := make(map[string]model.Name)
	var ops []content.Op
	currentColorSet := false
	var currentColor [3]float64

	ef := b.fc.embeddedFonts
	if embeddedFonts != nil {
		ef = embeddedFonts
	}
	for _, r := range page.TextRuns {
		baseName := r.FontName
		if baseName == "" {
			baseName = "Helvetica"
		}
		resName, ok := fontRes[baseName]
		if !ok {
			resName = model.Name(fmt.Sprintf("F%d", len(fontRes)+1))
			fontRes[baseName] = resName
		}
		size := r.FontSize
		if size <= 0 {
			size = 12
		}
		if !r.UseDefaultColor {
			if !currentColorSet || r.TextColorRGB != currentColor {
				ops = append(ops,
					content.Op{
						Name: "rg",
						Args: []model.Object{
							model.Real(r.TextColorRGB[0]),
							model.Real(r.TextColorRGB[1]),
							model.Real(r.TextColorRGB[2]),
						},
					},
				)
				currentColor = r.TextColorRGB
				currentColorSet = true
			}
		}
		if r.HasMCID {
			ops = append(ops,
				content.Op{
					Name: "BDC",
					Args: []model.Object{
						model.Name("Span"),
						model.Dict{model.Name("MCID"): model.Integer(int64(r.MCID))},
					},
				},
			)
		}

		var textArg model.Object
		if eu, ok := ef[baseName]; ok {
			eu.markText(r.Text)
			textArg = model.HexString(eu.font.Encode(r.Text))
		} else {
			// Basic UTF-8 to WinAnsi conversion for common characters like bullet point.
			// This avoids (â€) encoding issues when using standard fonts.
			textArg = model.HexString(winAnsiEncode(r.Text))
		}

		btOps := []content.Op{
			{Name: "BT", Args: nil},
			{Name: "Tf", Args: []model.Object{resName, model.Real(size)}},
		}
		if r.LetterSpacing != 0 {
			btOps = append(btOps, content.Op{
				Name: "Tc",
				Args: []model.Object{model.Real(r.LetterSpacing)},
			})
		}
		if r.WordSpacing != 0 {
			btOps = append(btOps, content.Op{
				Name: "Tw",
				Args: []model.Object{model.Real(r.WordSpacing)},
			})
		}
		btOps = append(btOps,
			content.Op{Name: "Td", Args: []model.Object{model.Real(r.X), model.Real(r.Y)}},
			content.Op{Name: "Tj", Args: []model.Object{textArg}},
			content.Op{Name: "ET", Args: nil},
		)
		ops = append(ops, btOps...)

		if r.Underline {
			tw := b.textWidth(r.Text, size, baseName)
			uy := r.Y - size*0.1
			thick := size * 0.07
			ops = append(ops,
				content.Op{Name: "q"},
				content.Op{Name: "w", Args: []model.Object{model.Real(thick)}},
				content.Op{Name: "RG", Args: []model.Object{
					model.Real(r.TextColorRGB[0]), model.Real(r.TextColorRGB[1]), model.Real(r.TextColorRGB[2]),
				}},
				content.Op{Name: "m", Args: []model.Object{model.Real(r.X), model.Real(uy)}},
				content.Op{Name: "l", Args: []model.Object{model.Real(r.X + tw), model.Real(uy)}},
				content.Op{Name: "S"},
				content.Op{Name: "Q"},
			)
		}
		if r.Strikethrough {
			tw := b.textWidth(r.Text, size, baseName)
			sy := r.Y + size*0.3
			thick := size * 0.07
			ops = append(ops,
				content.Op{Name: "q"},
				content.Op{Name: "w", Args: []model.Object{model.Real(thick)}},
				content.Op{Name: "RG", Args: []model.Object{
					model.Real(r.TextColorRGB[0]), model.Real(r.TextColorRGB[1]), model.Real(r.TextColorRGB[2]),
				}},
				content.Op{Name: "m", Args: []model.Object{model.Real(r.X), model.Real(sy)}},
				content.Op{Name: "l", Args: []model.Object{model.Real(r.X + tw), model.Real(sy)}},
				content.Op{Name: "S"},
				content.Op{Name: "Q"},
			)
		}

		if r.HasMCID {
			ops = append(ops, content.Op{Name: "EMC", Args: nil})
		}
	}
	return ops, fontRes
}

// buildImageOps returns content stream ops from image runs and XObject resources.
// Also returns imageExtGStates (opacity ExtGState dicts) to merge into page resources.
func (b *DocumentBuilder) buildImageOps(page *pageSpec, pageHeight float64, imageStreamNums []int) ([]content.Op, model.Dict, model.Dict) {
	xobj := make(model.Dict)
	imageGS := model.Dict{}
	var ops []content.Op
	imageGSIndex := 0

	for i, im := range page.ImageRuns {
		imName := model.Name("Im" + fmt.Sprintf("%d", i+1))
		w, h := im.WidthPt, im.HeightPt
		if w <= 0 {
			w = float64(im.WidthPx)
		}
		if h <= 0 {
			h = float64(im.HeightPx)
		}
		if i < len(imageStreamNums) {
			xobj[imName] = model.Ref{ObjectNumber: imageStreamNums[i], Generation: 0}
		}
		if im.HasMCID {
			ops = append(ops,
				content.Op{
					Name: "BDC",
					Args: []model.Object{
						model.Name("Span"),
						model.Dict{model.Name("MCID"): model.Integer(int64(im.MCID))},
					},
				},
			)
		}
		ops = append(ops, content.Op{Name: "q", Args: nil})
		if im.Opacity > 0 && im.Opacity < 1 {
			imageGSIndex++
			gsName := model.Name(fmt.Sprintf("IMGS%d", imageGSIndex))
			imageGS[gsName] = model.Dict{
				model.Name("Type"): model.Name("ExtGState"),
				model.Name("ca"):   model.Real(im.Opacity),
				model.Name("CA"):   model.Real(im.Opacity),
			}
			ops = append(ops, content.Op{Name: "gs", Args: []model.Object{gsName}})
		}
		if im.ClipCircle {
			ops = append(ops, circlePathOps(im.ClipCX, im.ClipCY, im.ClipR)...)
			ops = append(ops,
				content.Op{Name: "W", Args: nil},
				content.Op{Name: "n", Args: nil},
			)
		}
		var a, bb, c, d, e, f float64
		if im.RotateDeg != 0 {
			θ := im.RotateDeg * math.Pi / 180
			cosθ := math.Cos(θ)
			sinθ := math.Sin(θ)
			cx := im.X + w/2
			cy := im.Y + h/2
			a = w * cosθ
			bb = -w * sinθ
			c = h * sinθ
			d = h * cosθ
			e = cx - w/2*cosθ - h/2*sinθ
			f = cy + w/2*sinθ - h/2*cosθ
		} else {
			a, bb, c, d, e, f = w, 0, 0, h, im.X, im.Y
		}
		ops = append(ops,
			content.Op{Name: "cm", Args: []model.Object{
				model.Real(a), model.Real(bb), model.Real(c), model.Real(d),
				model.Real(e), model.Real(f),
			}},
			content.Op{Name: "Do", Args: []model.Object{imName}},
			content.Op{Name: "Q", Args: nil},
		)
		if im.HasMCID {
			ops = append(ops, content.Op{Name: "EMC", Args: nil})
		}
	}
	return ops, xobj, imageGS
}

// buildPageContent returns content stream bytes and /Resources for graphics, text, and image runs.
// Draw order: graphics first (backgrounds/borders), then text, then images.
// When compression is enabled and effective, returns errFlateCompressed as the error.
func (b *DocumentBuilder) buildPageContent(page *pageSpec, pageHeight float64, imageStreamNums []int) ([]byte, model.Dict, error) {
	if len(page.GraphicRuns) == 0 && len(page.TextRuns) == 0 && len(page.ImageRuns) == 0 {
		return nil, nil, fmt.Errorf("no content")
	}

	ops := b.buildGraphicsOps(page, pageHeight)
	textOps, fontRes := b.buildTextOps(page, pageHeight, nil)
	ops = append(ops, textOps...)
	imageOps, xobj, imageGS := b.buildImageOps(page, pageHeight, imageStreamNums)
	ops = append(ops, imageOps...)

	contentBytes, err := content.EncodeBytes(ops)
	if err != nil {
		return nil, nil, err
	}

	resources := model.Dict{}
	if len(page.TextRuns) > 0 {
		fontDict := model.Dict{}
		hasStandardFont := false
		for base, resName := range fontRes {
			if base == "" {
				base = "Helvetica"
			}
			if _, isEmbedded := b.fc.embeddedFonts[base]; isEmbedded {
				// Placeholder: will be replaced with a Ref to the Type0 font in Build().
				fontDict[resName] = model.Dict{
					model.Name("_embedded"): model.Name(base),
				}
			} else {
				fontDict[resName] = model.Dict{
					model.Name("Type"):     model.Name("Font"),
					model.Name("Subtype"):  model.Name("Type1"),
					model.Name("BaseFont"): model.Name(base),
					model.Name("Encoding"): model.Name("WinAnsiEncoding"),
				}
				hasStandardFont = true
			}
		}
		if hasStandardFont {
			// Signal that standard fonts need a ToUnicode map.
			// We use a sentinel that buildPageObjects will replace with a real Ref.
			resources[model.Name("_standardToUnicode")] = model.Boolean(true)
		}
		if len(fontDict) > 0 {
			resources[model.Name("Font")] = fontDict
		}
	}
	if len(page.ImageRuns) > 0 && len(xobj) > 0 {
		resources[model.Name("XObject")] = xobj
	}
	allGS := model.Dict{}
	for _, gr := range page.GraphicRuns {
		for name, dict := range gr.ExtGStates {
			allGS[name] = dict
		}
	}
	for name, dict := range imageGS {
		allGS[name] = dict
	}
	if len(allGS) > 0 {
		resources[model.Name("ExtGState")] = allGS
	}

	return contentBytes, resources, nil
}

// sentinel error kept for compatibility (no longer used internally).
var errFlateCompressed = fmt.Errorf("flate compressed")

func (b *DocumentBuilder) imageXObjectStream(im imageRun) *model.Stream {
	dict := model.Dict{
		model.Name("Type"):             model.Name("XObject"),
		model.Name("Subtype"):          model.Name("Image"),
		model.Name("Width"):            model.Integer(int64(im.WidthPx)),
		model.Name("Height"):           model.Integer(int64(im.HeightPx)),
		model.Name("BitsPerComponent"): model.Integer(int64(im.BitsPerComponent)),
		model.Name("ColorSpace"):       model.Name(im.ColorSpace),
		model.Name("Length"):           model.Integer(int64(len(im.Raw))),
	}
	return &model.Stream{Dict: dict, Content: im.Raw}
}

func (b *DocumentBuilder) jpegXObjectStream(im imageRun) *model.Stream {
	dict := model.Dict{
		model.Name("Type"):             model.Name("XObject"),
		model.Name("Subtype"):          model.Name("Image"),
		model.Name("Width"):            model.Integer(int64(im.WidthPx)),
		model.Name("Height"):           model.Integer(int64(im.HeightPx)),
		model.Name("BitsPerComponent"): model.Integer(int64(im.BitsPerComponent)),
		model.Name("ColorSpace"):       model.Name(im.ColorSpace),
		model.Name("Filter"):           model.Name("DCTDecode"),
		model.Name("Length"):           model.Integer(int64(len(im.Raw))),
	}
	return &model.Stream{Dict: dict, Content: im.Raw}
}

func winAnsiEncode(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r < 128:
			out.WriteByte(byte(r))
		case r == '\u2022': // Bullet
			out.WriteByte(0x95)
		case r == '\u20ac': // Euro
			out.WriteByte(0x80)
		case r == '\u201a': // Single low-9 quotation mark
			out.WriteByte(0x82)
		case r == '\u0192': // Latin small letter f with hook
			out.WriteByte(0x83)
		case r == '\u201e': // Double low-9 quotation mark
			out.WriteByte(0x84)
		case r == '\u2026': // Horizontal ellipsis
			out.WriteByte(0x85)
		case r == '\u2020': // Dagger
			out.WriteByte(0x86)
		case r == '\u2021': // Double dagger
			out.WriteByte(0x87)
		case r == '\u02c6': // Modifier letter circumflex accent
			out.WriteByte(0x88)
		case r == '\u2030': // Per mille sign
			out.WriteByte(0x89)
		case r == '\u0160': // Latin capital letter S with caron
			out.WriteByte(0x8a)
		case r == '\u2039': // Single left-pointing angle quotation mark
			out.WriteByte(0x8b)
		case r == '\u0152': // Latin capital ligature OE
			out.WriteByte(0x8c)
		case r == '\u017d': // Latin capital letter Z with caron
			out.WriteByte(0x8e)
		case r == '\u2018': // Left single quotation mark
			out.WriteByte(0x91)
		case r == '\u2019': // Right single quotation mark
			out.WriteByte(0x92)
		case r == '\u201c': // Left double quotation mark
			out.WriteByte(0x93)
		case r == '\u201d': // Right double quotation mark
			out.WriteByte(0x94)
		case r == '\u2013': // En dash
			out.WriteByte(0x96)
		case r == '\u2014': // Em dash
			out.WriteByte(0x97)
		case r == '\u02dc': // Small tilde
			out.WriteByte(0x98)
		case r == '\u2122': // Trademark sign
			out.WriteByte(0x99)
		case r == '\u0161': // Latin small letter s with caron
			out.WriteByte(0x9a)
		case r == '\u203a': // Single right-pointing angle quotation mark
			out.WriteByte(0x9b)
		case r == '\u0153': // Latin small ligature oe
			out.WriteByte(0x9c)
		case r == '\u017e': // Latin small letter z with caron
			out.WriteByte(0x9e)
		case r == '\u0178': // Latin capital letter Y with diaeresis
			out.WriteByte(0x9f)
		case r >= 160 && r <= 255:
			out.WriteByte(byte(r))
		default:
			out.WriteRune('?')
		}
	}
	return out.String()
}

func buildToUnicodeWinAnsiCMap() []byte {
	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\n")
	b.WriteString("begincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	b.WriteString("/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n")
	b.WriteString("<00> <FF>\n")
	b.WriteString("endcodespacerange\n")

	special := map[byte]rune{
		0x80: '\u20ac', 0x82: '\u201a', 0x83: '\u0192', 0x84: '\u201e', 0x85: '\u2026',
		0x86: '\u2020', 0x87: '\u2021', 0x88: '\u02c6', 0x89: '\u2030', 0x8a: '\u0160',
		0x8b: '\u2039', 0x8c: '\u0152', 0x8e: '\u017d', 0x91: '\u2018', 0x92: '\u2019',
		0x93: '\u201c', 0x94: '\u201d', 0x95: '\u2022', 0x96: '\u2013', 0x97: '\u2014',
		0x98: '\u02dc', 0x99: '\u2122', 0x9a: '\u0161', 0x9b: '\u203a', 0x9c: '\u0153',
		0x9e: '\u017e', 0x9f: '\u0178',
	}

	fmt.Fprintf(&b, "%d beginbfchar\n", len(special))
	keys := slices.Collect(maps.Keys(special))
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "<%02X> <%04X>\n", k, special[k])
	}
	b.WriteString("endbfchar\n")

	b.WriteString("endcmap\n")
	b.WriteString("CMapName currentdict /CMap defineresource pop\n")
	b.WriteString("end\nend\n")
	return []byte(b.String())
}

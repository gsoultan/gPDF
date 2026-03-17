package doc

import (
	"bytes"
	"compress/zlib"
	"fmt"

	"gpdf/content"
	"gpdf/model"
)

// buildPageContent returns content stream bytes and /Resources for graphics, text, and image runs.
// Draw order: graphics first (backgrounds/borders), then text, then images.
// When compression is enabled and effective, returns errFlateCompressed as the error.
func (b *DocumentBuilder) buildPageContent(graphicRuns []graphicRun, textRuns []textRun, imageRuns []imageRun, imageStreamNums []int) ([]byte, model.Dict, error) {
	if len(graphicRuns) == 0 && len(textRuns) == 0 && len(imageRuns) == 0 {
		return nil, nil, fmt.Errorf("no content")
	}
	var ops []content.Op

	// 1. Graphics runs (backgrounds, lines, shapes).
	for _, gr := range graphicRuns {
		ops = append(ops, gr.ops...)
	}

	// 2. Text runs.
	fontRes := make(map[string]model.Name)
	currentColorSet := false
	var currentColor [3]float64
	for _, r := range textRuns {
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

		// Determine text operand: hex-encoded glyph IDs for embedded fonts, literal string otherwise.
		var textArg model.Object
		if eu, ok := b.embeddedFonts[baseName]; ok {
			eu.markText(r.Text)
			textArg = model.HexString(eu.font.Encode(r.Text))
		} else {
			textArg = model.String(r.Text)
		}

		ops = append(ops,
			content.Op{Name: "BT", Args: nil},
			content.Op{Name: "Tf", Args: []model.Object{resName, model.Real(size)}},
			content.Op{Name: "Td", Args: []model.Object{model.Real(r.X), model.Real(r.Y)}},
			content.Op{Name: "Tj", Args: []model.Object{textArg}},
			content.Op{Name: "ET", Args: nil},
		)
		if r.HasMCID {
			ops = append(ops, content.Op{Name: "EMC", Args: nil})
		}
	}

	// 3. Image runs.
	for i, im := range imageRuns {
		imName := model.Name("Im" + fmt.Sprintf("%d", i+1))
		w, h := im.WidthPt, im.HeightPt
		if w <= 0 {
			w = float64(im.WidthPx)
		}
		if h <= 0 {
			h = float64(im.HeightPx)
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
		ops = append(ops,
			content.Op{Name: "q", Args: nil},
			content.Op{Name: "cm", Args: []model.Object{
				model.Real(w), model.Real(0), model.Real(0), model.Real(h),
				model.Real(im.X), model.Real(im.Y),
			}},
			content.Op{Name: "Do", Args: []model.Object{imName}},
			content.Op{Name: "Q", Args: nil},
		)
		if im.HasMCID {
			ops = append(ops, content.Op{Name: "EMC", Args: nil})
		}
	}

	contentBytes, err := content.EncodeBytes(ops)
	if err != nil {
		return nil, nil, err
	}

	// FlateDecode the content stream for smaller output (unless disabled).
	compressed := false
	if !b.noCompressContent {
		contentBytes, compressed = flateCompress(contentBytes)
	}

	resources := model.Dict{}
	if len(textRuns) > 0 {
		fontDict := model.Dict{}
		for base, resName := range fontRes {
			if base == "" {
				base = "Helvetica"
			}
			if _, isEmbedded := b.embeddedFonts[base]; isEmbedded {
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
			}
		}
		if len(fontDict) > 0 {
			resources[model.Name("Font")] = fontDict
		}
	}
	if len(imageRuns) > 0 {
		xobj := make(model.Dict)
		for i, num := range imageStreamNums {
			name := model.Name("Im" + fmt.Sprintf("%d", i+1))
			xobj[name] = model.Ref{ObjectNumber: num, Generation: 0}
		}
		resources[model.Name("XObject")] = xobj
	}
	// Collect ExtGState entries from graphicRuns that use transparency/blending.
	allGS := model.Dict{}
	for _, gr := range graphicRuns {
		for name, dict := range gr.extGStates {
			allGS[name] = dict
		}
	}
	if len(allGS) > 0 {
		resources[model.Name("ExtGState")] = allGS
	}

	// Return content bytes with an indicator of whether FlateDecode was applied.
	// The caller uses this to set /Filter on the stream dict.
	if compressed {
		return contentBytes, resources, errFlateCompressed
	}
	return contentBytes, resources, nil
}

// sentinel error to signal the caller that content bytes are FlateDecode-compressed.
var errFlateCompressed = fmt.Errorf("flate compressed")

// flateCompress compresses data with zlib. Returns compressed bytes and true on success,
// or the original bytes and false if compression would increase size or fails.
func flateCompress(data []byte) ([]byte, bool) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		w.Close()
		return data, false
	}
	if err := w.Close(); err != nil {
		return data, false
	}
	if buf.Len() >= len(data) {
		return data, false
	}
	return buf.Bytes(), true
}

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

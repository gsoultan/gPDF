package reader

import (
	"strings"

	"gpdf/content"
	contentimpl "gpdf/content/impl"
	"gpdf/model"
)

// ExtractLayout returns one PageLayout per page, each containing positioned TextBlocks.
func ExtractLayout(src contentSource) ([]PageLayout, error) {
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	parser := contentimpl.NewStreamParser()
	layouts := make([]PageLayout, len(pages))
	for i, page := range pages {
		pl := PageLayout{Page: i}

		size := resolvePageSize(page)
		pl.Width = size.Width
		pl.Height = size.Height

		raw, err := pageContentBytes(src, page)
		if err != nil || len(raw) == 0 {
			layouts[i] = pl
			continue
		}
		ops, err := parser.Parse(raw)
		if err != nil {
			layouts[i] = pl
			continue
		}

		resources, _ := page.Resources()
		blocks := extractBlocksFromOps(
			ops, src, parser, resources,
			make(map[model.Ref]struct{}, 4),
			make(map[model.Ref]*toUnicodeDecoder, 4),
		)
		pl.Blocks = blocks
		layouts[i] = pl
	}
	return layouts, nil
}

// layoutState tracks the graphics/text state needed for positioned extraction.
type layoutState struct {
	// text matrix position
	tmX, tmY float64
	tmSet    bool
	// current text line matrix (for Td/TD/T*)
	tlX, tlY float64
	// active font info
	fontName string
	fontSize float64
	// non-stroking colour (fill) – RGB [0,1]
	colorR, colorG, colorB float64
	// leading (TL)
	leading float64
	// character spacing (Tc), word spacing (Tw), horizontal scale (Th)
	charSpacing float64
	wordSpacing float64
	horizScale  float64
	// graphics state stack
	stack []layoutStateSnapshot
}

type layoutStateSnapshot struct {
	colorR, colorG, colorB float64
}

func newLayoutState() *layoutState {
	return &layoutState{
		horizScale: 100,
		colorR:     0, colorG: 0, colorB: 0,
	}
}

func applyTextMatrixOp(state *layoutState, op content.Op, inText bool) {
	switch op.Name {
	case "Tm":
		if inText && len(op.Args) >= 6 {
			state.tmX = toFloat64(op.Args[4])
			state.tmY = toFloat64(op.Args[5])
			state.tlX = state.tmX
			state.tlY = state.tmY
			state.tmSet = true
			if scaleY := toFloat64(op.Args[3]); scaleY != 0 && state.fontSize == 0 {
				if scaleY < 0 {
					scaleY = -scaleY
				}
				state.fontSize = scaleY
			}
		}
	case "Td":
		if inText && len(op.Args) >= 2 {
			state.tlX += toFloat64(op.Args[0])
			state.tlY += toFloat64(op.Args[1])
			state.tmX, state.tmY = state.tlX, state.tlY
			state.tmSet = true
		}
	case "TD":
		if inText && len(op.Args) >= 2 {
			state.leading = -toFloat64(op.Args[1])
			state.tlX += toFloat64(op.Args[0])
			state.tlY += toFloat64(op.Args[1])
			state.tmX, state.tmY = state.tlX, state.tlY
			state.tmSet = true
		}
	case "T*":
		if inText {
			state.tlY -= state.leading
			state.tmX, state.tmY = state.tlX, state.tlY
			state.tmSet = true
		}
	}
}

func applyTextShowOp(
	state *layoutState,
	op content.Op,
	resources model.Dict,
	doc contentSource,
	fontDecoders map[model.Ref]*toUnicodeDecoder,
) []TextBlock {
	var blocks []TextBlock
	switch op.Name {
	case "Tj":
		if len(op.Args) > 0 {
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text := decodeOpArg(op.Args[0], decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
				state.tmX += estimateWidth(text, state)
			}
		}
	case "TJ":
		if len(op.Args) > 0 {
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text, advance := decodeTJArg(op.Args[0], decoder, state)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
				state.tmX += advance
			}
		}
	case "'":
		if len(op.Args) > 0 {
			state.tlY -= state.leading
			state.tmX, state.tmY = state.tlX, state.tlY
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text := decodeOpArg(op.Args[0], decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
				state.tmX += estimateWidth(text, state)
			}
		}
	case "\"":
		if len(op.Args) >= 3 {
			state.wordSpacing = toFloat64(op.Args[0])
			state.charSpacing = toFloat64(op.Args[1])
			state.tlY -= state.leading
			state.tmX, state.tmY = state.tlX, state.tlY
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text := decodeOpArg(op.Args[len(op.Args)-1], decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
				state.tmX += estimateWidth(text, state)
			}
		}
	}
	return blocks
}

func applyGraphicsStateOp(state *layoutState, op content.Op) {
	switch op.Name {
	case "cm":
		// Concat matrix — not tracked for layout extraction; no-op
	case "q":
		state.stack = append(state.stack, layoutStateSnapshot{
			colorR: state.colorR, colorG: state.colorG, colorB: state.colorB,
		})
	case "Q":
		if len(state.stack) > 0 {
			top := state.stack[len(state.stack)-1]
			state.stack = state.stack[:len(state.stack)-1]
			state.colorR = top.colorR
			state.colorG = top.colorG
			state.colorB = top.colorB
		}
	}
}

func applyFontOp(state *layoutState, op content.Op) {
	if op.Name != "Tf" || len(op.Args) < 2 {
		return
	}
	if n, ok := op.Args[0].(model.Name); ok {
		state.fontName = string(n)
	}
	state.fontSize = toFloat64(op.Args[1])
}

func applyColorOp(state *layoutState, op content.Op) {
	switch op.Name {
	case "rg":
		if len(op.Args) >= 3 {
			state.colorR = toFloat64(op.Args[0])
			state.colorG = toFloat64(op.Args[1])
			state.colorB = toFloat64(op.Args[2])
		}
	case "g":
		if len(op.Args) >= 1 {
			g := toFloat64(op.Args[0])
			state.colorR, state.colorG, state.colorB = g, g, g
		}
	case "RG":
		if len(op.Args) >= 3 {
			state.colorR = toFloat64(op.Args[0])
			state.colorG = toFloat64(op.Args[1])
			state.colorB = toFloat64(op.Args[2])
		}
	case "G":
		if len(op.Args) >= 1 {
			g := toFloat64(op.Args[0])
			state.colorR, state.colorG, state.colorB = g, g, g
		}
	case "k":
		if len(op.Args) >= 4 {
			c := toFloat64(op.Args[0])
			m := toFloat64(op.Args[1])
			y := toFloat64(op.Args[2])
			k := toFloat64(op.Args[3])
			state.colorR = (1 - c) * (1 - k)
			state.colorG = (1 - m) * (1 - k)
			state.colorB = (1 - y) * (1 - k)
		}
	case "K":
		if len(op.Args) >= 4 {
			c := toFloat64(op.Args[0])
			m := toFloat64(op.Args[1])
			y := toFloat64(op.Args[2])
			k := toFloat64(op.Args[3])
			state.colorR = (1 - c) * (1 - k)
			state.colorG = (1 - m) * (1 - k)
			state.colorB = (1 - y) * (1 - k)
		}
	}
}

func extractBlocksFromOps(
	ops []content.Op,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
) []TextBlock {
	st := newLayoutState()
	inText := false
	var blocks []TextBlock

	for _, op := range ops {
		switch op.Name {
		case "q", "Q", "cm":
			applyGraphicsStateOp(st, op)
		case "rg", "g", "RG", "G", "k", "K":
			applyColorOp(st, op)
		case "Tf":
			applyFontOp(st, op)
			_ = resolveFontDecoder(src, resources, op.Args[0], fontDecoders)
		case "Tm", "Td", "TD", "T*":
			applyTextMatrixOp(st, op, inText)
		case "Tj", "TJ", "'", "\"":
			if inText {
				blocks = append(blocks, applyTextShowOp(st, op, resources, src, fontDecoders)...)
			}
		case "BT":
			inText = true
			st.tmX, st.tmY = 0, 0
			st.tlX, st.tlY = 0, 0
			st.tmSet = false
		case "ET":
			inText = false
		case "TL":
			if len(op.Args) >= 1 {
				st.leading = toFloat64(op.Args[0])
			}
		case "Tc":
			if len(op.Args) >= 1 {
				st.charSpacing = toFloat64(op.Args[0])
			}
		case "Tw":
			if len(op.Args) >= 1 {
				st.wordSpacing = toFloat64(op.Args[0])
			}
		case "Tz":
			if len(op.Args) >= 1 {
				st.horizScale = toFloat64(op.Args[0])
			}
		case "sc", "scn":
			if len(op.Args) == 1 {
				g := toFloat64(op.Args[0])
				st.colorR, st.colorG, st.colorB = g, g, g
			} else if len(op.Args) >= 3 {
				st.colorR = toFloat64(op.Args[0])
				st.colorG = toFloat64(op.Args[1])
				st.colorB = toFloat64(op.Args[2])
			}
		case "Do":
			nested := extractBlocksFromXObject(op.Args, src, parser, resources, visited, fontDecoders)
			blocks = append(blocks, nested...)
		}
	}
	return blocks
}

func makeBlock(text string, st *layoutState) TextBlock {
	fs := st.fontSize
	if fs <= 0 {
		fs = 12
	}
	w := estimateWidth(text, st)
	return TextBlock{
		Text:   text,
		X:      st.tmX,
		Y:      st.tmY,
		Width:  w,
		Height: fs,
		Style: TextStyle{
			FontName: st.fontName,
			FontSize: fs,
			ColorR:   st.colorR,
			ColorG:   st.colorG,
			ColorB:   st.colorB,
		},
	}
}

func estimateWidth(text string, st *layoutState) float64 {
	fs := st.fontSize
	if fs <= 0 {
		fs = 12
	}
	return float64(len([]rune(text))) * fs * 0.6 * st.horizScale / 100
}

func decodeOpArg(arg model.Object, decoder *toUnicodeDecoder) string {
	switch v := arg.(type) {
	case model.String:
		return decodeTextBytes([]byte(v), decoder)
	}
	return ""
}

func decodeTJArg(arg model.Object, decoder *toUnicodeDecoder, st *layoutState) (string, float64) {
	arr, ok := arg.(model.Array)
	if !ok {
		return decodeOpArg(arg, decoder), 0
	}
	var sb strings.Builder
	var advance float64
	for _, elem := range arr {
		switch item := elem.(type) {
		case model.String:
			text := decodeTextBytes([]byte(item), decoder)
			sb.WriteString(text)
			advance += estimateWidth(text, st)
		case model.Integer:
			// negative kerning adjustment → word gap if large enough
			if item <= -120 {
				sb.WriteByte(' ')
				advance += st.fontSize * 0.3
			} else {
				// positive or small negative: adjust advance
				advance -= float64(item) * st.fontSize / 1000
			}
		case model.Real:
			if item <= -120 {
				sb.WriteByte(' ')
				advance += st.fontSize * 0.3
			} else {
				advance -= float64(item) * st.fontSize / 1000
			}
		}
	}
	return sb.String(), advance
}

func extractBlocksFromXObject(
	args model.Array,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
) []TextBlock {
	if len(args) == 0 || resources == nil {
		return nil
	}
	name, ok := args[0].(model.Name)
	if !ok {
		return nil
	}
	xObjects, ok := resources[model.Name("XObject")].(model.Dict)
	if !ok {
		return nil
	}
	xoObj, ok := xObjects[name]
	if !ok {
		return nil
	}
	xObject, ref, ok := resolveStreamObject(src, xoObj)
	if !ok || len(xObject.Content) == 0 {
		return nil
	}
	subtype, _ := xObject.Dict[model.Name("Subtype")].(model.Name)
	if subtype != "Form" {
		return nil
	}
	if ref != nil {
		if _, seen := visited[*ref]; seen {
			return nil
		}
		visited[*ref] = struct{}{}
		defer delete(visited, *ref)
	}
	nestedOps, err := parser.Parse(xObject.Content)
	if err != nil {
		return nil
	}
	nestedResources := resources
	if r, ok := resolveDictObject(src, xObject.Dict[model.Name("Resources")]); ok {
		nestedResources = mergeResourceDict(resources, r)
	}
	return extractBlocksFromOps(nestedOps, src, parser, nestedResources, visited, fontDecoders)
}

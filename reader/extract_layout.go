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
			nil, // Initial state
		)
		pl.Blocks = MergeBlocks(blocks)
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
	fontName     string // Resource name (e.g., F18) for decoder lookup
	resolvedFont string // BaseFont name (e.g., Helvetica) for output
	fontSize     float64
	// resolved metrics for the active font
	fontMetrics fontInfo
	// last computed text advance (in text space units, before CTM)
	lastAdvance float64
	// non-stroking colour (fill) – RGB [0,1]
	colorR, colorG, colorB float64
	// leading (TL)
	leading float64
	// character spacing (Tc), word spacing (Tw), horizontal scale (Th)
	charSpacing float64
	wordSpacing float64
	horizScale  float64
	// text rise (Ts) and render mode (Tr)
	textRise   float64
	textRender int
	// graphics state stack
	ctm   matrix
	stack []layoutStateSnapshot
}

type layoutStateSnapshot struct {
	ctm                    matrix
	colorR, colorG, colorB float64
}

func newLayoutState() *layoutState {
	return &layoutState{
		horizScale: 100,
		colorR:     0, colorG: 0, colorB: 0,
		ctm: identityMatrix(),
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
			// decode display text for output
			text := decodeOpArg(op.Args[0], decoder)
			// compute precise advance based on font metrics and spacing
			state.lastAdvance = advanceForTextObject(op.Args[0], state, decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
			}
			state.tmX += state.lastAdvance
			state.lastAdvance = 0
		}
	case "TJ":
		if len(op.Args) > 0 {
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text, advance := decodeTJArg(op.Args[0], decoder, state)
			state.lastAdvance = advance
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
			}
			state.tmX += state.lastAdvance
			state.lastAdvance = 0
		}
	case "'":
		if len(op.Args) > 0 {
			state.tlY -= state.leading
			state.tmX, state.tmY = state.tlX, state.tlY
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text := decodeOpArg(op.Args[0], decoder)
			state.lastAdvance = advanceForTextObject(op.Args[0], state, decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
			}
			state.tmX += state.lastAdvance
			state.lastAdvance = 0
		}
	case "\"":
		if len(op.Args) >= 3 {
			state.wordSpacing = toFloat64(op.Args[0])
			state.charSpacing = toFloat64(op.Args[1])
			state.tlY -= state.leading
			state.tmX, state.tmY = state.tlX, state.tlY
			decoder := resolveFontDecoder(doc, resources, model.Name(state.fontName), fontDecoders)
			text := decodeOpArg(op.Args[len(op.Args)-1], decoder)
			state.lastAdvance = advanceForTextObject(op.Args[len(op.Args)-1], state, decoder)
			if text != "" {
				blocks = append(blocks, makeBlock(text, state))
			}
			state.tmX += state.lastAdvance
			state.lastAdvance = 0
		}
	}
	return blocks
}

func applyGraphicsStateOp(state *layoutState, op content.Op) {
	switch op.Name {
	case "cm":
		if len(op.Args) >= 6 {
			state.ctm = matrixFromObjects(op.Args).multiply(state.ctm)
		}
	case "q":
		state.stack = append(state.stack, layoutStateSnapshot{
			ctm:    state.ctm,
			colorR: state.colorR, colorG: state.colorG, colorB: state.colorB,
		})
	case "Q":
		if len(state.stack) > 0 {
			top := state.stack[len(state.stack)-1]
			state.stack = state.stack[:len(state.stack)-1]
			state.ctm = top.ctm
			state.colorR = top.colorR
			state.colorG = top.colorG
			state.colorB = top.colorB
		}
	}
}

func applyFontOp(state *layoutState, op content.Op, resources model.Dict, src contentSource) {
	if op.Name != "Tf" || len(op.Args) < 2 {
		return
	}
	if n, ok := op.Args[0].(model.Name); ok {
		state.fontName = string(n)
	}
	state.fontSize = toFloat64(op.Args[1])
	// try to resolve precise font metrics for accurate advance calculations
	state.fontMetrics = resolveFontInfo(src, resources, op.Args[0])
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
	st *layoutState,
) []TextBlock {
	if st == nil {
		st = newLayoutState()
	}
	inText := false
	var blocks []TextBlock

	for _, op := range ops {
		switch op.Name {
		case "q", "Q", "cm":
			applyGraphicsStateOp(st, op)
		case "rg", "g", "RG", "G", "k", "K":
			applyColorOp(st, op)
		case "Tf":
			applyFontOp(st, op, resources, src)
			_ = resolveFontDecoder(src, resources, op.Args[0], fontDecoders)
			st.resolvedFont = resolveFontName(src, resources, op.Args[0])
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
		case "Ts":
			if len(op.Args) >= 1 {
				st.textRise = toFloat64(op.Args[0])
			}
		case "Tr":
			if len(op.Args) >= 1 {
				// Tr integer render mode per PDF spec (0 fill, 1 stroke, 2 fill+stroke, 3 invisible, etc.)
				// We record it for potential downstream use; extraction still captures text content.
				switch v := op.Args[0].(type) {
				case model.Integer:
					st.textRender = int(v)
				case model.Real:
					if v >= 0 {
						st.textRender = int(v)
					} else {
						st.textRender = 0
					}
				}
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
			nested := extractBlocksFromXObject(op.Args, src, parser, resources, visited, fontDecoders, st)
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
	// Prefer the precise width computed during text-show op, if available
	w := st.lastAdvance
	if w <= 0 {
		w = estimateWidth(text, st)
	}
	font := st.resolvedFont
	if font == "" {
		font = st.fontName
	}

	// Apply text rise before CTM: baseline shifted by Ts in text space
	yWithRise := st.tmY + st.textRise
	// Apply CTM to get page-space coordinates
	px, py := st.ctm.apply(st.tmX, yWithRise)

	return TextBlock{
		Text:   text,
		X:      px,
		Y:      py,
		Width:  w * st.ctm.scaling(),
		Height: fs * st.ctm.scaling(),
		Style: TextStyle{
			FontName:    st.fontName,
			BaseFont:    font,
			FontSize:    fs * st.ctm.scaling(),
			CharSpacing: st.charSpacing,
			WordSpacing: st.wordSpacing,
			HorizontalScale: func() float64 {
				if st.horizScale == 0 {
					return 100
				}
				return st.horizScale
			}(),
			Leading: st.leading,
			ColorR:  st.colorR,
			ColorG:  st.colorG,
			ColorB:  st.colorB,
		},
	}
}

func estimateWidth(text string, st *layoutState) float64 {
	// If we have resolved metrics, compute width from glyph widths.
	if st != nil && len(st.fontMetrics.Widths) > 0 && st.fontSize > 0 {
		// Best-effort: use rune-by-rune and treat space specially for Tw application.
		var units float64
		fs := st.fontSize
		for _, r := range text {
			code := int(r)
			// PDF word spacing Tw applies only to the SPACE character in simple fonts.
			if r == ' ' {
				units += st.wordSpacing
			}
			w := st.fontMetrics.Widths[code]
			if w == 0 {
				w = st.fontMetrics.DefaultWidth
			}
			// Convert font units (1/1000 em) to text space using font size and add char spacing
			units += (w/1000.0)*fs + st.charSpacing
		}
		return units * st.horizScale / 100
	}
	// Fallback heuristic
	fs := st.fontSize
	if fs <= 0 {
		fs = 12
	}
	return float64(len([]rune(text))) * fs * 0.5 * st.horizScale / 100
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
			b := []byte(item)
			text := decodeTextBytes(b, decoder)
			sb.WriteString(text)
			advance += advanceForBytes(b, st, decoder)
		case model.Integer:
			// PDF spec: a number adjusts the text position by -value/1000 * fontSize
			advance += (-float64(item)) * st.fontSize / 1000
		case model.Real:
			advance += (-float64(item)) * st.fontSize / 1000
		}
	}
	// Apply horizontal scale
	return sb.String(), advance * st.horizScale / 100
}

func extractBlocksFromXObject(
	args model.Array,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
	st *layoutState,
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

	// Create child state with inherited graphics/color state
	childState := &layoutState{
		ctm:         st.ctm,
		colorR:      st.colorR,
		colorG:      st.colorG,
		colorB:      st.colorB,
		horizScale:  100,
		fontSize:    st.fontSize,
		fontName:    st.fontName,
		wordSpacing: st.wordSpacing,
		charSpacing: st.charSpacing,
		leading:     st.leading,
	}

	if formMatrixArray, ok := xObject.Dict[model.Name("Matrix")].(model.Array); ok && len(formMatrixArray) >= 6 {
		formMatrix := matrixFromObjects(formMatrixArray)
		childState.ctm = formMatrix.multiply(childState.ctm)
	}

	return extractBlocksFromOps(nestedOps, src, parser, nestedResources, visited, fontDecoders, childState)
}

// advanceForTextObject computes precise advance for a text-show operand (model.String),
// taking into account font metrics and spacing. If metrics are unavailable, falls back
// to heuristic estimateWidth.
func advanceForTextObject(obj model.Object, st *layoutState, decoder *toUnicodeDecoder) float64 {
	if s, ok := obj.(model.String); ok {
		return advanceForBytes([]byte(s), st, decoder)
	}
	// Fallback when operand is not a string (should not happen for Tj/'/")
	return estimateWidth(decodeOpArg(obj, decoder), st)
}

// advanceForBytes computes the text-space advance for a raw PDF string using
// the active font metrics, char/word spacing, and horizontal scaling.
func advanceForBytes(b []byte, st *layoutState, decoder *toUnicodeDecoder) float64 {
	if st == nil {
		return 0
	}
	fs := st.fontSize
	if fs <= 0 {
		fs = 12
	}
	var adv float64
	if len(st.fontMetrics.Widths) > 0 {
		// Best-effort: iterate bytes, apply widths in 1/1000 em units.
		for _, by := range b {
			// Word spacing Tw applies to SPACE (0x20) for simple fonts.
			if by == 0x20 {
				adv += st.wordSpacing
			}
			w := st.fontMetrics.Widths[int(by)]
			if w == 0 {
				w = st.fontMetrics.DefaultWidth
			}
			adv += (w/1000.0)*fs + st.charSpacing
		}
		return adv * st.horizScale / 100
	}
	// Fallback heuristic when no metrics known
	// Decode to runes to approximate spacing
	text := ""
	if decoder != nil {
		text = decodeTextBytes(b, decoder)
	} else {
		text = string(b)
	}
	return estimateWidth(text, st)
}

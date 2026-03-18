package doc

import (
	"fmt"

	taggedpkg "gpdf/doc/tagged"
	"gpdf/doc/text"
	"gpdf/model"
)

// DrawTextBox lays out text within a horizontal box on the given page.
// Text is wrapped to fit within opts.Width using a simple width heuristic based on fontSize.
// Alignment, line height, and optional multi-page continuation are controlled via opts.
// When opts.Width <= 0 or pageIndex is out of range, the call is a no-op.
func (b *DocumentBuilder) DrawTextBox(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder {
	if textStr == "" || opts.Width <= 0 {
		return b
	}
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}

	b.layoutTextIntoPages(pageIndex, textStr, x, y, fontName, fontSize, opts, false, model.Name(""))
	return b
}

// DrawTaggedParagraphBox behaves like DrawTextBox but creates a tagged paragraph (/P) in the structure tree.
// The resulting text runs participate in Tagged PDF reading order.
func (b *DocumentBuilder) DrawTaggedParagraphBox(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder {
	if textStr == "" || opts.Width <= 0 {
		return b
	}
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}

	b.layoutTextIntoPages(pageIndex, textStr, x, y, fontName, fontSize, opts, true, model.Name("P"))
	return b
}

// layoutTextIntoPages performs simple line layout for paragraph-like text and appends textRuns.
// For tagged text, layout is restricted to a single page; AllowPageBreak is ignored in that case.
func (b *DocumentBuilder) layoutTextIntoPages(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions, isTagged bool, role model.Name) {
	if !b.pc.validPageIndex(pageIndex) {
		return
	}
	ps := &b.pc.pages[pageIndex]

	lines := b.wrapTextLines(textStr, fontSize, opts.Width, fontName)
	if len(lines) == 0 {
		return
	}

	height := b.pageHeight(pageIndex)
	if height <= 0 {
		height = 842
	}
	const marginBottom = 40.0

	currentPage := pageIndex
	currentY := y
	var mcid int
	if isTagged {
		mcid = ps.NextMCID
		ps.NextMCID++
		b.tagging.Blocks = append(b.tagging.Blocks, taggedpkg.Block{
			PageIndex: pageIndex,
			MCID:      mcid,
			Role:      role,
		})
		b.tagging.RecordSectionBlock(len(b.tagging.Blocks) - 1)
	}

	for i, line := range lines {
		if isTagged && currentPage != pageIndex {
			break
		}
		if !isTagged && opts.AllowPageBreak && currentY < marginBottom {
			if currentPage+1 >= len(b.pc.pages) {
				break
			}
			currentPage++
			currentY = b.pageHeight(currentPage) - marginBottom
		}

		lineWidth := b.textWidth(line, fontSize, fontName)
		offsetX := x
		wordSpacing := 0.0
		free := opts.Width - lineWidth
		// A line is the last in its paragraph when it is the final line overall
		// or when the next line is empty (paragraph break emitted by WrapLines).
		isLastInParagraph := i == len(lines)-1 || lines[i+1] == ""
		switch opts.Align {
		case TextAlignCenter:
			if free > 0 {
				offsetX = x + free/2
			}
		case TextAlignRight:
			if free > 0 {
				offsetX = x + free
			}
		case TextAlignJustify:
			if !isLastInParagraph && free > 0 {
				numSpaces := countWordSpaces(line)
				if numSpaces > 0 {
					wordSpacing = free / float64(numSpaces)
				}
			}
		}

		targetPage := &b.pc.pages[currentPage]
		r := textRun{
			Text:        line,
			X:           offsetX,
			Y:           currentY,
			FontName:    fontName,
			FontSize:    fontSize,
			WordSpacing: wordSpacing,
		}
		if isTagged {
			r.MCID = mcid
			r.HasMCID = true
			r.UseDefaultColor = true
		}
		targetPage.TextRuns = append(targetPage.TextRuns, r)
		currentY -= opts.LineHeight
	}

	if opts.ParagraphSpacing > 0 {
		currentY -= opts.ParagraphSpacing
	}
}

// countWordSpaces returns the number of inter-word spaces in s (i.e. space characters).
func countWordSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		}
	}
	return count
}

// wrapTextLines splits text into lines that fit within width using font metrics or a heuristic.
func (b *DocumentBuilder) wrapTextLines(s string, fontSize, width float64, fontName string) []string {
	widthFn := b.fontWidthFunc(fontName)
	return text.WrapLines(s, fontSize, width, widthFn)
}

// fontWidthFunc returns a FontWidthFunc that uses registered font metrics or the fallback heuristic.
func (b *DocumentBuilder) fontWidthFunc(fontName string) text.FontWidthFunc {
	if f, ok := b.fc.fonts[fontName]; ok {
		return func(s string, fontSize float64) float64 {
			return f.TextWidth(s, fontSize)
		}
	}
	return func(s string, fontSize float64) float64 {
		return text.ApproxWidth(s, fontSize)
	}
}

// textWidth returns the width of text in points.
// Uses real glyph widths from a registered font when available, otherwise falls back to a heuristic.
func (b *DocumentBuilder) textWidth(s string, fontSize float64, fontName string) float64 {
	if s == "" || fontSize <= 0 {
		return 0
	}
	if f, ok := b.fc.fonts[fontName]; ok {
		return f.TextWidth(s, fontSize)
	}
	return text.ApproxWidth(s, fontSize)
}

// DrawText queues text to be drawn on the last added page at (x, y) using the given font and size.
// FontName should be a standard PDF base font (e.g. Helvetica, Times-Roman). Call after AddPage().
func (b *DocumentBuilder) DrawText(textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
	})
	return b
}

// DrawTextColored queues text drawn in the specified RGB color on the last added page.
// It behaves like DrawText but sets the fill color for that run.
func (b *DocumentBuilder) DrawTextColored(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB: [3]float64{color.R, color.G, color.B},
	})
	return b
}

// DrawTextCentered draws text horizontally centered around the point (cx, y).
// The text baseline is at y; cx is the horizontal center of the rendered text.
func (b *DocumentBuilder) DrawTextCentered(textStr string, cx, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	w := b.textWidth(textStr, fontSize, fontName)
	return b.DrawText(textStr, cx-w/2, y, fontName, fontSize)
}

// DrawTextCenteredColored draws colored text horizontally centered around the point (cx, y).
func (b *DocumentBuilder) DrawTextCenteredColored(textStr string, cx, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	w := b.textWidth(textStr, fontSize, fontName)
	return b.DrawTextColored(textStr, cx-w/2, y, fontName, fontSize, color)
}

// DrawTextRight draws text right-aligned so that its right edge is at x.
func (b *DocumentBuilder) DrawTextRight(textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	w := b.textWidth(textStr, fontSize, fontName)
	return b.DrawText(textStr, x-w, y, fontName, fontSize)
}

// DrawTextRightColored draws colored text right-aligned so that its right edge is at x.
func (b *DocumentBuilder) DrawTextRightColored(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	w := b.textWidth(textStr, fontSize, fontName)
	return b.DrawTextColored(textStr, x-w, y, fontName, fontSize, color)
}

// DrawTextWithUnderline draws text with an underline on the last added page.
func (b *DocumentBuilder) DrawTextWithUnderline(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB: [3]float64{color.R, color.G, color.B},
		Underline:    true,
	})
	return b
}

// DrawTextWithStrikethrough draws text with a strikethrough line on the last added page.
func (b *DocumentBuilder) DrawTextWithStrikethrough(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB:  [3]float64{color.R, color.G, color.B},
		Strikethrough: true,
	})
	return b
}

// DrawTextBoxColored lays out wrapped text like DrawTextBox but renders each line in the given color.
func (b *DocumentBuilder) DrawTextBoxColored(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions, color Color) *DocumentBuilder {
	if textStr == "" || opts.Width <= 0 {
		return b
	}
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	lines := b.wrapTextLines(textStr, fontSize, opts.Width, fontName)
	curY := y
	for i, line := range lines {
		lineWidth := b.textWidth(line, fontSize, fontName)
		offsetX := x
		wordSpacing := 0.0
		free := opts.Width - lineWidth
		isLastInParagraph := i == len(lines)-1 || lines[i+1] == ""
		switch opts.Align {
		case TextAlignCenter:
			if free > 0 {
				offsetX = x + free/2
			}
		case TextAlignRight:
			if free > 0 {
				offsetX = x + free
			}
		case TextAlignJustify:
			if !isLastInParagraph && free > 0 {
				numSpaces := countWordSpaces(line)
				if numSpaces > 0 {
					wordSpacing = free / float64(numSpaces)
				}
			}
		}
		b.pc.pages[pageIndex].TextRuns = append(b.pc.pages[pageIndex].TextRuns, textRun{
			Text:          line,
			X:             offsetX,
			Y:             curY,
			FontName:      fontName,
			FontSize:      fontSize,
			TextColorRGB:  [3]float64{color.R, color.G, color.B},
			LetterSpacing: opts.LetterSpacing,
			WordSpacing:   wordSpacing,
		})
		curY -= opts.LineHeight
	}
	return b
}

// DrawParagraph queues a tagged paragraph (/P) on the given page at (x, y).
// The text is associated with a structure element so it participates in Tagged PDF reading order.
func (b *DocumentBuilder) DrawParagraph(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if textStr == "" || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	ps.TextRuns = append(ps.TextRuns, textRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	b.tagging.Blocks = append(b.tagging.Blocks, taggedpkg.Block{
		PageIndex: pageIndex,
		MCID:      mcid,
		Role:      model.Name("P"),
	})
	b.tagging.RecordSectionBlock(len(b.tagging.Blocks) - 1)
	return b
}

// DrawTaggedCaption queues a tagged caption (/Caption) — e.g. for figure or table captions.
func (b *DocumentBuilder) DrawTaggedCaption(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.drawTaggedBlock(pageIndex, textStr, x, y, fontName, fontSize, model.Name("Caption"))
}

// DrawTaggedQuote queues a single-line tagged block quote (/Quote).
func (b *DocumentBuilder) DrawTaggedQuote(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.drawTaggedBlock(pageIndex, textStr, x, y, fontName, fontSize, model.Name("Quote"))
}

// DrawTaggedCode queues a single-line tagged code (/Code) — e.g. inline or one-line code.
func (b *DocumentBuilder) DrawTaggedCode(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.drawTaggedBlock(pageIndex, textStr, x, y, fontName, fontSize, model.Name("Code"))
}

// drawTaggedBlock adds a single-line tagged text block with the given structure role.
func (b *DocumentBuilder) drawTaggedBlock(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, role model.Name) *DocumentBuilder {
	if textStr == "" || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	ps.TextRuns = append(ps.TextRuns, textRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	b.tagging.Blocks = append(b.tagging.Blocks, taggedpkg.Block{
		PageIndex: pageIndex,
		MCID:      mcid,
		Role:      role,
	})
	b.tagging.RecordSectionBlock(len(b.tagging.Blocks) - 1)
	return b
}

// DrawTaggedQuoteBox lays out wrapped text as a tagged block quote (/Quote).
func (b *DocumentBuilder) DrawTaggedQuoteBox(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder {
	if textStr == "" || opts.Width <= 0 || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	b.layoutTextIntoPages(pageIndex, textStr, x, y, fontName, fontSize, opts, true, model.Name("Quote"))
	return b
}

// DrawTaggedCodeBlock lays out wrapped text as a tagged code block (/Code).
func (b *DocumentBuilder) DrawTaggedCodeBlock(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder {
	if textStr == "" || opts.Width <= 0 || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 10
	}
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	b.layoutTextIntoPages(pageIndex, textStr, x, y, fontName, fontSize, opts, true, model.Name("Code"))
	return b
}

// DrawHeading queues a tagged heading (/H1..H6) on the given page at (x, y).
// Level must be in [1,6]; values outside this range are clamped.
func (b *DocumentBuilder) DrawHeading(pageIndex int, level int, textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	if textStr == "" || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	if fontName == "" {
		fontName = "Helvetica-Bold"
	}
	if fontSize <= 0 {
		base := 12.0
		switch level {
		case 1:
			fontSize = base * 2
		case 2:
			fontSize = base * 1.6
		case 3:
			fontSize = base * 1.3
		default:
			fontSize = base * 1.1
		}
	}
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	ps.TextRuns = append(ps.TextRuns, textRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	role := model.Name(fmt.Sprintf("H%d", level))
	b.tagging.Blocks = append(b.tagging.Blocks, taggedpkg.Block{
		PageIndex: pageIndex,
		MCID:      mcid,
		Role:      role,
	})
	b.tagging.RecordSectionBlock(len(b.tagging.Blocks) - 1)
	return b
}

// DrawList queues a tagged list (/L with /LI children) on the given page.
// Items are rendered vertically starting at (x, y) with the given lineHeight (or a default when <= 0).
// When ordered is true, items are prefixed with "1. ", "2. ", ...; otherwise a bullet "• " is used.
func (b *DocumentBuilder) DrawList(pageIndex int, items []string, x, y, lineHeight float64, ordered bool, fontName string, fontSize float64) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(items) == 0 {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if lineHeight <= 0 {
		lineHeight = fontSize * 1.2
	}
	ps := &b.pc.pages[pageIndex]
	var listItems []taggedpkg.ListItem
	for idx, raw := range items {
		if raw == "" {
			continue
		}
		label := "• "
		if ordered {
			label = fmt.Sprintf("%d. ", idx+1)
		}
		itemText := label + raw
		itemY := y - float64(len(listItems))*lineHeight
		mcid := ps.NextMCID
		ps.NextMCID++
		ps.TextRuns = append(ps.TextRuns, textRun{
			Text:            itemText,
			X:               x,
			Y:               itemY,
			FontName:        fontName,
			FontSize:        fontSize,
			MCID:            mcid,
			HasMCID:         true,
			UseDefaultColor: true,
		})
		listItems = append(listItems, taggedpkg.ListItem{MCID: mcid})
	}
	if len(listItems) == 0 {
		return b
	}
	b.tagging.Lists = append(b.tagging.Lists, taggedpkg.List{
		PageIndex: pageIndex,
		Ordered:   ordered,
		Items:     listItems,
	})
	b.tagging.RecordSectionList(len(b.tagging.Lists) - 1)
	return b
}

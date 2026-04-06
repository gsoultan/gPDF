package doc

import (
	"fmt"
	"strings"

	"github.com/gsoultan/gpdf/doc/style"
	taggedpkg "github.com/gsoultan/gpdf/doc/tagged"
	"github.com/gsoultan/gpdf/doc/text"
	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/model"
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
func (b *DocumentBuilder) layoutTextIntoPages(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions, isTagged bool, role model.Name) (int, float64) {
	if !b.pc.validPageIndex(pageIndex) {
		return pageIndex, y
	}
	ps := &b.pc.pages[pageIndex]

	var lines []string
	if opts.LineRectFn != nil {
		lines = b.wrapTextLinesRect(textStr, fontSize, fontName, opts.LineRectFn)
	} else {
		lines = b.wrapTextLines(textStr, fontSize, opts.Width, fontName)
	}
	if len(lines) == 0 {
		return pageIndex, y
	}

	height := b.pageHeight(pageIndex)
	if height <= 0 {
		height = 842
	}
	const marginBottom = 20.0

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
			// If we need a new page and none exist, AddPage.
			if currentPage+1 >= len(b.pc.pages) {
				b.AddPage()
			}
			currentPage++
			currentY = b.pageHeight(currentPage) - marginBottom
		}

		lineOffsetX := 0.0
		lineWidthTarget := opts.Width
		if opts.LineRectFn != nil {
			lineOffsetX, lineWidthTarget = opts.LineRectFn(i)
		}

		lineWidth := b.textWidthStyle(line, fontSize, fontName, opts.LetterSpacing, 0)
		offsetX := x + lineOffsetX
		wordSpacing := 0.0
		free := lineWidthTarget - lineWidth
		// A line is the last in its paragraph when it is the final line overall
		// or when the next line is empty (paragraph break emitted by WrapLines).
		isLastInParagraph := i == len(lines)-1 || lines[i+1] == ""
		switch opts.Align {
		case TextAlignCenter:
			if free > 0 {
				offsetX = x + lineOffsetX + free/2
			}
		case TextAlignRight:
			if free > 0 {
				offsetX = x + lineOffsetX + free
			}
		case TextAlignJustify:
			if !isLastInParagraph && free > 0 {
				numSpaces := countWordSpaces(line)
				if numSpaces > 0 {
					wordSpacing = free / float64(numSpaces)
					// Recalculate width with the new word spacing for exact alignment
					lineWidth = b.textWidthStyle(line, fontSize, fontName, opts.LetterSpacing, wordSpacing)
					free = lineWidthTarget - lineWidth
				}
			}
		}

		targetPage := &b.pc.pages[currentPage]
		segments := b.fc.resolveFont(line, fontName)
		currentX := offsetX
		for _, seg := range segments {
			if opts.IsVertical {
				for _, r := range seg.text {
					charStr := string(r)
					b.pc.pages[currentPage].TextRuns = append(b.pc.pages[currentPage].TextRuns, textRun{
						Text:     charStr,
						X:        currentX,
						Y:        currentY,
						FontName: seg.fontName,
						FontSize: fontSize,
					})
					currentY -= fontSize * 1.0 // Simple vertical spacing
				}
				continue
			}

			textToDraw := seg.text
			if seg.isRTL {
				textToDraw = reverseRunes(textToDraw)
			}
			segWidth := b.textWidthStyle(seg.text, fontSize, seg.fontName, opts.LetterSpacing, 0)
			numSpaces := countWordSpaces(seg.text)

			r := textRun{
				Text:            textToDraw,
				X:               currentX,
				Y:               currentY,
				FontName:        seg.fontName,
				FontSize:        fontSize,
				WordSpacing:     wordSpacing,
				LetterSpacing:   opts.LetterSpacing,
				SyntheticBold:   opts.SyntheticBold,
				SyntheticItalic: opts.SyntheticItalic,
			}
			if opts.HasColor {
				r.TextColorRGB = [3]float64{opts.Color.R, opts.Color.G, opts.Color.B}
				r.UseDefaultColor = false
			}
			if isTagged {
				r.MCID = mcid
				r.HasMCID = true
				if !opts.HasColor {
					r.UseDefaultColor = true
				}
			}
			targetPage.TextRuns = append(targetPage.TextRuns, r)
			currentX += segWidth + float64(numSpaces)*wordSpacing
		}
		currentY -= opts.LineHeight
	}

	if opts.ParagraphSpacing > 0 {
		currentY -= opts.ParagraphSpacing
	}
	return currentPage, currentY
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

func (b *DocumentBuilder) wrapTextLinesRect(s string, fontSize float64, fontName string, lineRectFn text.LineRectFunc) []string {
	widthFn := b.fontWidthFunc(fontName)
	return text.WrapLinesRect(s, fontSize, widthFn, lineRectFn)
}

// wrapTextLinesDynamic splits text into lines that can have different widths.
func (b *DocumentBuilder) wrapTextLinesDynamic(s string, fontSize float64, fontName string, lineWidthFn text.LineWidthFunc) []string {
	widthFn := b.fontWidthFunc(fontName)
	return text.WrapLinesDynamic(s, fontSize, widthFn, lineWidthFn)
}

// fontWidthFunc returns a FontWidthFunc that uses registered font metrics or the fallback heuristic.
func (b *DocumentBuilder) fontWidthFunc(fontName string) text.FontWidthFunc {
	return func(s string, fontSize float64) float64 {
		return b.textWidthStyle(s, fontSize, fontName, 0, 0)
	}
}

// textWidth returns the width of text in points.
// Uses real glyph widths from a registered font when available, otherwise falls back to a heuristic.
func (b *DocumentBuilder) textWidth(s string, fontSize float64, fontName string) float64 {
	return b.textWidthStyle(s, fontSize, fontName, 0, 0)
}

func (b *DocumentBuilder) textWidthStyle(s string, fontSize float64, fontName string, charSpacing, wordSpacing float64) float64 {
	if s == "" || fontSize <= 0 {
		return 0
	}
	var w float64

	// Optimization: if it's a known font and no fallbacks would trigger, avoid resolveFont.
	// We can check if the string contains any non-ASCII characters that might trigger fallback.
	// But to be safe and simple, let's just check if we have any fallbacks at all.
	if len(b.fc.fallbackChain) == 0 && len(b.fc.blockFallbacks) == 0 {
		if f, ok := b.fc.fonts[fontName]; ok {
			w = f.TextWidth(s, fontSize)
		} else {
			w = b.standardOrApproxWidth(s, fontSize, fontName)
		}
	} else {
		segments := b.fc.resolveFont(s, fontName)
		for _, seg := range segments {
			if f, ok := b.fc.fonts[seg.fontName]; ok {
				w += f.TextWidth(seg.text, fontSize)
			} else {
				w += b.standardOrApproxWidth(seg.text, fontSize, seg.fontName)
			}
		}
	}

	if charSpacing != 0 {
		w += float64(len([]rune(s))-1) * charSpacing
	}
	if wordSpacing != 0 {
		w += float64(countWordSpaces(s)) * wordSpacing
	}
	return w
}

func (b *DocumentBuilder) standardOrApproxWidth(s string, fontSize float64, fontName string) float64 {
	var w int
	allFound := true
	for _, r := range s {
		width := font.GetStandardWidth(fontName, r)
		if width == 0 && r != ' ' { // space might be 0 in some fonts but we checked GetStandardWidth
			allFound = false
			break
		}
		w += width
	}

	if allFound && len(s) > 0 {
		return float64(w) * fontSize / 1000.0
	}
	return text.ApproxWidth(s, fontSize)
}

// DrawRichText renders styled rich text on the given page starting at (x, y).
// It automatically calculates the horizontal position for each segment.
func (b *DocumentBuilder) DrawRichText(pageIndex int, rt *RichText, x, y float64) *DocumentBuilder {
	if rt == nil || !b.pc.validPageIndex(pageIndex) {
		return b
	}
	currentX := x
	for _, seg := range rt.Segments {
		b.drawTextColoredAt(pageIndex, seg.Text, currentX, y, seg.Style.FontName, seg.Style.FontSize, seg.Style.Color, seg.Style.LetterSpacing)
		currentX += b.textWidthStyle(seg.Text, seg.Style.FontSize, seg.Style.FontName, seg.Style.LetterSpacing, 0)
	}
	return b
}

// DrawText queues text to be drawn on the last added page at (x, y) using the given font and size.
// FontName should be a standard PDF base font (e.g. Helvetica, Times-Roman). Call after AddPage().
func (b *DocumentBuilder) DrawText(textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.DrawTextColored(textStr, x, y, fontName, fontSize, ColorBlack)
}

// DrawTextColored queues text drawn in the specified RGB color on the last added page.
// It behaves like DrawText but sets the fill color for that run.
func (b *DocumentBuilder) DrawTextColored(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	return b.drawTextColoredAt(b.lastPageIndex(), textStr, x, y, fontName, fontSize, color, 0)
}

// DrawTextWithSpacing draws text at (x, y) with explicit color, letter and word spacing.
// This is useful for reconstructing PDFs where Tc/Tw were used and must be preserved.
func (b *DocumentBuilder) DrawTextWithSpacing(textStr string, x, y float64, fontName string, fontSize float64, color Color, letterSpacing, wordSpacing float64) *DocumentBuilder {
	idx := b.lastPageIndex()
	if !b.pc.validPageIndex(idx) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	segments := b.fc.resolveFont(textStr, fontName)
	currentX := x
	for _, seg := range segments {
		segWidth := b.textWidthStyle(seg.text, fontSize, seg.fontName, letterSpacing, wordSpacing)
		b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
			Text: seg.text, X: currentX, Y: y, FontName: seg.fontName, FontSize: fontSize,
			TextColorRGB:  [3]float64{color.R, color.G, color.B},
			LetterSpacing: letterSpacing,
			WordSpacing:   wordSpacing,
		})
		// Advance by width plus extra spacing for spaces (already accounted in textWidthStyle);
		// since textWidthStyle included wordSpacing, here we only move by segWidth.
		currentX += segWidth
	}
	return b
}

// DrawTextWithSpacingScale is like DrawTextWithSpacing but also sets HorizontalScale (Tz percent).
func (b *DocumentBuilder) DrawTextWithSpacingScale(textStr string, x, y float64, fontName string, fontSize float64, color Color, letterSpacing, wordSpacing, horizontalScale float64) *DocumentBuilder {
	idx := b.lastPageIndex()
	if !b.pc.validPageIndex(idx) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if horizontalScale == 0 {
		horizontalScale = 100
	}
	segments := b.fc.resolveFont(textStr, fontName)
	currentX := x
	for _, seg := range segments {
		segWidth := b.textWidthStyle(seg.text, fontSize, seg.fontName, letterSpacing, wordSpacing) * horizontalScale / 100
		b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
			Text: seg.text, X: currentX, Y: y, FontName: seg.fontName, FontSize: fontSize,
			TextColorRGB:    [3]float64{color.R, color.G, color.B},
			LetterSpacing:   letterSpacing,
			WordSpacing:     wordSpacing,
			HorizontalScale: horizontalScale,
		})
		currentX += segWidth
	}
	return b
}

// DrawTextRotated draws text at (x, y) rotated by deg degrees counter-clockwise.
// It uses DrawTextWithSpacingScale defaults (letterSpacing=0, wordSpacing=0, horizontalScale=100).
func (b *DocumentBuilder) DrawTextRotated(textStr string, x, y float64, fontName string, fontSize float64, color Color, deg float64) *DocumentBuilder {
	idx := b.lastPageIndex()
	if !b.pc.validPageIndex(idx) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	segments := b.fc.resolveFont(textStr, fontName)
	currentX := x
	for _, seg := range segments {
		segWidth := b.textWidthStyle(seg.text, fontSize, seg.fontName, 0, 0)
		b.pc.pages[idx].TextRuns = append(b.pc.pages[idx].TextRuns, textRun{
			Text:         seg.text,
			X:            currentX,
			Y:            y,
			FontName:     seg.fontName,
			FontSize:     fontSize,
			TextColorRGB: [3]float64{color.R, color.G, color.B},
			Rotation:     deg,
		})
		currentX += segWidth
	}
	return b
}

func (b *DocumentBuilder) drawTextColoredAt(pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, color Color, letterSpacing float64) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}

	segments := b.fc.resolveFont(textStr, fontName)
	currentX := x
	for _, seg := range segments {
		segWidth := b.textWidthStyle(seg.text, fontSize, seg.fontName, letterSpacing, 0)
		b.pc.pages[pageIndex].TextRuns = append(b.pc.pages[pageIndex].TextRuns, textRun{
			Text: seg.text, X: currentX, Y: y, FontName: seg.fontName, FontSize: fontSize,
			TextColorRGB:  [3]float64{color.R, color.G, color.B},
			LetterSpacing: letterSpacing,
		})
		currentX += segWidth
	}
	return b
}

// DrawTextCentered draws text horizontally centered around the point (cx, y).
// The text baseline is at y; cx is the horizontal center of the rendered text.
func (b *DocumentBuilder) DrawTextCentered(textStr string, cx, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.DrawTextCenteredColored(textStr, cx, y, fontName, fontSize, ColorBlack, 0)
}

// DrawTextCenteredColored draws colored text horizontally centered around the point (cx, y).
func (b *DocumentBuilder) DrawTextCenteredColored(textStr string, cx, y float64, fontName string, fontSize float64, color Color, letterSpacing float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	w := b.textWidthStyle(textStr, fontSize, fontName, letterSpacing, 0)
	return b.drawTextColoredAt(b.lastPageIndex(), textStr, cx-w/2, y, fontName, fontSize, color, letterSpacing)
}

// DrawTextRight draws text right-aligned so that its right edge is at x.
func (b *DocumentBuilder) DrawTextRight(textStr string, x, y float64, fontName string, fontSize float64) *DocumentBuilder {
	return b.DrawTextRightColored(textStr, x, y, fontName, fontSize, ColorBlack, 0)
}

// DrawTextRightColored draws colored text right-aligned so that its right edge is at x.
func (b *DocumentBuilder) DrawTextRightColored(textStr string, x, y float64, fontName string, fontSize float64, color Color, letterSpacing float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	w := b.textWidthStyle(textStr, fontSize, fontName, letterSpacing, 0)
	return b.drawTextColoredAt(b.lastPageIndex(), textStr, x-w, y, fontName, fontSize, color, letterSpacing)
}

// DrawTextWithUnderline draws text with an underline on the last added page.
func (b *DocumentBuilder) DrawTextWithUnderline(textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	idx := b.lastPageIndex()
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
	idx := b.lastPageIndex()
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
	return b.DrawHeadingColored(pageIndex, level, textStr, x, y, fontName, fontSize, ColorBlack)
}

// DrawHeadingColored behaves like DrawHeading but sets the text color.
func (b *DocumentBuilder) DrawHeadingColored(pageIndex int, level int, textStr string, x, y float64, fontName string, fontSize float64, color Color) *DocumentBuilder {
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
		TextColorRGB:    [3]float64{color.R, color.G, color.B},
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: color == ColorBlack,
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
	ls := style.DefaultListStyle()
	ls.FontName = fontName
	ls.FontSize = fontSize
	if ordered {
		ls.Marker = style.ListMarkerDecimal
	}
	return b.DrawListEnhanced(pageIndex, items, x, y, lineHeight, ls)
}

// DrawListEnhanced renders a list with advanced styling options.
func (b *DocumentBuilder) DrawListEnhanced(pageIndex int, items []string, x, y, lineHeight float64, ls style.ListStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(items) == 0 {
		return b
	}
	if ls.FontName == "" {
		ls.FontName = "Helvetica"
	}
	if ls.FontSize <= 0 {
		ls.FontSize = 12
	}
	if lineHeight <= 0 {
		lineHeight = ls.FontSize * 1.2
	}

	ps := &b.pc.pages[pageIndex]
	var listItems []taggedpkg.ListItem

	for idx, raw := range items {
		if raw == "" {
			continue
		}

		markerText, markerFont := b.getMarkerText(ls, idx)
		itemY := y - float64(len(listItems))*lineHeight

		markerFontSize := ls.MarkerFontSize
		if markerFontSize <= 0 {
			markerFontSize = ls.FontSize
		}

		// Draw Marker
		mcidMarker := ps.NextMCID
		ps.NextMCID++
		ps.TextRuns = append(ps.TextRuns, textRun{
			Text:            markerText,
			X:               x + float64(ls.Level)*ls.Indent,
			Y:               itemY,
			FontName:        markerFont,
			FontSize:        markerFontSize,
			TextColorRGB:    ls.Color.ToRGB(),
			UseDefaultColor: ls.Color == style.Black,
			MCID:            mcidMarker,
			HasMCID:         true,
		})

		// Draw Item Text
		mcidText := ps.NextMCID
		ps.NextMCID++
		ps.TextRuns = append(ps.TextRuns, textRun{
			Text:            raw,
			X:               x + float64(ls.Level)*ls.Indent + ls.MarkerOffset,
			Y:               itemY,
			FontName:        ls.FontName,
			FontSize:        ls.FontSize,
			TextColorRGB:    ls.Color.ToRGB(),
			UseDefaultColor: ls.Color == style.Black,
			MCID:            mcidText,
			HasMCID:         true,
		})

		listItems = append(listItems, taggedpkg.ListItem{MCID: mcidText}) // Simplification: only track text
	}

	if len(listItems) > 0 {
		b.tagging.Lists = append(b.tagging.Lists, taggedpkg.List{
			PageIndex: pageIndex,
			Ordered:   ls.Marker == style.ListMarkerDecimal,
			Items:     listItems,
		})
		b.tagging.RecordSectionList(len(b.tagging.Lists) - 1)
	}

	return b
}

func (b *DocumentBuilder) getMarkerText(ls style.ListStyle, index int) (string, string) {
	switch ls.Marker {
	case style.ListMarkerDisc:
		return "\u2022", "Symbol"
	case style.ListMarkerCircle:
		return "o", "Helvetica"
	case style.ListMarkerSquare:
		return "\u25a0", "ZapfDingbats"
	case style.ListMarkerDecimal:
		return fmt.Sprintf("%d.", index+1), ls.FontName
	case style.ListMarkerRomanUpper:
		return toRoman(index+1) + ".", ls.FontName
	case style.ListMarkerAlphaUpper:
		return string(rune('A'+(index%26))) + ".", ls.FontName
	case style.ListMarkerCustom:
		return ls.CustomMarker, ls.FontName
	default:
		return "\u2022", "Helvetica"
	}
}

func toRoman(n int) string {
	if n <= 0 {
		return ""
	}
	var res strings.Builder
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	for i, v := range vals {
		for n >= v {
			res.WriteString(syms[i])
			n -= v
		}
	}
	return res.String()
}

func reverseRunes(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

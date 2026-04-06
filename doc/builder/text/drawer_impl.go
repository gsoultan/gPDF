package text

import (
	"fmt"

	"github.com/gsoultan/gpdf/doc/builder"
	"github.com/gsoultan/gpdf/doc/style"
	btext "github.com/gsoultan/gpdf/doc/text"
)

type drawer struct {
	ta builder.TaggingAccess
}

// NewDrawer creates a Drawer. Pass nil for ta when tagged PDF is not needed.
func NewDrawer(ta builder.TaggingAccess) Drawer {
	return &drawer{ta: ta}
}

func (d *drawer) DrawText(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	page := pa.PageAt(pa.PageCount() - 1)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
	})
}

func (d *drawer) DrawTextColored(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64, color style.Color) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	page := pa.PageAt(pa.PageCount() - 1)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB: [3]float64{color.R, color.G, color.B},
	})
}

func (d *drawer) DrawTextCentered(pa builder.PageAccess, textStr string, cx, y float64, fontName string, fontSize float64) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	w := measureText(pa, textStr, fontSize, fontName)
	d.DrawText(pa, textStr, cx-w/2, y, fontName, fontSize)
}

func (d *drawer) DrawTextCenteredColored(pa builder.PageAccess, textStr string, cx, y float64, fontName string, fontSize float64, color style.Color) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	w := measureText(pa, textStr, fontSize, fontName)
	d.DrawTextColored(pa, textStr, cx-w/2, y, fontName, fontSize, color)
}

func (d *drawer) DrawTextRight(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	w := measureText(pa, textStr, fontSize, fontName)
	d.DrawText(pa, textStr, x-w, y, fontName, fontSize)
}

func (d *drawer) DrawTextRightColored(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64, color style.Color) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	w := measureText(pa, textStr, fontSize, fontName)
	d.DrawTextColored(pa, textStr, x-w, y, fontName, fontSize, color)
}

func (d *drawer) DrawTextWithUnderline(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64, color style.Color) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	page := pa.PageAt(pa.PageCount() - 1)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB: [3]float64{color.R, color.G, color.B},
		Underline:    true,
	})
}

func (d *drawer) DrawTextWithStrikethrough(pa builder.PageAccess, textStr string, x, y float64, fontName string, fontSize float64, color style.Color) {
	if pa.PageCount() == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	page := pa.PageAt(pa.PageCount() - 1)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text: textStr, X: x, Y: y, FontName: fontName, FontSize: fontSize,
		TextColorRGB:  [3]float64{color.R, color.G, color.B},
		Strikethrough: true,
	})
}

func (d *drawer) DrawTextBox(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions) {
	if textStr == "" || opts.Width <= 0 || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	d.layoutTextIntoPages(pa, pageIndex, textStr, x, y, fontName, fontSize, opts, false, "")
}

func (d *drawer) DrawTextBoxColored(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions, color style.Color) {
	if textStr == "" || opts.Width <= 0 || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	lines := wrapLines(pa, textStr, fontSize, opts.Width, fontName)
	curY := y
	for i, line := range lines {
		offsetX, wordSpacing := alignLine(pa, line, x, fontSize, fontName, opts, i, lines)
		page := pa.PageAt(pageIndex)
		page.TextRuns = append(page.TextRuns, builder.TextRun{
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
}

func (d *drawer) DrawTaggedParagraphBox(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions) {
	if textStr == "" || opts.Width <= 0 || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	d.layoutTextIntoPages(pa, pageIndex, textStr, x, y, fontName, fontSize, opts, true, "P")
}

func (d *drawer) DrawParagraph(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) {
	if textStr == "" || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	mcid := pa.NextMCID(pageIndex)
	page := pa.PageAt(pageIndex)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	if d.ta != nil {
		blockIdx := d.ta.RecordBlock(pageIndex, mcid, "P")
		d.ta.RecordSectionBlock(blockIdx)
	}
}

func (d *drawer) DrawTaggedCaption(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) {
	d.drawTaggedBlock(pa, pageIndex, textStr, x, y, fontName, fontSize, "Caption")
}

func (d *drawer) DrawTaggedQuote(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) {
	d.drawTaggedBlock(pa, pageIndex, textStr, x, y, fontName, fontSize, "Quote")
}

func (d *drawer) DrawTaggedCode(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64) {
	d.drawTaggedBlock(pa, pageIndex, textStr, x, y, fontName, fontSize, "Code")
}

func (d *drawer) DrawTaggedQuoteBox(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions) {
	if textStr == "" || opts.Width <= 0 || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if opts.LineHeight <= 0 {
		opts.LineHeight = fontSize * 1.2
	}
	d.layoutTextIntoPages(pa, pageIndex, textStr, x, y, fontName, fontSize, opts, true, "Quote")
}

func (d *drawer) DrawTaggedCodeBlock(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions) {
	if textStr == "" || opts.Width <= 0 || !pa.ValidPageIndex(pageIndex) {
		return
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
	d.layoutTextIntoPages(pa, pageIndex, textStr, x, y, fontName, fontSize, opts, true, "Code")
}

func (d *drawer) DrawHeading(pa builder.PageAccess, pageIndex int, level int, textStr string, x, y float64, fontName string, fontSize float64) {
	if textStr == "" || !pa.ValidPageIndex(pageIndex) {
		return
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
	mcid := pa.NextMCID(pageIndex)
	page := pa.PageAt(pageIndex)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	if d.ta != nil {
		role := fmt.Sprintf("H%d", level)
		blockIdx := d.ta.RecordBlock(pageIndex, mcid, role)
		d.ta.RecordSectionBlock(blockIdx)
	}
}

func (d *drawer) DrawList(pa builder.PageAccess, pageIndex int, items []string, x, y, lineHeight float64, ordered bool, fontName string, fontSize float64) {
	if !pa.ValidPageIndex(pageIndex) || len(items) == 0 {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if lineHeight <= 0 {
		lineHeight = fontSize * 1.2
	}
	page := pa.PageAt(pageIndex)
	var mcids []int
	itemCount := 0
	for idx, raw := range items {
		if raw == "" {
			continue
		}
		label := "• "
		if ordered {
			label = fmt.Sprintf("%d. ", idx+1)
		}
		itemText := label + raw
		itemY := y - float64(itemCount)*lineHeight
		mcid := pa.NextMCID(pageIndex)
		page.TextRuns = append(page.TextRuns, builder.TextRun{
			Text:            itemText,
			X:               x,
			Y:               itemY,
			FontName:        fontName,
			FontSize:        fontSize,
			MCID:            mcid,
			HasMCID:         true,
			UseDefaultColor: true,
		})
		mcids = append(mcids, mcid)
		itemCount++
	}
	if len(mcids) == 0 {
		return
	}
	if d.ta != nil {
		listIdx := d.ta.RecordList(pageIndex, ordered, mcids)
		d.ta.RecordSectionList(listIdx)
	}
}

// drawTaggedBlock adds a single-line tagged text block with the given structure role.
func (d *drawer) drawTaggedBlock(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, role string) {
	if textStr == "" || !pa.ValidPageIndex(pageIndex) {
		return
	}
	fontName, fontSize = defaultFont(fontName, fontSize)
	if d.ta != nil {
		d.ta.MarkTagged()
	}
	mcid := pa.NextMCID(pageIndex)
	page := pa.PageAt(pageIndex)
	page.TextRuns = append(page.TextRuns, builder.TextRun{
		Text:            textStr,
		X:               x,
		Y:               y,
		FontName:        fontName,
		FontSize:        fontSize,
		MCID:            mcid,
		HasMCID:         true,
		UseDefaultColor: true,
	})
	if d.ta != nil {
		blockIdx := d.ta.RecordBlock(pageIndex, mcid, role)
		d.ta.RecordSectionBlock(blockIdx)
	}
}

// layoutTextIntoPages performs line layout for paragraph-like text and appends textRuns.
// For tagged text, layout is restricted to a single page; AllowPageBreak is ignored.
func (d *drawer) layoutTextIntoPages(pa builder.PageAccess, pageIndex int, textStr string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions, isTagged bool, role string) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	lines := wrapLines(pa, textStr, fontSize, opts.Width, fontName)
	if len(lines) == 0 {
		return
	}

	height := pa.PageHeight(pageIndex)
	if height <= 0 {
		height = 842
	}
	const marginBottom = 40.0

	currentPage := pageIndex
	currentY := y
	var mcid int
	if isTagged && d.ta != nil {
		mcid = pa.NextMCID(pageIndex)
		blockIdx := d.ta.RecordBlock(pageIndex, mcid, role)
		d.ta.RecordSectionBlock(blockIdx)
	}

	for i, line := range lines {
		if isTagged && currentPage != pageIndex {
			break
		}
		if !isTagged && opts.AllowPageBreak && currentY < marginBottom {
			if currentPage+1 >= pa.PageCount() {
				break
			}
			currentPage++
			currentY = pa.PageHeight(currentPage) - marginBottom
		}

		offsetX, wordSpacing := alignLine(pa, line, x, fontSize, fontName, opts, i, lines)

		page := pa.PageAt(currentPage)
		r := builder.TextRun{
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
		page.TextRuns = append(page.TextRuns, r)
		currentY -= opts.LineHeight
	}
}

// alignLine computes the x-offset and word spacing for a line based on alignment options.
func alignLine(pa builder.PageAccess, line string, x, fontSize float64, fontName string, opts btext.LayoutOptions, lineIndex int, lines []string) (offsetX, wordSpacing float64) {
	lineWidth := measureText(pa, line, fontSize, fontName)
	offsetX = x
	free := opts.Width - lineWidth
	isLastInParagraph := lineIndex == len(lines)-1 || lines[lineIndex+1] == ""

	switch opts.Align {
	case btext.AlignCenter:
		if free > 0 {
			offsetX = x + free/2
		}
	case btext.AlignRight:
		if free > 0 {
			offsetX = x + free
		}
	case btext.AlignJustify:
		if !isLastInParagraph && free > 0 {
			numSpaces := countWordSpaces(line)
			if numSpaces > 0 {
				wordSpacing = free / float64(numSpaces)
			}
		}
	}
	return offsetX, wordSpacing
}

func defaultFont(fontName string, fontSize float64) (string, float64) {
	if fontName == "" {
		fontName = "Helvetica"
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	return fontName, fontSize
}

func fontWidthFunc(pa builder.PageAccess, fontName string) btext.FontWidthFunc {
	f := pa.FontByName(fontName)
	if f != nil {
		return func(s string, fontSize float64) float64 {
			return f.TextWidth(s, fontSize)
		}
	}
	return func(s string, fontSize float64) float64 {
		return btext.ApproxWidth(s, fontSize)
	}
}

func wrapLines(pa builder.PageAccess, s string, fontSize, width float64, fontName string) []string {
	return btext.WrapLines(s, fontSize, width, fontWidthFunc(pa, fontName))
}

func measureText(pa builder.PageAccess, s string, fontSize float64, fontName string) float64 {
	if s == "" || fontSize <= 0 {
		return 0
	}
	f := pa.FontByName(fontName)
	if f != nil {
		return f.TextWidth(s, fontSize)
	}
	return btext.ApproxWidth(s, fontSize)
}

func countWordSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		}
	}
	return count
}

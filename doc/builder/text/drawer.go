package text

import (
	"gpdf/doc/builder"
	"gpdf/doc/style"
	btext "gpdf/doc/text"
)

// Drawer handles text drawing operations on PDF pages.
type Drawer interface {
	DrawText(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64)
	DrawTextColored(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64, color style.Color)
	DrawTextCentered(pa builder.PageAccess, text string, cx, y float64, fontName string, fontSize float64)
	DrawTextCenteredColored(pa builder.PageAccess, text string, cx, y float64, fontName string, fontSize float64, color style.Color)
	DrawTextRight(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64)
	DrawTextRightColored(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64, color style.Color)
	DrawTextWithUnderline(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64, color style.Color)
	DrawTextWithStrikethrough(pa builder.PageAccess, text string, x, y float64, fontName string, fontSize float64, color style.Color)
	DrawTextBox(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions)
	DrawTextBoxColored(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions, color style.Color)
	DrawTaggedParagraphBox(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions)
	DrawParagraph(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64)
	DrawTaggedCaption(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64)
	DrawTaggedQuote(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64)
	DrawTaggedCode(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64)
	DrawTaggedQuoteBox(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions)
	DrawTaggedCodeBlock(pa builder.PageAccess, pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts btext.LayoutOptions)
	DrawHeading(pa builder.PageAccess, pageIndex int, level int, text string, x, y float64, fontName string, fontSize float64)
	DrawList(pa builder.PageAccess, pageIndex int, items []string, x, y, lineHeight float64, ordered bool, fontName string, fontSize float64)
}

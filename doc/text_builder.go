package doc

// TextBuilder covers basic and tagged text layout on pages.
type TextBuilder interface {
	DrawText(text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTextBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawTaggedParagraphBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawParagraph(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawHeading(pageIndex int, level int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawList(pageIndex int, items []string, x, y, lineHeight float64, ordered bool, fontName string, fontSize float64) *DocumentBuilder
}

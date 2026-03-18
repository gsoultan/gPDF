package doc

// TaggedContentBuilder exposes text-oriented tagged content helpers (sections, captions, quotes, code blocks).
type TaggedContentBuilder interface {
	SetTagged() *DocumentBuilder
	BeginSection() *DocumentBuilder
	EndSection() *DocumentBuilder
	DrawTaggedCaption(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedQuote(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedCode(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedQuoteBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawTaggedCodeBlock(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
}

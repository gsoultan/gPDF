package doc

import "gpdf/doc/layer"

// PageBuilder configures page size and manages pages.
type PageBuilder interface {
	PageSize(width, height float64) *DocumentBuilder
	AddPage() *DocumentBuilder
}

// MetadataBuilder sets document-level metadata fields.
type MetadataBuilder interface {
	Title(s string) *DocumentBuilder
	Author(s string) *DocumentBuilder
	Subject(s string) *DocumentBuilder
	Keywords(s string) *DocumentBuilder
	Creator(s string) *DocumentBuilder
	Producer(s string) *DocumentBuilder
	Metadata(xmp []byte) *DocumentBuilder
	SetLanguage(lang string) *DocumentBuilder
}

// OutlineBuilder manages bookmarks, named destinations, and link annotations.
type OutlineBuilder interface {
	AddOutline(title string, pageIndex int) *DocumentBuilder
	AddOutlineURL(title string, url string) *DocumentBuilder
	AddOutlineToDest(title string, destName string) *DocumentBuilder
	AddNamedDest(name string, pageIndex int) *DocumentBuilder
	AddLinkToPage(pageIndex int, llx, lly, urx, ury float64, destPageIndex int) *DocumentBuilder
	AddLinkToDest(pageIndex int, llx, lly, urx, ury float64, destName string) *DocumentBuilder
	AddLinkToURL(pageIndex int, llx, lly, urx, ury float64, url string) *DocumentBuilder
}

// TextBuilder covers basic and tagged text layout on pages.
type TextBuilder interface {
	DrawText(text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTextBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawTaggedParagraphBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawParagraph(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawHeading(pageIndex int, level int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawList(pageIndex int, items []string, x, y, lineHeight float64, ordered bool, fontName string, fontSize float64) *DocumentBuilder
}

// ImageBuilder covers image placement on pages.
type ImageBuilder interface {
	DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder
	DrawJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string) *DocumentBuilder
	DrawPNG(x, y, widthPt, heightPt float64, pngData []byte) *DocumentBuilder
}

// TaggedBuilder exposes high-level tagged content helpers (sections, figures, code/quote blocks).
type TaggedBuilder interface {
	SetTagged() *DocumentBuilder
	BeginSection() *DocumentBuilder
	EndSection() *DocumentBuilder
	DrawTaggedCaption(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedQuote(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedCode(pageIndex int, text string, x, y float64, fontName string, fontSize float64) *DocumentBuilder
	DrawTaggedQuoteBox(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawTaggedCodeBlock(pageIndex int, text string, x, y float64, fontName string, fontSize float64, opts TextLayoutOptions) *DocumentBuilder
	DrawTaggedFigure(pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, alt string) *DocumentBuilder
	DrawTaggedJPEG(pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, alt string) *DocumentBuilder
	DrawTaggedPNG(pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) *DocumentBuilder
}

// TableBuilderAPI describes the fluent API for building tagged tables.
type TableBuilderAPI interface {
	BeginTable(pageIndex int, x, y, width, height float64, cols int) *TableBuilder
}

// FormBuilder configures AcroForm and creates fields/widgets.
type FormBuilder interface {
	SetAcroForm() *DocumentBuilder
	SetAcroFormSigFlags(flags int) *DocumentBuilder
	AddTextField(pageIndex int, llx, lly, urx, ury float64, name, value, tooltip string, required bool) *DocumentBuilder
	AddCheckBox(pageIndex int, llx, lly, urx, ury float64, name string, checked bool, tooltip string, required bool) *DocumentBuilder
	AddRadioButton(pageIndex int, llx, lly, urx, ury float64, groupName, value string, checked bool, tooltip string) *DocumentBuilder
	AddSubmitButton(pageIndex int, llx, lly, urx, ury float64, name, label, submitURL, tooltip string) *DocumentBuilder
}

// LayerBuilder controls optional content groups (layers).
type LayerBuilder interface {
	BeginLayer(name string, onByDefault bool) *layer.Handle
	DrawInLayer(lh *layer.Handle, pageIndex int, draw func(db *DocumentBuilder)) *DocumentBuilder
}

// Ensure DocumentBuilder implements the capability interfaces.
var (
	_ PageBuilder     = (*DocumentBuilder)(nil)
	_ MetadataBuilder = (*DocumentBuilder)(nil)
	_ OutlineBuilder  = (*DocumentBuilder)(nil)
	_ TextBuilder     = (*DocumentBuilder)(nil)
	_ ImageBuilder    = (*DocumentBuilder)(nil)
	_ TaggedBuilder   = (*DocumentBuilder)(nil)
	_ TableBuilderAPI = (*DocumentBuilder)(nil)
	_ FormBuilder     = (*DocumentBuilder)(nil)
	_ LayerBuilder    = (*DocumentBuilder)(nil)
)

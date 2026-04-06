package doc

import (
	"github.com/gsoultan/gpdf/doc/image"
	"github.com/gsoultan/gpdf/doc/style"
	"github.com/gsoultan/gpdf/model"
)

// Page represents a single page in the document and provides a fluent API for drawing on it.
type Page struct {
	builder   *DocumentBuilder
	pageIndex int
}

// At returns a point for positioning elements.
func (p *Page) At(x, y float64) Pt {
	return Pt{X: x, Y: y}
}

// Text creates a new TextObject for drawing on this page.
func (p *Page) Text(s string) *TextObject {
	return &TextObject{
		page:  p,
		text:  s,
		style: p.builder.getEffectiveStyle(),
	}
}

// Heading creates a new HeadingObject for drawing tagged headings.
func (p *Page) Heading(text string, level int) *HeadingObject {
	return &HeadingObject{
		page:  p,
		text:  text,
		level: level,
	}
}

// Line creates a new LineObject for drawing straight lines.
func (p *Page) Line(start, end Pt) *LineObject {
	return &LineObject{
		page:  p,
		start: start,
		end:   end,
		style: LineStyle{Width: 1, Color: ColorBlack},
	}
}

// Rect creates a new RectObject for drawing rectangles.
func (p *Page) Rect(rect Rect) *RectObject {
	return &RectObject{
		page: p,
		rect: rect,
	}
}

// Table starts a table on this page with the given number of columns.
func (p *Page) Table(cols int) ITableBuilder {
	ps := &p.builder.pc.pages[p.pageIndex]
	return p.builder.BeginTable(p.pageIndex, ps.CurrX, ps.CurrY, p.builder.pageWidth(p.pageIndex)-ps.CurrX-72, 0, cols)
}

// Image creates a new ImageObject for drawing on this page.
func (p *Page) Image(data []byte, w, h float64) *ImageObject {
	obj := &ImageObject{
		page: p,
		data: data,
		w:    w,
		h:    h,
	}
	// Attempt to auto-detect dimensions for PNG/JPEG
	if pxW, pxH, isJPEG, isPNG, err := image.DetectDimensions(data); err == nil {
		obj.pxW = pxW
		obj.pxH = pxH
		obj.isJPEG = isJPEG
		obj.isPNG = isPNG
	}
	return obj
}

// Flow starts a flowing layout region on this page.
func (p *Page) Flow(opts FlowOptions) *FlowBuilder {
	top := opts.Top
	if top == 0 {
		top = opts.Margin
	}
	if top == 0 {
		top = 72
	}
	return &FlowBuilder{
		builder:   p.builder,
		opts:      opts,
		pageIndex: p.pageIndex,
		currY:     p.builder.pageHeight(p.pageIndex) - top,
	}
}

// BeginSection starts a logical section for tagged content on this page.
func (p *Page) BeginSection() *Page {
	p.builder.BeginSection()
	return p
}

// EndSection ends the current section on this page.
func (p *Page) EndSection() *Page {
	p.builder.EndSection()
	return p
}

// TextBox creates a new TextBoxObject for drawing wrapped text.
func (p *Page) TextBox(s string) *TextBoxObject {
	style := p.builder.getEffectiveStyle()
	return &TextBoxObject{
		page:  p,
		text:  s,
		style: style,
		opts:  TextLayoutOptions{Width: 200, LineHeight: style.FontSize * 1.2},
	}
}

// Paragraph creates a new TextBoxObject for drawing wrapped text as a paragraph.
func (p *Page) Paragraph(s string) *TextBoxObject {
	ps := &p.builder.pc.pages[p.pageIndex]
	return p.TextBox(s).AsParagraph().Width(p.builder.pageWidth(p.pageIndex) - ps.CurrX - 72)
}

// List creates a new ListObject for drawing bulleted or numbered lists.
func (p *Page) List(items []string) *ListObject {
	textStyle := p.builder.getEffectiveStyle()
	listStyle := style.DefaultListStyle()
	listStyle.FontName = textStyle.FontName
	listStyle.FontSize = textStyle.FontSize
	listStyle.Color = textStyle.Color

	return &ListObject{
		page:       p,
		items:      items,
		listStyle:  listStyle,
		textStyle:  textStyle,
		lineHeight: textStyle.FontSize * 1.25,
	}
}

// CurrentY returns the current Y position of the "flow" if any, or some default.
func (p *Page) CurrentY() float64 {
	return p.builder.pc.pages[p.pageIndex].CurrY
}

// End returns the underlying DocumentBuilder for chaining.
func (p *Page) End() *DocumentBuilder {
	return p.builder
}

// TextObject handles fluent text drawing.
type TextObject struct {
	page  *Page
	text  string
	pt    Pt
	atSet bool
	style TextStyle
	align TextAlignment
}

// At sets the position for the text.
func (o *TextObject) At(x, y float64) *TextObject {
	o.pt = Pt{X: x, Y: y}
	o.atSet = true
	return o
}

// Align sets the horizontal alignment for the text.
func (o *TextObject) Align(a TextAlignment) *TextObject {
	o.align = a
	return o
}

// Font sets the font for the text.
func (o *TextObject) Font(name string) *TextObject {
	o.style.FontName = name
	return o
}

// Size sets the font size for the text.
func (o *TextObject) Size(size float64) *TextObject {
	o.style.FontSize = size
	return o
}

// Color sets the color for the text.
func (o *TextObject) Color(c Color) *TextObject {
	o.style.Color = c
	return o
}

// LetterSpacing sets the character spacing for the text.
func (o *TextObject) LetterSpacing(spacing float64) *TextObject {
	o.style.LetterSpacing = spacing
	return o
}

// Style sets the complete text style.
func (o *TextObject) Style(s TextStyle) *TextObject {
	o.style = s
	return o
}

// Draw renders the text to the page.
func (o *TextObject) Draw() *Page {
	ps := &o.page.builder.pc.pages[o.page.pageIndex]
	if !o.atSet {
		o.pt.X = ps.CurrX
		o.pt.Y = ps.CurrY
	}

	switch o.align {
	case AlignCenter:
		o.page.builder.DrawTextCenteredColored(o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize, o.style.Color, o.style.LetterSpacing)
	case AlignRight:
		o.page.builder.DrawTextRightColored(o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize, o.style.Color, o.style.LetterSpacing)
	default:
		o.page.builder.drawTextColoredAt(o.page.pageIndex, o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize, o.style.Color, o.style.LetterSpacing)
	}

	if !o.atSet {
		ps.CurrY -= o.style.FontSize * 1.2
	}
	return o.page
}

// TextBoxObject handles fluent wrapped text drawing.
type TextBoxObject struct {
	page  *Page
	text  string
	pt    Pt
	atSet bool
	style TextStyle
	opts  TextLayoutOptions
	role  model.Name
}

func (o *TextBoxObject) AsQuote() *TextBoxObject {
	o.role = model.Name("Quote")
	return o
}

func (o *TextBoxObject) AsCode() *TextBoxObject {
	o.role = model.Name("Code")
	return o
}

func (o *TextBoxObject) AsParagraph() *TextBoxObject {
	o.role = model.Name("P")
	return o
}

func (o *TextBoxObject) At(x, y float64) *TextBoxObject {
	o.pt = Pt{X: x, Y: y}
	o.atSet = true
	return o
}

func (o *TextBoxObject) Width(w float64) *TextBoxObject {
	o.opts.Width = w
	return o
}

func (o *TextBoxObject) LineHeight(h float64) *TextBoxObject {
	o.opts.LineHeight = h
	return o
}

func (o *TextBoxObject) Align(align TextAlignment) *TextBoxObject {
	switch align {
	case AlignCenter:
		o.opts.Align = TextAlignCenter
	case AlignRight:
		o.opts.Align = TextAlignRight
	case AlignJustify:
		o.opts.Align = TextAlignJustify
	default:
		o.opts.Align = TextAlignLeft
	}
	return o
}

func (o *TextBoxObject) LetterSpacing(spacing float64) *TextBoxObject {
	o.opts.LetterSpacing = spacing
	return o
}

func (o *TextBoxObject) Font(name string) *TextBoxObject {
	o.style.FontName = name
	return o
}

func (o *TextBoxObject) Size(size float64) *TextBoxObject {
	o.style.FontSize = size
	return o
}

func (o *TextBoxObject) Color(c Color) *TextBoxObject {
	o.style.Color = c
	o.opts.Color = c
	o.opts.HasColor = true
	return o
}

func (o *TextBoxObject) Style(s TextStyle) *TextBoxObject {
	o.style = s
	o.opts.LetterSpacing = s.LetterSpacing
	o.opts.Color = s.Color
	o.opts.HasColor = true
	o.opts.IsVertical = s.IsVertical
	return o
}

func (o *TextBoxObject) Draw() *Page {
	ps := &o.page.builder.pc.pages[o.page.pageIndex]
	if !o.atSet {
		o.pt.X = ps.CurrX
		o.pt.Y = ps.CurrY
	}

	var newIdx int
	var newY float64
	if o.role == "" {
		newIdx, newY = o.page.builder.layoutTextIntoPages(o.page.pageIndex, o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize, o.opts, false, "")
	} else {
		newIdx, newY = o.page.builder.layoutTextIntoPages(o.page.pageIndex, o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize, o.opts, true, o.role)
	}

	if !o.atSet {
		o.page.pageIndex = newIdx
		o.page.builder.pc.pages[newIdx].CurrY = newY
	}
	return o.page
}

// ListObject handles fluent list drawing.
type ListObject struct {
	page       *Page
	items      []string
	pt         Pt
	atSet      bool
	listStyle  style.ListStyle
	textStyle  TextStyle
	lineHeight float64
}

func (o *ListObject) At(x, y float64) *ListObject {
	o.pt = Pt{X: x, Y: y}
	o.atSet = true
	return o
}

func (o *ListObject) LineHeight(h float64) *ListObject {
	o.lineHeight = h
	return o
}

func (o *ListObject) Marker(m style.ListMarker) *ListObject {
	o.listStyle.Marker = m
	return o
}

func (o *ListObject) CustomMarker(s string) *ListObject {
	o.listStyle.Marker = style.ListMarkerCustom
	o.listStyle.CustomMarker = s
	return o
}

func (o *ListObject) Indent(v float64) *ListObject {
	o.listStyle.Indent = v
	return o
}

func (o *ListObject) Level(l int) *ListObject {
	o.listStyle.Level = l
	return o
}

func (o *ListObject) Ordered(b bool) *ListObject {
	if b {
		o.listStyle.Marker = style.ListMarkerDecimal
	} else {
		o.listStyle.Marker = style.ListMarkerDisc
	}
	return o
}

func (o *ListObject) Font(name string) *ListObject {
	o.listStyle.FontName = name
	o.textStyle.FontName = name
	return o
}

func (o *ListObject) Size(size float64) *ListObject {
	o.listStyle.FontSize = size
	o.textStyle.FontSize = size
	return o
}

func (o *ListObject) Color(c Color) *ListObject {
	o.listStyle.Color = c
	o.textStyle.Color = c
	return o
}

func (o *ListObject) Draw() *Page {
	ps := &o.page.builder.pc.pages[o.page.pageIndex]
	if !o.atSet {
		o.pt.X = ps.CurrX
		o.pt.Y = ps.CurrY
	}

	o.page.builder.DrawListEnhanced(o.page.pageIndex, o.items, o.pt.X, o.pt.Y, o.lineHeight, o.listStyle)

	if !o.atSet {
		ps.CurrY -= float64(len(o.items)) * o.lineHeight
	}
	return o.page
}

// HeadingObject handles fluent heading drawing.
type HeadingObject struct {
	page  *Page
	text  string
	level int
	pt    Pt
	atSet bool
	style TextStyle
}

// At sets the position for the heading.
func (o *HeadingObject) At(x, y float64) *HeadingObject {
	o.pt = Pt{X: x, Y: y}
	o.atSet = true
	return o
}

// Draw renders the heading to the page.
func (o *HeadingObject) Draw() *Page {
	ps := &o.page.builder.pc.pages[o.page.pageIndex]
	if !o.atSet {
		o.pt.X = ps.CurrX
		o.pt.Y = ps.CurrY
	}

	o.page.builder.DrawHeading(o.page.pageIndex, o.level, o.text, o.pt.X, o.pt.Y, o.style.FontName, o.style.FontSize)

	if !o.atSet {
		// Heading usually has more space
		fs := o.style.FontSize
		if fs <= 0 {
			fs = 12 * 1.5 // Rough estimate if not set
		}
		ps.CurrY -= fs * 1.5
	}
	return o.page
}

// LineObject handles fluent line drawing.
type LineObject struct {
	page  *Page
	start Pt
	end   Pt
	style LineStyle
}

// Style sets the line style.
func (o *LineObject) Style(s LineStyle) *LineObject {
	o.style = s
	return o
}

// Color sets the line color.
func (o *LineObject) Color(c Color) *LineObject {
	o.style.Color = c
	return o
}

// Width sets the line width.
func (o *LineObject) Width(w float64) *LineObject {
	o.style.Width = w
	return o
}

// Draw renders the line to the page.
func (o *LineObject) Draw() *Page {
	o.page.builder.DrawLine(o.page.pageIndex, o.start.X, o.start.Y, o.end.X, o.end.Y, o.style)
	return o.page
}

// RectObject handles fluent rectangle drawing.
type RectObject struct {
	page      *Page
	rect      Rect
	lineStyle LineStyle
	fillColor Color
	hasFill   bool
}

// Stroke sets the stroke style for the rectangle.
func (o *RectObject) Stroke(s LineStyle) *RectObject {
	o.lineStyle = s
	return o
}

// Fill sets the fill color for the rectangle.
func (o *RectObject) Fill(c Color) *RectObject {
	o.fillColor = c
	o.hasFill = true
	return o
}

// Draw renders the rectangle to the page.
func (o *RectObject) Draw() *Page {
	if o.hasFill && o.lineStyle.Width > 0 {
		o.page.builder.FillStrokeRect(o.page.pageIndex, o.rect.X, o.rect.Y, o.rect.W, o.rect.H, o.fillColor, o.lineStyle)
	} else if o.hasFill {
		o.page.builder.FillRect(o.page.pageIndex, o.rect.X, o.rect.Y, o.rect.W, o.rect.H, o.fillColor)
	} else {
		o.page.builder.DrawRect(o.page.pageIndex, o.rect.X, o.rect.Y, o.rect.W, o.rect.H, o.lineStyle)
	}
	return o.page
}

// ImageObject handles fluent image drawing.
type ImageObject struct {
	page          *Page
	data          []byte
	w, h          float64
	pxW, pxH      int
	colorSpace    string
	bitsPerComp   int
	isJPEG, isPNG bool
	pt            Pt
	atSet         bool
}

// At sets the position for the image.
func (o *ImageObject) At(x, y float64) *ImageObject {
	o.pt = Pt{X: x, Y: y}
	o.atSet = true
	return o
}

// Px sets the pixel dimensions of the image. Required for raw RGB data.
func (o *ImageObject) Px(w, h int) *ImageObject {
	o.pxW = w
	o.pxH = h
	return o
}

// ColorSpace sets the color space of the image (e.g., "DeviceRGB").
func (o *ImageObject) ColorSpace(cs string) *ImageObject {
	o.colorSpace = cs
	return o
}

// Draw renders the image to the page.
func (o *ImageObject) Draw() *Page {
	ps := &o.page.builder.pc.pages[o.page.pageIndex]
	if !o.atSet {
		o.pt.X = ps.CurrX
		o.pt.Y = ps.CurrY - o.h
	}

	data := o.data
	pxW, pxH := o.pxW, o.pxH
	cs := o.colorSpace
	bpc := o.bitsPerComp
	isJPEG := o.isJPEG

	// If it's a PNG, we must decode it to raw RGB/RGBA because PDF doesn't support PNG directly
	if o.isPNG {
		if raw, w, h, colorspace, err := image.DecodePNGToRaw(o.data); err == nil {
			data = raw
			pxW = w
			pxH = h
			cs = colorspace
			bpc = 8
		}
	}

	if cs == "" {
		cs = "DeviceRGB"
	}
	if bpc <= 0 {
		bpc = 8
	}

	o.page.builder.DrawImage(o.pt.X, o.pt.Y, o.w, o.h, data, pxW, pxH, bpc, cs)
	// Set IsJPEG if detected
	if isJPEG {
		idx := o.page.pageIndex
		if len(o.page.builder.pc.pages) > idx {
			runs := o.page.builder.pc.pages[idx].ImageRuns
			if len(runs) > 0 {
				runs[len(runs)-1].IsJPEG = true
				o.page.builder.pc.pages[idx].ImageRuns = runs
			}
		}
	}

	if !o.atSet {
		ps.CurrY -= o.h
	}
	return o.page
}

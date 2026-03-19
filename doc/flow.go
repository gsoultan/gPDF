package doc

import (
	"gpdf/doc/image"
	"gpdf/doc/style"
	"gpdf/doc/text"
)

// FlowOptions configures the layout of a flowing content area.
type FlowOptions struct {
	Margin float64
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

// FlowBuilder provides a high-level API for automatic content flow and relative positioning.
type FlowBuilder struct {
	builder   *DocumentBuilder
	opts      FlowOptions
	pageIndex int
	currY     float64
	style     TextStyle
	align     text.Align
}

// Flow starts a flowing layout region on the current page.
func (b *DocumentBuilder) Flow(opts FlowOptions) *FlowBuilder {
	pIdx := len(b.pc.pages) - 1
	if pIdx < 0 {
		b.AddPage()
		pIdx = 0
	}

	top := opts.Top
	if top == 0 {
		top = opts.Margin
	}
	if top == 0 {
		top = 72
	}

	f := &FlowBuilder{
		builder:   b,
		opts:      opts,
		pageIndex: pIdx,
		currY:     b.pageHeight(pIdx) - top,
	}
	return f
}

func (f *FlowBuilder) Left() float64 {
	if f.opts.Left > 0 {
		return f.opts.Left
	}
	if f.opts.Margin > 0 {
		return f.opts.Margin
	}
	return f.builder.pc.pages[f.pageIndex].CurrX
}

func (f *FlowBuilder) Right() float64 {
	if f.opts.Right > 0 {
		return f.opts.Right
	}
	if f.opts.Margin > 0 {
		return f.opts.Margin
	}
	return 72
}

func (f *FlowBuilder) Bottom() float64 {
	if f.opts.Bottom > 0 {
		return f.opts.Bottom
	}
	if f.opts.Margin > 0 {
		return f.opts.Margin
	}
	return 72
}

func (f *FlowBuilder) Width() float64 {
	w, _ := f.builder.pc.pageSize[0], f.builder.pc.pageSize[1]
	if w == 0 {
		w = 595
	}
	return w - f.Left() - f.Right()
}

// Font sets the font for subsequent elements in the flow.
func (f *FlowBuilder) Font(name string) *FlowBuilder {
	f.style.FontName = name
	return f
}

// Size sets the font size for subsequent elements in the flow.
func (f *FlowBuilder) Size(size float64) *FlowBuilder {
	f.style.FontSize = size
	return f
}

// Color sets the text color for subsequent elements in the flow.
func (f *FlowBuilder) Color(c Color) *FlowBuilder {
	f.style.Color = c
	return f
}

// Align sets the text alignment for subsequent paragraphs in the flow.
func (f *FlowBuilder) Align(a TextAlignment) *FlowBuilder {
	f.align = text.Align(a)
	return f
}

func (f *FlowBuilder) getEffectiveStyle() TextStyle {
	s := f.builder.getEffectiveStyle()
	if f.style.FontName != "" {
		s.FontName = f.style.FontName
	}
	if f.style.FontSize != 0 {
		s.FontSize = f.style.FontSize
	}
	if f.style.Color != (Color{}) {
		s.Color = f.style.Color
	}
	if f.style.LetterSpacing != 0 {
		s.LetterSpacing = f.style.LetterSpacing
	}
	return s
}

// Heading adds a heading to the flow.
func (f *FlowBuilder) Heading(text string, level int) *FlowBuilder {
	style := f.getEffectiveStyle()
	fontSize := style.FontSize
	if f.style.FontSize == 0 {
		switch level {
		case 1:
			fontSize = 24
		case 2:
			fontSize = 18
		case 3:
			fontSize = 14
		default:
			fontSize = 12
		}
	}

	h := fontSize * 1.5
	if f.currY-h < f.Bottom() {
		f.newPage()
	}

	f.builder.DrawHeadingColored(f.pageIndex, level, text, f.Left(), f.currY, style.FontName, fontSize, style.Color)
	f.currY -= h
	f.syncCursor()
	return f
}

// Paragraph adds a wrapping paragraph to the flow.
func (f *FlowBuilder) Paragraph(text string) *FlowBuilder {
	style := f.getEffectiveStyle()
	if style.FontSize <= 0 {
		style.FontSize = 12
	}
	if style.FontName == "" {
		style.FontName = "Helvetica"
	}

	opts := TextLayoutOptions{
		Width:          f.Width(),
		AllowPageBreak: true,
		Color:          style.Color,
		HasColor:       style.Color != (Color{}),
		Align:          f.align,
	}

	f.pageIndex, f.currY = f.builder.layoutTextIntoPages(f.pageIndex, text, f.Left(), f.currY, style.FontName, style.FontSize, opts, false, "")
	f.syncCursor()
	return f
}

// Space adds vertical space to the flow.
func (f *FlowBuilder) Space(h float64) *FlowBuilder {
	f.currY -= h
	f.syncCursor()
	return f
}

// Image adds an image to the flow.
func (f *FlowBuilder) Image(data []byte, w, h float64) *FlowBuilder {
	if f.currY-h < f.Bottom() {
		f.newPage()
	}

	info, err := image.ProcessImage(data)
	if err != nil {
		return f
	}

	f.builder.DrawImage(f.Left(), f.currY-h, w, h, info.Raw, info.WidthPx, info.HeightPx, info.BitsPerComponent, info.ColorSpace)
	if info.IsJPEG {
		idx := f.pageIndex
		if idx >= 0 && idx < len(f.builder.pc.pages) {
			runs := f.builder.pc.pages[idx].ImageRuns
			if len(runs) > 0 {
				runs[len(runs)-1].IsJPEG = true
				f.builder.pc.pages[idx].ImageRuns = runs
			}
		}
	}
	f.currY -= h
	f.syncCursor()
	return f
}

// List adds a list of items to the flow.
func (f *FlowBuilder) List(items []string, ordered bool) *FlowBuilder {
	s := f.getEffectiveStyle()
	lineHeight := s.FontSize * 1.25
	h := float64(len(items)) * lineHeight
	if f.currY-h < f.Bottom() {
		f.newPage()
	}

	ls := style.DefaultListStyle()
	ls.FontName = s.FontName
	ls.FontSize = s.FontSize
	ls.Color = s.Color
	if ordered {
		ls.Marker = style.ListMarkerDecimal
	}

	f.builder.DrawListEnhanced(f.pageIndex, items, f.Left(), f.currY, lineHeight, ls)
	f.currY -= h
	f.syncCursor()
	return f
}

// Line adds a horizontal line across the flow.
func (f *FlowBuilder) Line(width float64, c Color) *FlowBuilder {
	if f.currY-10 < f.Bottom() {
		f.newPage()
	}
	if c == (Color{}) {
		c = f.style.Color
		if c == (Color{}) {
			c = ColorBlack
		}
	}
	f.builder.DrawLine(f.pageIndex, f.Left(), f.currY-5, f.Left()+f.Width(), f.currY-5, LineStyle{Width: width, Color: c})
	f.currY -= 10
	f.syncCursor()
	return f
}

// Rect adds a rectangle to the flow.
func (f *FlowBuilder) Rect(h float64, style LineStyle, fill Color, hasFill bool) *FlowBuilder {
	if f.currY-h < f.Bottom() {
		f.newPage()
	}
	if hasFill && style.Width > 0 {
		f.builder.FillStrokeRect(f.pageIndex, f.Left(), f.currY-h, f.Width(), h, fill, style)
	} else if hasFill {
		f.builder.FillRect(f.pageIndex, f.Left(), f.currY-h, f.Width(), h, fill)
	} else {
		f.builder.DrawRect(f.pageIndex, f.Left(), f.currY-h, f.Width(), h, style)
	}
	f.currY -= h
	f.syncCursor()
	return f
}

func (f *FlowBuilder) syncCursor() {
	if f.builder.pc.validPageIndex(f.pageIndex) {
		ps := &f.builder.pc.pages[f.pageIndex]
		ps.CurrX = f.Left()
		ps.CurrY = f.currY
	}
}

func (f *FlowBuilder) newPage() {
	f.builder.AddPage()
	f.pageIndex = len(f.builder.pc.pages) - 1
	top := f.opts.Top
	if top == 0 {
		top = f.opts.Margin
	}
	if top == 0 {
		top = 72
	}
	f.currY = f.builder.pageHeight(f.pageIndex) - top
}

// Table starts a table in the flow.
func (f *FlowBuilder) Table(cols int) ITableBuilder {
	tb := f.builder.BeginTable(f.pageIndex, f.Left(), f.currY, f.Width(), 0, cols)
	if t, ok := tb.(*TableBuilder); ok {
		t.flow = f
	}
	return tb
}

// End returns the underlying DocumentBuilder.
func (f *FlowBuilder) End() *DocumentBuilder {
	return f.builder
}

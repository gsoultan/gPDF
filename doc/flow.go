package doc

import (
	"github.com/gsoultan/gpdf/doc/image"
	"github.com/gsoultan/gpdf/doc/style"
	"github.com/gsoultan/gpdf/doc/text"
	"github.com/gsoultan/gpdf/model"
)

// FloatingImage represents an image that text should wrap around.
type FloatingImage struct {
	X, Y          float64
	Width, Height float64
	Margin        float64
	Wrap          style.ImageWrap
}

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
	floating  []FloatingImage
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

// Image adds an image to the flow.
func (f *FlowBuilder) Image(data []byte, w, h float64) *FlowBuilder {
	return f.ImageWithLayout(data, style.ImageLayout{Width: w, Height: h}, style.DefaultImageStyle())
}

// ImageWithLayout adds an image to the flow with the given layout and style.
func (f *FlowBuilder) ImageWithLayout(imgData []byte, layout style.ImageLayout, s style.ImageStyle) *FlowBuilder {
	// 1. Resolve dimensions
	w, h := layout.Width, layout.Height
	if w <= 0 || h <= 0 {
		// Try to decode image to get dimensions
		imgInfo, err := image.ProcessImage(imgData)
		if err == nil {
			if w <= 0 && h <= 0 {
				w = float64(imgInfo.WidthPx) * 0.75 // basic scale
				h = float64(imgInfo.HeightPx) * 0.75
			} else if w <= 0 {
				w = h * float64(imgInfo.WidthPx) / float64(imgInfo.HeightPx)
			} else {
				h = w * float64(imgInfo.HeightPx) / float64(imgInfo.WidthPx)
			}
		} else {
			if w <= 0 {
				w = 100
			}
			if h <= 0 {
				h = 100
			}
		}
	}

	// 2. Handle alignment/wrapping
	areaWidth := f.builder.pageWidth(f.pageIndex) - f.opts.Left - f.opts.Right
	x := f.opts.Left

	switch layout.Align {
	case style.ImageAlignCenter:
		x = f.opts.Left + (areaWidth-w)/2
	case style.ImageAlignRight:
		x = f.opts.Left + areaWidth - w
	}

	// Check if we need a page break (for TopBottom or None)
	if layout.Wrap == style.ImageWrapNone || layout.Wrap == style.ImageWrapTopBottom {
		if f.currY-h < f.opts.Bottom {
			f.builder.AddPage()
			f.pageIndex = len(f.builder.pc.pages) - 1
			f.currY = f.builder.pageHeight(f.pageIndex) - f.opts.Top
		}
	}

	// 3. Draw image
	y := f.currY - h
	f.builder.DrawImageWith(ImageOptions{
		Data: imgData, X: x, Y: y, Width: w, Height: h,
		Opacity: s.Opacity, RotateDeg: s.Rotation,
		ClipCircle: s.ClipCircle, ClipCX: s.ClipCX, ClipCY: s.ClipCY, ClipRadius: s.ClipR,
		PageIndex: f.pageIndex,
	})

	// 4. Handle wrapping logic
	if layout.Wrap == style.ImageWrapSquare || layout.Wrap == style.ImageWrapTight {
		f.floating = append(f.floating, FloatingImage{
			X: x, Y: y, Width: w, Height: h, Margin: layout.Margin, Wrap: layout.Wrap,
		})
		// Keep currY where it is for square wrapping
	} else {
		f.currY -= h + layout.Margin
	}

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
		LineRectFn: func(lineIdx int) (float64, float64) {
			y := f.currY - float64(lineIdx)*style.FontSize*1.25
			xOffset := 0.0
			width := f.Width()

			for _, img := range f.floating {
				// Check if this line overlaps with the floating image Y range
				imgTop := img.Y + img.Height
				imgBottom := img.Y
				lineTop := y
				lineBottom := y - style.FontSize

				if lineTop > imgBottom-img.Margin && lineBottom < imgTop+img.Margin {
					// Overlap!
					if img.X < f.Left()+f.Width()/2 {
						// Image is on the left
						overlapX := img.X + img.Width + img.Margin - f.Left()
						if overlapX > xOffset {
							xOffset = overlapX
							width = f.Width() - xOffset
						}
					} else {
						// Image is on the right
						overlapWidth := f.Left() + f.Width() - (img.X - img.Margin)
						if overlapWidth > (f.Width() - width) {
							width = f.Width() - overlapWidth - xOffset
						}
					}
				}
			}
			return xOffset, width
		},
	}

	f.pageIndex, f.currY = f.builder.layoutTextIntoPages(f.pageIndex, text, f.Left(), f.currY, style.FontName, style.FontSize, opts, f.builder.useTagged, model.Name("P"))
	f.syncCursor()
	return f
}

// Space adds vertical space to the flow.
func (f *FlowBuilder) Space(h float64) *FlowBuilder {
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

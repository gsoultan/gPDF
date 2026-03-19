package table

import "gpdf/doc/style"

// CellVerticalAlign controls vertical placement of content inside a table cell.
type CellVerticalAlign string

const (
	CellVAlignTop    CellVerticalAlign = "top"
	CellVAlignMiddle CellVerticalAlign = "middle"
	CellVAlignBottom CellVerticalAlign = "bottom"
)

// CellHorizontalAlign controls horizontal placement of content inside a table cell.
type CellHorizontalAlign string

const (
	CellHAlignLeft    CellHorizontalAlign = "left"
	CellHAlignCenter  CellHorizontalAlign = "center"
	CellHAlignRight   CellHorizontalAlign = "right"
	CellHAlignJustify CellHorizontalAlign = "justify"
)

// CellStyle controls padding and vertical alignment for a table cell.
// Zero value uses sensible defaults (padding 4pt on all sides, top alignment).
type CellStyle struct {
	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	VAlign CellVerticalAlign
	HAlign CellHorizontalAlign

	TextColorRGB [3]float64
	TextColor    style.Color
	HasTextColor bool

	FillColor    style.Color
	HasFillColor bool

	FontName string
	FontSize float64
}

// ResolvedPadding returns the effective padding.
// If a padding value is 0, it defaults to 4.0.
func (s CellStyle) ResolvedPadding() (top, right, bottom, left float64) {
	const defaultPad = 4.0
	top = s.PaddingTop
	if top == 0 {
		top = defaultPad
	}
	right = s.PaddingRight
	if right == 0 {
		right = defaultPad
	}
	bottom = s.PaddingBottom
	if bottom == 0 {
		bottom = defaultPad
	}
	left = s.PaddingLeft
	if left == 0 {
		left = defaultPad
	}
	return
}

// ResolvedFont returns the effective font name and size.
// Defaults to Helvetica 10pt if not specified.
func (s CellStyle) ResolvedFont() (name string, size float64) {
	name = s.FontName
	if name == "" {
		name = "Helvetica"
	}
	size = s.FontSize
	if size <= 0 {
		size = 10.0
	}
	return
}

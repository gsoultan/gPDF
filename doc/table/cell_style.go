package table

import "gpdf/doc/style"

// CellVerticalAlign controls vertical placement of content inside a table cell.
type CellVerticalAlign string

const (
	CellVAlignTop    CellVerticalAlign = "top"
	CellVAlignMiddle CellVerticalAlign = "middle"
	CellVAlignBottom CellVerticalAlign = "bottom"
)

// CellStyle controls padding and vertical alignment for a table cell.
// Zero value uses sensible defaults (padding 4pt on all sides, top alignment).
type CellStyle struct {
	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	VAlign CellVerticalAlign

	TextColorRGB [3]float64

	FillColor    style.Color
	HasFillColor bool
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

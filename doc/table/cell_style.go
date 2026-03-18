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

// ResolvedPadding returns the effective padding, defaulting to 4pt on all sides.
func (s CellStyle) ResolvedPadding() (top, right, bottom, left float64) {
	const defaultPad = 4.0
	if s.PaddingTop == 0 && s.PaddingRight == 0 && s.PaddingBottom == 0 && s.PaddingLeft == 0 {
		return defaultPad, defaultPad, defaultPad, defaultPad
	}
	return s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft
}

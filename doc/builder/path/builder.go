package path

import "gpdf/doc/style"

// Builder accumulates path construction operators and commits the
// finished path as a graphic run on a page.
// Chain path-construction methods (MoveTo, LineTo, …), then call a
// commit method (Stroke, Fill, …) to flush the operators.
type Builder interface {
	MoveTo(x, y float64) Builder
	LineTo(x, y float64) Builder
	CurveTo(x1, y1, x2, y2, x3, y3 float64) Builder
	Rect(x, y, w, h float64) Builder
	ClosePath() Builder
	Arc(cx, cy, rx, ry, startDeg, sweepDeg float64) Builder
	RoundedRect(x, y, w, h, r float64) Builder

	Stroke(ls style.LineStyle)
	Fill(c style.Color)
	FillStroke(fill style.Color, stroke style.LineStyle)
	EndPath()

	StrokeWithState(ls style.LineStyle, gs style.GraphicsState)
	FillWithState(c style.Color, gs style.GraphicsState)
	FillStrokeWithState(fill style.Color, stroke style.LineStyle, gs style.GraphicsState)
}

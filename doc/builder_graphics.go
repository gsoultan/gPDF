package doc

import (
	"fmt"

	"gpdf/content"
	"gpdf/doc/builder"
	"gpdf/model"
)

type graphicRun = builder.GraphicRun

// addGraphicRunWithState wraps body ops in q/Q and optionally inserts a gs operator for the graphics state.
func (b *DocumentBuilder) addGraphicRunWithState(pageIndex int, state GraphicsState, body []content.Op) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ps := &b.pc.pages[pageIndex]
	ops := make([]content.Op, 0, len(body)+4)
	ops = append(ops, content.Op{Name: "q"})

	var extGStates map[model.Name]model.Dict
	if !state.IsDefault() {
		gsName := model.Name(fmt.Sprintf("GS%d", ps.NextGSIndex+1))
		ps.NextGSIndex++
		ops = append(ops, content.Op{
			Name: "gs",
			Args: []model.Object{gsName},
		})
		extGStates = map[model.Name]model.Dict{gsName: state.ExtGStateDict()}
	}

	ops = append(ops, body...)
	ops = append(ops, content.Op{Name: "Q"})
	ps.GraphicRuns = append(ps.GraphicRuns, graphicRun{Ops: ops, ExtGStates: extGStates})
	return b
}

// BeginPath starts a custom path on the given page.
// Use MoveTo, LineTo, CurveTo, Rect, ClosePath to construct the path,
// then Stroke, Fill, FillStroke, or EndPath to finish.
func (b *DocumentBuilder) BeginPath(pageIndex int) *PathBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return nil
	}
	return &PathBuilder{
		builder:   b,
		pageIndex: pageIndex,
	}
}

// DrawLine draws a straight line from (x1, y1) to (x2, y2) on the given page.
func (b *DocumentBuilder) DrawLine(pageIndex int, x1, y1, x2, y2 float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{
		{Name: "q"},
	}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops,
		content.Op{Name: "m", Args: []model.Object{model.Real(x1), model.Real(y1)}},
		content.Op{Name: "l", Args: []model.Object{model.Real(x2), model.Real(y2)}},
		content.Op{Name: "S"},
		content.Op{Name: "Q"},
	)
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// DrawRect draws a stroked rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) DrawRect(pageIndex int, x, y, width, height float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{
		{Name: "q"},
	}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "S"},
		content.Op{Name: "Q"},
	)
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillRect draws a filled rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) FillRect(pageIndex int, x, y, width, height float64, fill Color) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{
		{Name: "q"},
	}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "f"},
		content.Op{Name: "Q"},
	)
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillStrokeRect draws a filled and stroked rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) FillStrokeRect(pageIndex int, x, y, width, height float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{
		{Name: "q"},
	}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "B"},
		content.Op{Name: "Q"},
	)
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// DrawCircle draws a stroked circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) DrawCircle(pageIndex int, cx, cy, radius float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillCircle draws a filled circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) FillCircle(pageIndex int, cx, cy, radius float64, fill Color) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillStrokeCircle draws a filled and stroked circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) FillStrokeCircle(pageIndex int, cx, cy, radius float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// DrawLineWithState draws a line with the given graphics state (e.g. opacity/blend mode).
func (b *DocumentBuilder) DrawLineWithState(pageIndex int, x1, y1, x2, y2 float64, style LineStyle, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, strokeStateOps(style)...)
	body = append(body,
		content.Op{Name: "m", Args: []model.Object{model.Real(x1), model.Real(y1)}},
		content.Op{Name: "l", Args: []model.Object{model.Real(x2), model.Real(y2)}},
		content.Op{Name: "S"},
	)
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// DrawRectWithState draws a stroked rectangle with the given graphics state.
func (b *DocumentBuilder) DrawRectWithState(pageIndex int, x, y, width, height float64, style LineStyle, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, strokeStateOps(style)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "S"},
	)
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// FillRectWithState draws a filled rectangle with the given graphics state.
func (b *DocumentBuilder) FillRectWithState(pageIndex int, x, y, width, height float64, fill Color, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, fillColorOps(fill)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "f"},
	)
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// FillStrokeRectWithState draws a filled and stroked rectangle with the given graphics state.
func (b *DocumentBuilder) FillStrokeRectWithState(pageIndex int, x, y, width, height float64, fill Color, stroke LineStyle, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, strokeStateOps(stroke)...)
	body = append(body, fillColorOps(fill)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "B"},
	)
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// DrawCircleWithState draws a stroked circle with the given graphics state.
func (b *DocumentBuilder) DrawCircleWithState(pageIndex int, cx, cy, radius float64, style LineStyle, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, strokeStateOps(style)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "S"})
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// FillCircleWithState draws a filled circle with the given graphics state.
func (b *DocumentBuilder) FillCircleWithState(pageIndex int, cx, cy, radius float64, fill Color, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, fillColorOps(fill)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "f"})
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// FillStrokeCircleWithState draws a filled and stroked circle with the given graphics state.
func (b *DocumentBuilder) FillStrokeCircleWithState(pageIndex int, cx, cy, radius float64, fill Color, stroke LineStyle, state GraphicsState) *DocumentBuilder {
	var body []content.Op
	body = append(body, strokeStateOps(stroke)...)
	body = append(body, fillColorOps(fill)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "B"})
	return b.addGraphicRunWithState(pageIndex, state, body)
}

// ── Rounded Rectangle ───────────────────────────────────────────────────────

// DrawRoundedRect strokes a rectangle with rounded corners.
// radius is the corner radius in points; clamped to min(width,height)/2.
func (b *DocumentBuilder) DrawRoundedRect(pageIndex int, x, y, width, height, radius float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillRoundedRect fills a rectangle with rounded corners.
func (b *DocumentBuilder) FillRoundedRect(pageIndex int, x, y, width, height, radius float64, fill Color) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillStrokeRoundedRect fills and strokes a rectangle with rounded corners.
func (b *DocumentBuilder) FillStrokeRoundedRect(pageIndex int, x, y, width, height, radius float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// ── Ellipse ──────────────────────────────────────────────────────────────────

// DrawEllipse strokes an axis-aligned ellipse centered at (cx, cy) with semi-axes rx and ry.
func (b *DocumentBuilder) DrawEllipse(pageIndex int, cx, cy, rx, ry float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillEllipse fills an axis-aligned ellipse centered at (cx, cy) with semi-axes rx and ry.
func (b *DocumentBuilder) FillEllipse(pageIndex int, cx, cy, rx, ry float64, fill Color) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillStrokeEllipse fills and strokes an axis-aligned ellipse centered at (cx, cy) with semi-axes rx and ry.
func (b *DocumentBuilder) FillStrokeEllipse(pageIndex int, cx, cy, rx, ry float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// ── Polygon ───────────────────────────────────────────────────────────────────

// DrawPolygon strokes a closed polygon defined by a list of (x,y) point pairs.
// points must have an even number of elements: [x0,y0, x1,y1, ...].
func (b *DocumentBuilder) DrawPolygon(pageIndex int, points []float64, style LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillPolygon fills a closed polygon defined by a list of (x,y) point pairs.
func (b *DocumentBuilder) FillPolygon(pageIndex int, points []float64, fill Color) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// FillStrokePolygon fills and strokes a closed polygon defined by a list of (x,y) point pairs.
func (b *DocumentBuilder) FillStrokePolygon(pageIndex int, points []float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	b.pc.pages[pageIndex].GraphicRuns = append(b.pc.pages[pageIndex].GraphicRuns, graphicRun{Ops: ops})
	return b
}

// circleControlFactor is 4*(sqrt(2)-1)/3, the Bézier control-point offset for a quarter circle.
const circleControlFactor = 0.5522847498

// circlePathOps returns path-construction operators that approximate a circle
// centered at (cx, cy) with radius r using four cubic Bézier curves.
func circlePathOps(cx, cy, r float64) []content.Op {
	k := r * circleControlFactor
	return []content.Op{
		{Name: "m", Args: []model.Object{model.Real(cx + r), model.Real(cy)}},
		{Name: "c", Args: []model.Object{
			model.Real(cx + r), model.Real(cy + k),
			model.Real(cx + k), model.Real(cy + r),
			model.Real(cx), model.Real(cy + r),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx - k), model.Real(cy + r),
			model.Real(cx - r), model.Real(cy + k),
			model.Real(cx - r), model.Real(cy),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx - r), model.Real(cy - k),
			model.Real(cx - k), model.Real(cy - r),
			model.Real(cx), model.Real(cy - r),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx + k), model.Real(cy - r),
			model.Real(cx + r), model.Real(cy - k),
			model.Real(cx + r), model.Real(cy),
		}},
		{Name: "h"},
	}
}

// ellipsePathOps returns path ops for an axis-aligned ellipse centered at (cx,cy)
// with horizontal semi-axis rx and vertical semi-axis ry.
func ellipsePathOps(cx, cy, rx, ry float64) []content.Op {
	kx := rx * circleControlFactor
	ky := ry * circleControlFactor
	return []content.Op{
		{Name: "m", Args: []model.Object{model.Real(cx + rx), model.Real(cy)}},
		{Name: "c", Args: []model.Object{
			model.Real(cx + rx), model.Real(cy + ky),
			model.Real(cx + kx), model.Real(cy + ry),
			model.Real(cx), model.Real(cy + ry),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx - kx), model.Real(cy + ry),
			model.Real(cx - rx), model.Real(cy + ky),
			model.Real(cx - rx), model.Real(cy),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx - rx), model.Real(cy - ky),
			model.Real(cx - kx), model.Real(cy - ry),
			model.Real(cx), model.Real(cy - ry),
		}},
		{Name: "c", Args: []model.Object{
			model.Real(cx + kx), model.Real(cy - ry),
			model.Real(cx + rx), model.Real(cy - ky),
			model.Real(cx + rx), model.Real(cy),
		}},
		{Name: "h"},
	}
}

// roundedRectPathOps returns path ops for a rectangle with rounded corners.
// radius is clamped to min(width,height)/2.
func roundedRectPathOps(x, y, w, h, r float64) []content.Op {
	maxR := min(w, h) / 2
	if r > maxR {
		r = maxR
	}
	if r <= 0 {
		return []content.Op{
			{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(w), model.Real(h)}},
		}
	}
	k := r * circleControlFactor
	// Start at bottom-left corner arc end, go clockwise.
	return []content.Op{
		{Name: "m", Args: []model.Object{model.Real(x + r), model.Real(y)}},
		// bottom edge → bottom-right arc
		{Name: "l", Args: []model.Object{model.Real(x + w - r), model.Real(y)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + w - r + k), model.Real(y),
			model.Real(x + w), model.Real(y + r - k),
			model.Real(x + w), model.Real(y + r),
		}},
		// right edge → top-right arc
		{Name: "l", Args: []model.Object{model.Real(x + w), model.Real(y + h - r)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + w), model.Real(y + h - r + k),
			model.Real(x + w - r + k), model.Real(y + h),
			model.Real(x + w - r), model.Real(y + h),
		}},
		// top edge → top-left arc
		{Name: "l", Args: []model.Object{model.Real(x + r), model.Real(y + h)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + r - k), model.Real(y + h),
			model.Real(x), model.Real(y + h - r + k),
			model.Real(x), model.Real(y + h - r),
		}},
		// left edge → bottom-left arc
		{Name: "l", Args: []model.Object{model.Real(x), model.Real(y + r)}},
		{Name: "c", Args: []model.Object{
			model.Real(x), model.Real(y + r - k),
			model.Real(x + r - k), model.Real(y),
			model.Real(x + r), model.Real(y),
		}},
		{Name: "h"},
	}
}

// polygonPathOps returns path ops for a closed polygon.
// points is a flat slice [x0,y0, x1,y1, ...].
func polygonPathOps(points []float64) []content.Op {
	ops := []content.Op{
		{Name: "m", Args: []model.Object{model.Real(points[0]), model.Real(points[1])}},
	}
	for i := 2; i+1 < len(points); i += 2 {
		ops = append(ops, content.Op{
			Name: "l",
			Args: []model.Object{model.Real(points[i]), model.Real(points[i+1])},
		})
	}
	ops = append(ops, content.Op{Name: "h"})
	return ops
}

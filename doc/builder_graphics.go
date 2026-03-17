package doc

import (
	"fmt"

	"gpdf/content"
	"gpdf/model"
)

// graphicRun holds pre-built content stream operators for one vector drawing operation.
// When extGStates is non-nil, its entries must be added to the page's /Resources /ExtGState dict.
type graphicRun struct {
	ops        []content.Op
	extGStates map[model.Name]model.Dict
}

// addGraphicRunWithState wraps body ops in q/Q and optionally inserts a gs operator for the graphics state.
func (b *DocumentBuilder) addGraphicRunWithState(pageIndex int, state GraphicsState, body []content.Op) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return b
	}
	ps := &b.pages[pageIndex]
	ops := make([]content.Op, 0, len(body)+4)
	ops = append(ops, content.Op{Name: "q"})

	var extGStates map[model.Name]model.Dict
	if !state.isDefault() {
		gsName := model.Name(fmt.Sprintf("GS%d", ps.nextGSIndex+1))
		ps.nextGSIndex++
		ops = append(ops, content.Op{
			Name: "gs",
			Args: []model.Object{gsName},
		})
		extGStates = map[model.Name]model.Dict{gsName: state.extGStateDict()}
	}

	ops = append(ops, body...)
	ops = append(ops, content.Op{Name: "Q"})
	ps.graphicRuns = append(ps.graphicRuns, graphicRun{ops: ops, extGStates: extGStates})
	return b
}

// BeginPath starts a custom path on the given page.
// Use MoveTo, LineTo, CurveTo, Rect, ClosePath to construct the path,
// then Stroke, Fill, FillStroke, or EndPath to finish.
func (b *DocumentBuilder) BeginPath(pageIndex int) *PathBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return nil
	}
	return &PathBuilder{
		builder:   b,
		pageIndex: pageIndex,
	}
}

// DrawLine draws a straight line from (x1, y1) to (x2, y2) on the given page.
func (b *DocumentBuilder) DrawLine(pageIndex int, x1, y1, x2, y2 float64, style LineStyle) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
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
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// DrawRect draws a stroked rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) DrawRect(pageIndex int, x, y, width, height float64, style LineStyle) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
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
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// FillRect draws a filled rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) FillRect(pageIndex int, x, y, width, height float64, fill Color) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
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
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// FillStrokeRect draws a filled and stroked rectangle at (x, y) with dimensions width × height.
func (b *DocumentBuilder) FillStrokeRect(pageIndex int, x, y, width, height float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
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
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// DrawCircle draws a stroked circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) DrawCircle(pageIndex int, cx, cy, radius float64, style LineStyle) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// FillCircle draws a filled circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) FillCircle(pageIndex int, cx, cy, radius float64, fill Color) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
	return b
}

// FillStrokeCircle draws a filled and stroked circle centered at (cx, cy) with the given radius.
func (b *DocumentBuilder) FillStrokeCircle(pageIndex int, cx, cy, radius float64, fill Color, stroke LineStyle) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return b
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	b.pages[pageIndex].graphicRuns = append(b.pages[pageIndex].graphicRuns, graphicRun{ops: ops})
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

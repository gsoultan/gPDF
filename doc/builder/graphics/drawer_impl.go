package graphics

import (
	"fmt"

	"gpdf/content"
	"gpdf/doc/builder"
	"gpdf/doc/style"
	"gpdf/model"
)

const circleControlFactor = 0.5522847498

type drawer struct{}

// NewDrawer returns a Drawer that renders vector shapes onto PDF pages.
func NewDrawer() Drawer { return drawer{} }

// ── Lines ────────────────────────────────────────────────────────────────────

func (d drawer) DrawLine(pa builder.PageAccess, pageIndex int, x1, y1, x2, y2 float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops,
		content.Op{Name: "m", Args: []model.Object{model.Real(x1), model.Real(y1)}},
		content.Op{Name: "l", Args: []model.Object{model.Real(x2), model.Real(y2)}},
		content.Op{Name: "S"},
		content.Op{Name: "Q"},
	)
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) DrawLineWithState(pa builder.PageAccess, pageIndex int, x1, y1, x2, y2 float64, ls style.LineStyle, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, strokeStateOps(ls)...)
	body = append(body,
		content.Op{Name: "m", Args: []model.Object{model.Real(x1), model.Real(y1)}},
		content.Op{Name: "l", Args: []model.Object{model.Real(x2), model.Real(y2)}},
		content.Op{Name: "S"},
	)
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

// ── Rectangles ───────────────────────────────────────────────────────────────

func (d drawer) DrawRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "S"},
		content.Op{Name: "Q"},
	)
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, c style.Color) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "f"},
		content.Op{Name: "Q"},
	)
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillStrokeRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, stroke style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "B"},
		content.Op{Name: "Q"},
	)
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) DrawRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, ls style.LineStyle, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, strokeStateOps(ls)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "S"},
	)
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

func (d drawer) FillRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, fillColorOps(fill)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "f"},
	)
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

func (d drawer) FillStrokeRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, stroke style.LineStyle, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, strokeStateOps(stroke)...)
	body = append(body, fillColorOps(fill)...)
	body = append(body,
		content.Op{Name: "re", Args: []model.Object{model.Real(x), model.Real(y), model.Real(width), model.Real(height)}},
		content.Op{Name: "B"},
	)
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

// ── Circles ──────────────────────────────────────────────────────────────────

func (d drawer) DrawCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, c style.Color) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillStrokeCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, stroke style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, circlePathOps(cx, cy, radius)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) DrawCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, ls style.LineStyle, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, strokeStateOps(ls)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "S"})
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

func (d drawer) FillCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, fillColorOps(fill)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "f"})
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

func (d drawer) FillStrokeCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, stroke style.LineStyle, gs style.GraphicsState) {
	var body []content.Op
	body = append(body, strokeStateOps(stroke)...)
	body = append(body, fillColorOps(fill)...)
	body = append(body, circlePathOps(cx, cy, radius)...)
	body = append(body, content.Op{Name: "B"})
	addGraphicRunWithState(pa, pageIndex, gs, body)
}

// ── Rounded Rectangles ───────────────────────────────────────────────────────

func (d drawer) DrawRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, c style.Color) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillStrokeRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, fill style.Color, stroke style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, roundedRectPathOps(x, y, width, height, radius)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

// ── Ellipses ─────────────────────────────────────────────────────────────────

func (d drawer) DrawEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, c style.Color) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillStrokeEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, fill style.Color, stroke style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, ellipsePathOps(cx, cy, rx, ry)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

// ── Polygons ─────────────────────────────────────────────────────────────────

func (d drawer) DrawPolygon(pa builder.PageAccess, pageIndex int, points []float64, ls style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "S"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillPolygon(pa builder.PageAccess, pageIndex int, points []float64, c style.Color) {
	if !pa.ValidPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "f"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

func (d drawer) FillStrokePolygon(pa builder.PageAccess, pageIndex int, points []float64, fill style.Color, stroke style.LineStyle) {
	if !pa.ValidPageIndex(pageIndex) || len(points) < 4 || len(points)%2 != 0 {
		return
	}
	ops := []content.Op{{Name: "q"}}
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, polygonPathOps(points)...)
	ops = append(ops, content.Op{Name: "B"}, content.Op{Name: "Q"})
	pa.PageAt(pageIndex).GraphicRuns = append(pa.PageAt(pageIndex).GraphicRuns, builder.GraphicRun{Ops: ops})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func addGraphicRunWithState(pa builder.PageAccess, pageIndex int, state style.GraphicsState, body []content.Op) {
	if !pa.ValidPageIndex(pageIndex) {
		return
	}
	ps := pa.PageAt(pageIndex)
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
	ps.GraphicRuns = append(ps.GraphicRuns, builder.GraphicRun{Ops: ops, ExtGStates: extGStates})
}

func strokeStateOps(ls style.LineStyle) []content.Op {
	var ops []content.Op
	ops = append(ops, content.Op{
		Name: "w",
		Args: []model.Object{model.Real(ls.ResolvedWidth())},
	})
	if ls.Cap != style.LineCapButt {
		ops = append(ops, content.Op{
			Name: "J",
			Args: []model.Object{model.Integer(int64(ls.Cap))},
		})
	}
	if ls.Join != style.LineJoinMiter {
		ops = append(ops, content.Op{
			Name: "j",
			Args: []model.Object{model.Integer(int64(ls.Join))},
		})
	}
	if len(ls.DashArray) > 0 {
		arr := make(model.Array, len(ls.DashArray))
		for i, d := range ls.DashArray {
			arr[i] = model.Real(d)
		}
		ops = append(ops, content.Op{
			Name: "d",
			Args: []model.Object{arr, model.Real(ls.DashPhase)},
		})
	}
	ops = append(ops, content.Op{
		Name: "RG",
		Args: []model.Object{
			model.Real(ls.Color.R),
			model.Real(ls.Color.G),
			model.Real(ls.Color.B),
		},
	})
	return ops
}

func fillColorOps(c style.Color) []content.Op {
	return []content.Op{
		{
			Name: "rg",
			Args: []model.Object{
				model.Real(c.R),
				model.Real(c.G),
				model.Real(c.B),
			},
		},
	}
}

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
	return []content.Op{
		{Name: "m", Args: []model.Object{model.Real(x + r), model.Real(y)}},
		{Name: "l", Args: []model.Object{model.Real(x + w - r), model.Real(y)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + w - r + k), model.Real(y),
			model.Real(x + w), model.Real(y + r - k),
			model.Real(x + w), model.Real(y + r),
		}},
		{Name: "l", Args: []model.Object{model.Real(x + w), model.Real(y + h - r)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + w), model.Real(y + h - r + k),
			model.Real(x + w - r + k), model.Real(y + h),
			model.Real(x + w - r), model.Real(y + h),
		}},
		{Name: "l", Args: []model.Object{model.Real(x + r), model.Real(y + h)}},
		{Name: "c", Args: []model.Object{
			model.Real(x + r - k), model.Real(y + h),
			model.Real(x), model.Real(y + h - r + k),
			model.Real(x), model.Real(y + h - r),
		}},
		{Name: "l", Args: []model.Object{model.Real(x), model.Real(y + r)}},
		{Name: "c", Args: []model.Object{
			model.Real(x), model.Real(y + r - k),
			model.Real(x + r - k), model.Real(y),
			model.Real(x + r), model.Real(y),
		}},
		{Name: "h"},
	}
}

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

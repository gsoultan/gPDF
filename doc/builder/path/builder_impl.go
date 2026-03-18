package path

import (
	"fmt"
	"math"

	"gpdf/content"
	"gpdf/doc/builder"
	"gpdf/doc/style"
	"gpdf/model"
)

const circleControlFactor = 0.5522847498

type pathBuilder struct {
	pa        builder.PageAccess
	pageIndex int
	pathOps   []content.Op
}

// NewBuilder returns a Builder that accumulates path operators for the
// given page and commits them as a GraphicRun via pa.
func NewBuilder(pa builder.PageAccess, pageIndex int) Builder {
	return &pathBuilder{
		pa:        pa,
		pageIndex: pageIndex,
	}
}

// ── Path construction ────────────────────────────────────────────────────────

func (p *pathBuilder) MoveTo(x, y float64) Builder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "m",
		Args: []model.Object{model.Real(x), model.Real(y)},
	})
	return p
}

func (p *pathBuilder) LineTo(x, y float64) Builder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "l",
		Args: []model.Object{model.Real(x), model.Real(y)},
	})
	return p
}

func (p *pathBuilder) CurveTo(x1, y1, x2, y2, x3, y3 float64) Builder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "c",
		Args: []model.Object{
			model.Real(x1), model.Real(y1),
			model.Real(x2), model.Real(y2),
			model.Real(x3), model.Real(y3),
		},
	})
	return p
}

func (p *pathBuilder) Rect(x, y, w, h float64) Builder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "re",
		Args: []model.Object{model.Real(x), model.Real(y), model.Real(w), model.Real(h)},
	})
	return p
}

func (p *pathBuilder) ClosePath() Builder {
	p.pathOps = append(p.pathOps, content.Op{Name: "h"})
	return p
}

func (p *pathBuilder) Arc(cx, cy, rx, ry, startDeg, sweepDeg float64) Builder {
	if rx <= 0 || ry <= 0 || sweepDeg == 0 {
		return p
	}
	steps := int(math.Ceil(math.Abs(sweepDeg) / 90.0))
	segSweep := sweepDeg / float64(steps)
	angle := startDeg
	for range steps {
		a0 := angle * math.Pi / 180
		a1 := (angle + segSweep) * math.Pi / 180
		da := a1 - a0
		alpha := math.Sin(da) * (math.Sqrt(4+3*math.Pow(math.Tan(da/2), 2)) - 1) / 3

		cos0, sin0 := math.Cos(a0), math.Sin(a0)
		cos1, sin1 := math.Cos(a1), math.Sin(a1)

		x0 := cx + rx*cos0
		y0 := cy + ry*sin0
		x3 := cx + rx*cos1
		y3 := cy + ry*sin1
		x1 := x0 + alpha*(-rx*sin0)
		y1 := y0 + alpha*(ry*cos0)
		x2 := x3 - alpha*(-rx*sin1)
		y2 := y3 - alpha*(ry*cos1)

		if angle == startDeg {
			p.pathOps = append(p.pathOps, content.Op{
				Name: "m",
				Args: []model.Object{model.Real(x0), model.Real(y0)},
			})
		}
		p.pathOps = append(p.pathOps, content.Op{
			Name: "c",
			Args: []model.Object{
				model.Real(x1), model.Real(y1),
				model.Real(x2), model.Real(y2),
				model.Real(x3), model.Real(y3),
			},
		})
		angle += segSweep
	}
	return p
}

func (p *pathBuilder) RoundedRect(x, y, w, h, r float64) Builder {
	p.pathOps = append(p.pathOps, roundedRectPathOps(x, y, w, h, r)...)
	return p
}

// ── Commit methods ───────────────────────────────────────────────────────────

func (p *pathBuilder) Stroke(ls style.LineStyle) {
	var ops []content.Op
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "S"})
	p.commit(ops)
}

func (p *pathBuilder) Fill(c style.Color) {
	var ops []content.Op
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "f"})
	p.commit(ops)
}

func (p *pathBuilder) FillStroke(fill style.Color, stroke style.LineStyle) {
	var ops []content.Op
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "B"})
	p.commit(ops)
}

func (p *pathBuilder) EndPath() {
	ops := make([]content.Op, len(p.pathOps), len(p.pathOps)+1)
	copy(ops, p.pathOps)
	ops = append(ops, content.Op{Name: "n"})
	p.commit(ops)
}

func (p *pathBuilder) StrokeWithState(ls style.LineStyle, gs style.GraphicsState) {
	var ops []content.Op
	ops = append(ops, strokeStateOps(ls)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "S"})
	p.commitWithState(ops, &gs)
}

func (p *pathBuilder) FillWithState(c style.Color, gs style.GraphicsState) {
	var ops []content.Op
	ops = append(ops, fillColorOps(c)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "f"})
	p.commitWithState(ops, &gs)
}

func (p *pathBuilder) FillStrokeWithState(fill style.Color, stroke style.LineStyle, gs style.GraphicsState) {
	var ops []content.Op
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "B"})
	p.commitWithState(ops, &gs)
}

// ── Internal helpers ─────────────────────────────────────────────────────────

func (p *pathBuilder) commit(ops []content.Op) {
	p.commitWithState(ops, nil)
}

func (p *pathBuilder) commitWithState(ops []content.Op, state *style.GraphicsState) {
	ps := p.pa.PageAt(p.pageIndex)

	wrapped := make([]content.Op, 0, len(ops)+4)
	wrapped = append(wrapped, content.Op{Name: "q"})

	var extGStates map[model.Name]model.Dict
	if state != nil && !state.IsDefault() {
		gsName := model.Name(fmt.Sprintf("GS%d", ps.NextGSIndex+1))
		ps.NextGSIndex++
		wrapped = append(wrapped, content.Op{
			Name: "gs",
			Args: []model.Object{gsName},
		})
		extGStates = map[model.Name]model.Dict{gsName: state.ExtGStateDict()}
	}

	wrapped = append(wrapped, ops...)
	wrapped = append(wrapped, content.Op{Name: "Q"})
	ps.GraphicRuns = append(ps.GraphicRuns, builder.GraphicRun{Ops: wrapped, ExtGStates: extGStates})
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

package doc

import (
	"fmt"
	"math"

	"github.com/gsoultan/gpdf/content"
	"github.com/gsoultan/gpdf/model"
)

// PathBuilder accumulates path construction operators for a single drawing on a page.
// Finish with Stroke, Fill, FillStroke, or EndPath to apply the path and return to the DocumentBuilder.
type PathBuilder struct {
	builder   *DocumentBuilder
	pageIndex int
	pathOps   []content.Op
}

// MoveTo begins a new subpath at (x, y).
func (p *PathBuilder) MoveTo(x, y float64) *PathBuilder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "m",
		Args: []model.Object{model.Real(x), model.Real(y)},
	})
	return p
}

// LineTo appends a straight line from the current point to (x, y).
func (p *PathBuilder) LineTo(x, y float64) *PathBuilder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "l",
		Args: []model.Object{model.Real(x), model.Real(y)},
	})
	return p
}

// CurveTo appends a cubic Bézier curve from the current point to (x3, y3)
// using (x1, y1) and (x2, y2) as control points.
func (p *PathBuilder) CurveTo(x1, y1, x2, y2, x3, y3 float64) *PathBuilder {
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

// Rect appends a rectangle subpath at (x, y) with dimensions w × h.
func (p *PathBuilder) Rect(x, y, w, h float64) *PathBuilder {
	p.pathOps = append(p.pathOps, content.Op{
		Name: "re",
		Args: []model.Object{model.Real(x), model.Real(y), model.Real(w), model.Real(h)},
	})
	return p
}

// ClosePath closes the current subpath with a straight line back to its starting point.
func (p *PathBuilder) ClosePath() *PathBuilder {
	p.pathOps = append(p.pathOps, content.Op{Name: "h"})
	return p
}

// Arc appends an elliptical arc centered at (cx, cy) with semi-axes rx and ry.
// startDeg is the start angle in degrees (0 = rightmost point), sweepDeg is the
// angular extent (positive = counter-clockwise in PDF user space).
// The arc is approximated with cubic Bézier curves (max 90° per segment).
func (p *PathBuilder) Arc(cx, cy, rx, ry, startDeg, sweepDeg float64) *PathBuilder {
	if rx <= 0 || ry <= 0 || sweepDeg == 0 {
		return p
	}
	// Break into ≤90° segments.
	steps := int(math.Ceil(math.Abs(sweepDeg) / 90.0))
	segSweep := sweepDeg / float64(steps)
	angle := startDeg
	for range steps {
		a0 := angle * math.Pi / 180
		a1 := (angle + segSweep) * math.Pi / 180
		da := a1 - a0
		// Bézier control-point length for the arc segment.
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

// RoundedRect appends a closed rounded-rectangle subpath at (x,y) with size w×h
// and corner radius r.
func (p *PathBuilder) RoundedRect(x, y, w, h, r float64) *PathBuilder {
	p.pathOps = append(p.pathOps, roundedRectPathOps(x, y, w, h, r)...)
	return p
}

// Stroke paints the path with the given line style and returns to the DocumentBuilder.
func (p *PathBuilder) Stroke(style LineStyle) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "S"})
	return p.commit(ops)
}

// Fill paints the interior of the path with the given color and returns to the DocumentBuilder.
func (p *PathBuilder) Fill(color Color) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, fillColorOps(color)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "f"})
	return p.commit(ops)
}

// FillStroke paints the interior and strokes the outline of the path.
func (p *PathBuilder) FillStroke(fill Color, stroke LineStyle) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "B"})
	return p.commit(ops)
}

// EndPath discards the path without painting (useful after clipping setup).
func (p *PathBuilder) EndPath() *DocumentBuilder {
	ops := make([]content.Op, len(p.pathOps), len(p.pathOps)+1)
	copy(ops, p.pathOps)
	ops = append(ops, content.Op{Name: "n"})
	return p.commit(ops)
}

// StrokeWithState paints the path with the given line style and graphics state.
func (p *PathBuilder) StrokeWithState(style LineStyle, state GraphicsState) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, strokeStateOps(style)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "S"})
	return p.commitWithState(ops, &state)
}

// FillWithState paints the interior with the given color and graphics state.
func (p *PathBuilder) FillWithState(color Color, state GraphicsState) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, fillColorOps(color)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "f"})
	return p.commitWithState(ops, &state)
}

// FillStrokeWithState paints the interior and strokes the outline with the given graphics state.
func (p *PathBuilder) FillStrokeWithState(fill Color, stroke LineStyle, state GraphicsState) *DocumentBuilder {
	var ops []content.Op
	ops = append(ops, strokeStateOps(stroke)...)
	ops = append(ops, fillColorOps(fill)...)
	ops = append(ops, p.pathOps...)
	ops = append(ops, content.Op{Name: "B"})
	return p.commitWithState(ops, &state)
}

func (p *PathBuilder) commit(ops []content.Op) *DocumentBuilder {
	return p.commitWithState(ops, nil)
}

func (p *PathBuilder) commitWithState(ops []content.Op, state *GraphicsState) *DocumentBuilder {
	wrapped := make([]content.Op, 0, len(ops)+4)
	wrapped = append(wrapped, content.Op{Name: "q"})

	var extGStates map[model.Name]model.Dict
	if state != nil && !state.IsDefault() {
		ps := &p.builder.pc.pages[p.pageIndex]
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
	ps := &p.builder.pc.pages[p.pageIndex]
	ps.GraphicRuns = append(ps.GraphicRuns, graphicRun{Ops: wrapped, ExtGStates: extGStates})
	return p.builder
}

// strokeStateOps builds content operators that configure stroke graphics state.
func strokeStateOps(style LineStyle) []content.Op {
	var ops []content.Op
	ops = append(ops, content.Op{
		Name: "w",
		Args: []model.Object{model.Real(style.ResolvedWidth())},
	})
	if style.Cap != LineCapButt {
		ops = append(ops, content.Op{
			Name: "J",
			Args: []model.Object{model.Integer(int64(style.Cap))},
		})
	}
	if style.Join != LineJoinMiter {
		ops = append(ops, content.Op{
			Name: "j",
			Args: []model.Object{model.Integer(int64(style.Join))},
		})
	}
	if len(style.DashArray) > 0 {
		arr := make(model.Array, len(style.DashArray))
		for i, d := range style.DashArray {
			arr[i] = model.Real(d)
		}
		ops = append(ops, content.Op{
			Name: "d",
			Args: []model.Object{arr, model.Real(style.DashPhase)},
		})
	}
	ops = append(ops, content.Op{
		Name: "RG",
		Args: []model.Object{
			model.Real(style.Color.R),
			model.Real(style.Color.G),
			model.Real(style.Color.B),
		},
	})
	return ops
}

// fillColorOps builds a content operator that sets the non-stroking fill color.
func fillColorOps(color Color) []content.Op {
	return []content.Op{
		{
			Name: "rg",
			Args: []model.Object{
				model.Real(color.R),
				model.Real(color.G),
				model.Real(color.B),
			},
		},
	}
}

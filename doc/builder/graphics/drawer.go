package graphics

import (
	"github.com/gsoultan/gpdf/doc/builder"
	"github.com/gsoultan/gpdf/doc/style"
)

// Drawer draws vector shapes onto PDF pages via PageAccess.
type Drawer interface {
	DrawLine(pa builder.PageAccess, pageIndex int, x1, y1, x2, y2 float64, ls style.LineStyle)
	DrawRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, ls style.LineStyle)
	FillRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, c style.Color)
	FillStrokeRect(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, stroke style.LineStyle)

	DrawCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, ls style.LineStyle)
	FillCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, c style.Color)
	FillStrokeCircle(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, stroke style.LineStyle)

	DrawLineWithState(pa builder.PageAccess, pageIndex int, x1, y1, x2, y2 float64, ls style.LineStyle, gs style.GraphicsState)
	DrawRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, ls style.LineStyle, gs style.GraphicsState)
	FillRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, gs style.GraphicsState)
	FillStrokeRectWithState(pa builder.PageAccess, pageIndex int, x, y, width, height float64, fill style.Color, stroke style.LineStyle, gs style.GraphicsState)

	DrawCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, ls style.LineStyle, gs style.GraphicsState)
	FillCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, gs style.GraphicsState)
	FillStrokeCircleWithState(pa builder.PageAccess, pageIndex int, cx, cy, radius float64, fill style.Color, stroke style.LineStyle, gs style.GraphicsState)

	DrawRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, ls style.LineStyle)
	FillRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, c style.Color)
	FillStrokeRoundedRect(pa builder.PageAccess, pageIndex int, x, y, width, height, radius float64, fill style.Color, stroke style.LineStyle)

	DrawEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, ls style.LineStyle)
	FillEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, c style.Color)
	FillStrokeEllipse(pa builder.PageAccess, pageIndex int, cx, cy, rx, ry float64, fill style.Color, stroke style.LineStyle)

	DrawPolygon(pa builder.PageAccess, pageIndex int, points []float64, ls style.LineStyle)
	FillPolygon(pa builder.PageAccess, pageIndex int, points []float64, c style.Color)
	FillStrokePolygon(pa builder.PageAccess, pageIndex int, points []float64, fill style.Color, stroke style.LineStyle)
}

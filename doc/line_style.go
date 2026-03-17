package doc

// LineCap controls the shape at the endpoints of open subpaths (PDF J operator).
type LineCap int

const (
	LineCapButt   LineCap = 0
	LineCapRound  LineCap = 1
	LineCapSquare LineCap = 2
)

// LineJoin controls the shape at the corners of joined path segments (PDF j operator).
type LineJoin int

const (
	LineJoinMiter LineJoin = 0
	LineJoinRound LineJoin = 1
	LineJoinBevel LineJoin = 2
)

// LineStyle controls the appearance of stroked paths.
type LineStyle struct {
	Width     float64
	Color     Color
	DashArray []float64
	DashPhase float64
	Cap       LineCap
	Join      LineJoin
}

func (s LineStyle) resolvedWidth() float64 {
	if s.Width <= 0 {
		return 1
	}
	return s.Width
}

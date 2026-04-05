package reader

// VectorShape stores simple extracted vector geometry suitable for regeneration.
type VectorShape struct {
	Kind        string
	X1          float64
	Y1          float64
	X2          float64
	Y2          float64
	LineWidth   float64
	Stroke      bool
	Fill        bool
	StrokeColor ColorRGB
	FillColor   ColorRGB
	// Stroke style details
	DashArray []float64
	DashPhase float64
	LineCap   int // 0=Butt,1=Round,2=Square
	LineJoin  int // 0=Miter,1=Round,2=Bevel
	// Opacity (0..1). 0 = default (treated as 1 by codegen)
	StrokeOpacity float64
	FillOpacity   float64
	// Optional rectangular clip applied to this shape
	Clip  bool
	ClipX float64
	ClipY float64
	ClipW float64
	ClipH float64
}

// ColorRGB represents an RGB color triplet.
type ColorRGB struct {
	R float64
	G float64
	B float64
}

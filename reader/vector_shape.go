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
}

// ColorRGB represents an RGB color triplet.
type ColorRGB struct {
	R float64
	G float64
	B float64
}

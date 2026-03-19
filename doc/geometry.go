package doc

// Pt represents a point in 2D space.
type Pt struct {
	X, Y float64
}

// Rect represents a rectangle in 2D space.
type Rect struct {
	X, Y, W, H float64
}

// At returns a Pt from x and y.
func At(x, y float64) Pt {
	return Pt{X: x, Y: y}
}

// R returns a Rect from x, y, w, h.
func R(x, y, w, h float64) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

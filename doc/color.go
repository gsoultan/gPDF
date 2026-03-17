package doc

// Color represents an RGB color with components in [0,1].
type Color struct {
	R, G, B float64
}

var (
	ColorBlack = Color{0, 0, 0}
	ColorWhite = Color{1, 1, 1}
	ColorRed   = Color{1, 0, 0}
	ColorGreen = Color{0, 1, 0}
	ColorBlue  = Color{0, 0, 1}
	ColorGray  = Color{0.5, 0.5, 0.5}
)

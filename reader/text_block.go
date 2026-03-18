package reader

// TextBlock is a positioned text fragment extracted from a page content stream.
type TextBlock struct {
	Text   string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Style  TextStyle
}

package builder

// ImageRun describes one image draw on a page.
type ImageRun struct {
	X, Y              float64
	WidthPt, HeightPt float64
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string
	IsJPEG            bool

	ClipCircle bool
	ClipCX     float64
	ClipCY     float64
	ClipR      float64

	Opacity   float64
	RotateDeg float64

	MCID    int
	HasMCID bool
}

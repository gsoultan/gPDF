package builder

// ImageRun describes one image draw on a page.
type ImageRun struct {
	X, Y              float64
	WidthPt, HeightPt float64
	// Optional exact placement matrix (a,b,c,d,e,f). When non-zero, overrides X/Y/WidthPt/HeightPt.
	Matrix            [6]float64
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

	// Optional soft mask (alpha channel)
	HasMask    bool
	Mask       []byte
	MaskWidth  int
	MaskHeight int

	MCID    int
	HasMCID bool

	IsArtifact bool
}

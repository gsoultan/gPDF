package image

// Options configures image placement. Used by DrawImageWith and DrawTaggedImageWith
// as a cleaner alternative to methods with many positional parameters.
type Options struct {
	PageIndex     int
	X, Y          float64
	Width, Height float64
	// Optional exact placement matrix (a,b,c,d,e,f) in page space. When non-zero, overrides X/Y/Width/Height.
	Matrix           [6]float64
	Data             []byte
	PixelWidth       int
	PixelHeight      int
	BitsPerComponent int
	ColorSpace       string
	AltText          string
	IsJPEG           bool
	Opacity          float64
	RotateDeg        float64
	ClipCircle       bool
	ClipCX, ClipCY   float64
	ClipRadius       float64
	IsArtifact       bool
	// Optional soft mask for alpha
	HasMask    bool
	Mask       []byte
	MaskWidth  int
	MaskHeight int
}

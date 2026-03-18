package image

// Options configures image placement. Used by DrawImageWith and DrawTaggedImageWith
// as a cleaner alternative to methods with many positional parameters.
type Options struct {
	PageIndex        int
	X, Y             float64
	Width, Height    float64
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
}

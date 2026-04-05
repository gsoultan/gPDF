package reader

// ImageInfo holds metadata and raw bytes for a single image XObject found on a page.
type ImageInfo struct {
	Name     string
	Page     int
	X        float64
	Y        float64
	WidthPt  float64
	HeightPt float64
	Rotation float64
	// Exact placement matrix (a,b,c,d,e,f) in page space; if all zeros, use X/Y/WidthPt/HeightPt
	Matrix           [6]float64
	Width            int
	Height           int
	BitsPerComponent int
	ColorSpace       string
	Filter           string
	Format           string
	Data             []byte
	// Optional soft mask (alpha channel) bytes when present via /SMask.
	// SMaskDecoded is true when SMaskData has been fully decompressed and
	// PNG-predictor un-filtered (i.e. it contains raw pixel bytes, not a
	// compressed stream).
	HasSMask     bool
	SMaskData    []byte
	SMaskWidth   int
	SMaskHeight  int
	SMaskDecoded bool

	// NeedsColorConvert is true when ColorSpace is DeviceCMYK and the image
	// data has not been converted to RGB. The consumer must apply CMYK→RGB
	// conversion (or an ICC profile) before display to avoid color shift.
	NeedsColorConvert bool

	// Inferred layout and style
	Opacity    float64
	ClipCircle bool
	ClipCX     float64
	ClipCY     float64
	ClipR      float64
	Wrap       int // 0: None, 1: TopBottom, 2: Square, 3: Tight
	Alignment  int // 0: Left, 1: Center, 2: Right
}

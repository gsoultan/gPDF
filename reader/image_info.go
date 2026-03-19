package reader

// ImageInfo holds metadata and raw bytes for a single image XObject found on a page.
type ImageInfo struct {
	Name             string
	Page             int
	X                float64
	Y                float64
	WidthPt          float64
	HeightPt         float64
	Rotation         float64
	Width            int
	Height           int
	BitsPerComponent int
	ColorSpace       string
	Filter           string
	Format           string
	Data             []byte

	// Inferred layout and style
	Opacity    float64
	ClipCircle bool
	ClipCX     float64
	ClipCY     float64
	ClipR      float64
	Wrap       int // 0: None, 1: TopBottom, 2: Square, 3: Tight
	Alignment  int // 0: Left, 1: Center, 2: Right
}

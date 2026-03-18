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
}

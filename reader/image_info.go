package reader

// ImageInfo holds metadata and raw bytes for a single image XObject found on a page.
type ImageInfo struct {
	Name             string
	Page             int
	Width            int
	Height           int
	BitsPerComponent int
	ColorSpace       string
	Filter           string
	Data             []byte
}

package doc

// ImageBuilder covers image placement on pages.
type ImageBuilder interface {
	DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder
	DrawJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string) *DocumentBuilder
	DrawPNG(x, y, widthPt, heightPt float64, pngData []byte) *DocumentBuilder
}

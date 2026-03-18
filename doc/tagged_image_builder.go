package doc

// TaggedImageBuilder exposes tagged image helpers (figures, JPEG, PNG with alt text).
type TaggedImageBuilder interface {
	DrawTaggedFigure(pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, alt string) *DocumentBuilder
	DrawTaggedJPEG(pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, alt string) *DocumentBuilder
	DrawTaggedPNG(pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) *DocumentBuilder
}

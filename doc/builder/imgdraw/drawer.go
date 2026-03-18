package imgdraw

import (
	"gpdf/doc/builder"
	"gpdf/doc/tagged"
)

// Drawer draws images onto PDF pages via PageAccess.
// Non-tagged methods target the last page; tagged methods accept an explicit page index
// and return a tagged.Figure for the caller to register in tagging state.
type Drawer interface {
	DrawImage(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string)
	DrawJPEG(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string)
	DrawPNG(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte) error

	DrawCircularImage(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, cx, cy, radius float64)
	DrawCircularJPEG(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, cx, cy, radius float64)
	DrawCircularPNG(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, cx, cy, radius float64) error

	DrawImageWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, opacity float64)
	DrawJPEGWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, opacity float64)
	DrawPNGWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, opacity float64) error

	DrawJPEGRotated(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, rotateDeg float64)
	DrawPNGRotated(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, rotateDeg float64) error

	DrawTaggedFigure(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace, alt string) tagged.Figure
	DrawTaggedJPEG(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace, alt string) tagged.Figure
	DrawTaggedPNG(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) (tagged.Figure, error)
}

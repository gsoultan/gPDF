package table

import "github.com/gsoultan/gpdf/doc/style"

// CellImageSpec describes an image to be placed inside a table cell.
type CellImageSpec struct {
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string
	WidthPt, HeightPt float64
	IsJPEG            bool

	Wrap style.ImageWrap
	Side style.ImageSide

	// PaddingPt defines extra space around the image.
	PaddingPt float64
}

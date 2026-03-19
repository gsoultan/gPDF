package table

// ImageWrap defines the text wrapping style around an image in a table cell.
type ImageWrap int

const (
	// ImageWrapTopBottom is the default: image takes up full width (or aligned width) and text begins below it.
	ImageWrapTopBottom ImageWrap = iota
	// ImageWrapInline is similar to TopBottom, but if the image height is small enough, it could be part of a line.
	// For gPDF tables, we can treat it similarly to TopBottom but perhaps with different spacing.
	ImageWrapInline
	// ImageWrapSquare makes the text wrap around the image's rectangular boundary.
	ImageWrapSquare
	// ImageWrapTight and ImageWrapThrough are similar to Square but follow image contours more closely.
	// We will treat them similarly to Square for now.
	ImageWrapTight
	// ImageWrapThrough
)

// ImageSide defines which side of the cell the image is aligned to for wrapping.
type ImageSide int

const (
	ImageSideLeft ImageSide = iota
	ImageSideRight
)

// CellImageSpec describes an image to be placed inside a table cell.
type CellImageSpec struct {
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string
	WidthPt, HeightPt float64
	IsJPEG            bool

	Wrap ImageWrap
	Side ImageSide

	// PaddingPt defines extra space around the image.
	PaddingPt float64
}

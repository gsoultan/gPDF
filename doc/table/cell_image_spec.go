package table

// CellImageSpec describes an image to be placed inside a table cell.
type CellImageSpec struct {
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string
	WidthPt, HeightPt float64
}

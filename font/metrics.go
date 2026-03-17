package font

// Metrics holds typographic measurements for a font, all in font units (typically 1/1000 em or unitsPerEm).
type Metrics struct {
	UnitsPerEm  int
	Ascent      int
	Descent     int
	LineGap     int
	CapHeight   int
	XHeight     int
	ItalicAngle float64
	Flags       int
	BBox        [4]float64 // font bounding box: llx, lly, urx, ury
	StemV       int
}

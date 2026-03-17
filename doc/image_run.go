package doc

// imageRun describes one image draw on a page (Image XObject placed via Do).
type imageRun struct {
	X, Y              float64 // position in points (lower-left)
	WidthPt, HeightPt float64 // display size in points
	Raw               []byte  // raw image stream bytes (decoded, or JPEG when isJPEG)
	WidthPx, HeightPx int     // dimensions in samples
	BitsPerComponent  int     // 1, 2, 4, 8, 12, 16
	ColorSpace        string  // e.g. DeviceGray, DeviceRGB, DeviceCMYK
	isJPEG            bool    // when true, Raw contains JPEG data (DCTDecode)

	// tagging metadata; used when this image is part of a tagged structure element.
	MCID    int
	HasMCID bool
}

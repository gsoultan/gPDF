package doc

import "gpdf/model"

// textRun describes one text draw on a page (simple PDF text; no Unicode/CMap).
type textRun struct {
	Text     string
	X, Y     float64
	FontName string
	FontSize float64

	// TextColorRGB is optional RGB color in [0,1] for this run.
	// When all components are zero and UseDefaultColor is true, the current color is used.
	TextColorRGB    [3]float64
	UseDefaultColor bool

	// tagging metadata; used when this text is part of a tagged structure element.
	MCID    int
	HasMCID bool
	Role    model.Name // /TH or /TD for table cells; empty for untagged text
}

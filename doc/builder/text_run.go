package builder

import "github.com/gsoultan/gpdf/model"

// TextRun describes one text draw on a page.
type TextRun struct {
	Text     string
	X, Y     float64
	FontName string
	FontSize float64

	TextColorRGB    [3]float64
	UseDefaultColor bool

	Underline       bool
	Strikethrough   bool
	LetterSpacing   float64
	WordSpacing     float64
	HorizontalScale float64 // percent, 100 = normal

	Rotation float64 // Degrees counter-clockwise; 0 = upright

	SyntheticBold   bool
	SyntheticItalic bool

	MCID    int
	HasMCID bool
	Role    model.Name

	IsArtifact bool
}

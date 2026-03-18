package style

import "gpdf/model"

// BlendMode names the PDF blend mode (ExtGState /BM value).
type BlendMode string

const (
	BlendNormal     BlendMode = "Normal"
	BlendMultiply   BlendMode = "Multiply"
	BlendScreen     BlendMode = "Screen"
	BlendOverlay    BlendMode = "Overlay"
	BlendDarken     BlendMode = "Darken"
	BlendLighten    BlendMode = "Lighten"
	BlendColorDodge BlendMode = "ColorDodge"
	BlendColorBurn  BlendMode = "ColorBurn"
	BlendHardLight  BlendMode = "HardLight"
	BlendSoftLight  BlendMode = "SoftLight"
	BlendDifference BlendMode = "Difference"
	BlendExclusion  BlendMode = "Exclusion"
)

// GraphicsState controls transparency and blending for drawing operations.
// Zero value produces default behavior (fully opaque, Normal blend).
type GraphicsState struct {
	FillOpacity   float64
	StrokeOpacity float64
	BlendMode     BlendMode
}

// IsDefault returns true when the state represents default (fully opaque, Normal blend).
func (s GraphicsState) IsDefault() bool {
	noFillChange := s.FillOpacity <= 0 || s.FillOpacity >= 1
	noStrokeChange := s.StrokeOpacity <= 0 || s.StrokeOpacity >= 1
	noBlendChange := s.BlendMode == "" || s.BlendMode == BlendNormal
	return noFillChange && noStrokeChange && noBlendChange
}

// ExtGStateDict builds the PDF ExtGState dictionary for this state.
func (s GraphicsState) ExtGStateDict() model.Dict {
	d := model.Dict{
		model.Name("Type"): model.Name("ExtGState"),
	}
	if s.FillOpacity > 0 && s.FillOpacity < 1 {
		d[model.Name("ca")] = model.Real(s.FillOpacity)
	}
	if s.StrokeOpacity > 0 && s.StrokeOpacity < 1 {
		d[model.Name("CA")] = model.Real(s.StrokeOpacity)
	}
	if s.BlendMode != "" && s.BlendMode != BlendNormal {
		d[model.Name("BM")] = model.Name(string(s.BlendMode))
	}
	return d
}

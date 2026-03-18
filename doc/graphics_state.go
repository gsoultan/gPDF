package doc

import "gpdf/doc/style"

type BlendMode = style.BlendMode
type GraphicsState = style.GraphicsState

const (
	BlendNormal     = style.BlendNormal
	BlendMultiply   = style.BlendMultiply
	BlendScreen     = style.BlendScreen
	BlendOverlay    = style.BlendOverlay
	BlendDarken     = style.BlendDarken
	BlendLighten    = style.BlendLighten
	BlendColorDodge = style.BlendColorDodge
	BlendColorBurn  = style.BlendColorBurn
	BlendHardLight  = style.BlendHardLight
	BlendSoftLight  = style.BlendSoftLight
	BlendDifference = style.BlendDifference
	BlendExclusion  = style.BlendExclusion
)

package doc

import "gpdf/font"

// fontConfig holds font-related state for the builder.
type fontConfig struct {
	fonts         map[string]font.Font
	embeddedFonts map[string]*embeddedFontUsage
}

func (fc *fontConfig) registerFont(f font.Font) {
	if fc.fonts == nil {
		fc.fonts = make(map[string]font.Font)
	}
	fc.fonts[f.PostScriptName()] = f
	if ef, ok := f.(font.EmbeddableFont); ok {
		if fc.embeddedFonts == nil {
			fc.embeddedFonts = make(map[string]*embeddedFontUsage)
		}
		fc.embeddedFonts[ef.PostScriptName()] = newEmbeddedFontUsage(ef)
	}
}

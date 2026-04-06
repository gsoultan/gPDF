package doc

import (
	"strings"
	"unicode"

	"github.com/gsoultan/gpdf/font"
)

var standardSubstitutions = map[string]string{
	"Arial":           "Helvetica",
	"Times New Roman": "Times-Roman",
	"Courier New":     "Courier",
}

// fontConfig holds font-related state for the builder.
type fontConfig struct {
	fonts          map[string]font.Font
	embeddedFonts  map[string]*embeddedFontUsage
	fallbackChain  []string
	blockFallbacks map[string]string
	onWarning      func(string)
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

func (fc *fontConfig) addFallback(fontName string) {
	fc.fallbackChain = append(fc.fallbackChain, fontName)
}

func (fc *fontConfig) setBlockFallback(block string, fontName string) {
	if fc.blockFallbacks == nil {
		fc.blockFallbacks = make(map[string]string)
	}
	fc.blockFallbacks[block] = fontName
}

func (fc *fontConfig) resolveFontName(name string) string {
	if f, ok := fc.fonts[name]; ok {
		return f.PostScriptName()
	}
	if sub, ok := standardSubstitutions[name]; ok {
		return sub
	}
	return name
}

// resolveFont segments text by font support. It tries primaryFontName first,
// then iterates through the fallbackChain.
func (fc *fontConfig) resolveFont(text string, primaryFontName string) []fontSegment {
	if text == "" {
		return nil
	}

	primaryFontName = fc.resolveFontName(primaryFontName)

	var segments []fontSegment
	var currentSegment strings.Builder
	var currentFont string
	var currentRTL bool

	for _, r := range text {
		resolved := fc.findFontForRune(r, primaryFontName)
		if resolved == "" {
			resolved = primaryFontName // Fallback to primary if nothing found (will render placeholder)
		}
		isRTL := unicode.Is(unicode.Arabic, r) || unicode.Is(unicode.Hebrew, r)

		if (resolved != currentFont || isRTL != currentRTL) && currentSegment.Len() > 0 {
			segments = append(segments, fontSegment{
				text:     currentSegment.String(),
				fontName: currentFont,
				isRTL:    currentRTL,
			})
			currentSegment.Reset()
		}

		currentFont = resolved
		currentRTL = isRTL
		currentSegment.WriteRune(r)
	}

	if currentSegment.Len() > 0 {
		segments = append(segments, fontSegment{
			text:     currentSegment.String(),
			fontName: currentFont,
			isRTL:    currentRTL,
		})
	}

	return segments
}

func (fc *fontConfig) findFontForRune(r rune, primary string) string {
	// Try primary
	if f, ok := fc.fonts[primary]; ok {
		if f.Contains(r) {
			return primary
		}
	} else {
		// For standard fonts, we don't have a Font object, but we check if it's a known standard font.
		if font.GetStandardWidth(primary, r) > 0 || r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return primary
		}
	}

	// Try block-specific fallback
	block := getRuneBlock(r)
	if name, ok := fc.blockFallbacks[block]; ok {
		if f, ok := fc.fonts[name]; ok && f.Contains(r) {
			return name
		}
	}

	// Try fallback chain
	for _, name := range fc.fallbackChain {
		if name == primary {
			continue
		}
		if f, ok := fc.fonts[name]; ok && f.Contains(r) {
			return name
		}
	}

	if fc.onWarning != nil {
		fc.onWarning(strings.Join([]string{"no glyph for", string(r), "in font", primary, "or fallbacks"}, " "))
	}
	return ""
}

func getRuneBlock(r rune) string {
	switch {
	case unicode.Is(unicode.Han, r), unicode.Is(unicode.Hiragana, r), unicode.Is(unicode.Katakana, r), unicode.Is(unicode.Hangul, r):
		return "CJK"
	case unicode.Is(unicode.Arabic, r):
		return "Arabic"
	case unicode.Is(unicode.Hebrew, r):
		return "Hebrew"
	case unicode.Is(unicode.Thai, r):
		return "Thai"
	case unicode.Is(unicode.Bengali, r), unicode.Is(unicode.Devanagari, r), unicode.Is(unicode.Gujarati, r), unicode.Is(unicode.Gurmukhi, r), unicode.Is(unicode.Kannada, r), unicode.Is(unicode.Malayalam, r), unicode.Is(unicode.Oriya, r), unicode.Is(unicode.Tamil, r), unicode.Is(unicode.Telugu, r):
		return "Indic"
	default:
		return "Default"
	}
}

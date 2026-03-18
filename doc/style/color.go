package style

import (
	"fmt"
	"strings"
)

// Color represents an RGB color with components in [0,1].
type Color struct {
	R, G, B float64
}

// ColorFromHex parses a CSS-style hex color string (e.g. "#FF5733" or "FF5733")
// and returns the corresponding Color. Returns an error for invalid input.
func ColorFromHex(hex string) (Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b uint8
	switch len(hex) {
	case 6:
		n, _ := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if n != 3 {
			return Color{}, fmt.Errorf("invalid hex color %q: expected 6 hex digits", hex)
		}
	case 3:
		expanded := string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		n, _ := fmt.Sscanf(expanded, "%02x%02x%02x", &r, &g, &b)
		if n != 3 {
			return Color{}, fmt.Errorf("invalid hex color %q: expected 3 hex digits", hex)
		}
	default:
		return Color{}, fmt.Errorf("invalid hex color %q: expected 3 or 6 hex digits, got %d", hex, len(hex))
	}
	return Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}, nil
}

// Gray returns a gray with the given intensity in [0,1] (0=black, 1=white).
func Gray(v float64) Color { return Color{R: v, G: v, B: v} }

var (
	Black     = Color{0, 0, 0}
	White     = Color{1, 1, 1}
	Red       = Color{1, 0, 0}
	Green     = Color{0, 0.502, 0}
	Blue      = Color{0, 0, 1}
	GrayColor = Color{0.5, 0.5, 0.5}
	LightGray = Color{0.827, 0.827, 0.827}
	DarkGray  = Color{0.251, 0.251, 0.251}
	Orange    = Color{1, 0.647, 0}
	Yellow    = Color{1, 1, 0}
	Purple    = Color{0.502, 0, 0.502}
	Cyan      = Color{0, 1, 1}
	Magenta   = Color{1, 0, 1}
	Pink      = Color{1, 0.753, 0.796}
	Brown     = Color{0.647, 0.165, 0.165}
	Teal      = Color{0, 0.502, 0.502}
	Navy      = Color{0, 0, 0.502}
	Maroon    = Color{0.502, 0, 0}
	Olive     = Color{0.502, 0.502, 0}
	Silver    = Color{0.753, 0.753, 0.753}
	Lime      = Color{0, 1, 0}
)

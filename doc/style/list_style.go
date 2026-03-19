package style

// ListMarker defines the type of bullet or numbering used for lists.
type ListMarker string

const (
	ListMarkerDisc       ListMarker = "disc"        // ●
	ListMarkerCircle     ListMarker = "circle"      // ○
	ListMarkerSquare     ListMarker = "square"      // ■
	ListMarkerDecimal    ListMarker = "decimal"     // 1, 2, 3...
	ListMarkerRomanUpper ListMarker = "roman-upper" // I, II, III...
	ListMarkerRomanLower ListMarker = "roman-lower" // i, ii, iii...
	ListMarkerAlphaUpper ListMarker = "alpha-upper" // A, B, C...
	ListMarkerAlphaLower ListMarker = "alpha-lower" // a, b, c...
	ListMarkerCustom     ListMarker = "custom"      // use CustomMarker
	ListMarkerNone       ListMarker = "none"
)

// ListStyle defines how a list should be rendered.
type ListStyle struct {
	Marker         ListMarker
	CustomMarker   string
	Indent         float64
	MarkerOffset   float64
	FontName       string
	FontSize       float64
	MarkerFontSize float64
	Color          Color
	Level          int
}

// DefaultListStyle returns a sensible default for lists.
func DefaultListStyle() ListStyle {
	return ListStyle{
		Marker:       ListMarkerDisc,
		Indent:       18.0,
		MarkerOffset: 12.0,
		FontSize:     11.0,
		Color:        Black,
	}
}

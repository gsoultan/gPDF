package doc

// TextAlignment defines horizontal alignment for text.
type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
	AlignJustify
)

// TextStyle describes the visual appearance of a text run.
type TextStyle struct {
	FontName        string
	FontSize        float64
	Color           Color
	LetterSpacing   float64 // CharSpacing in PDF terms
	WordSpacing     float64 // Additional space between words
	HorizontalScale float64 // Percent; 100 = normal, 0 = use default
	Leading         float64 // Line spacing override; 0 = use default
	Rotation        float64 // Degrees; 0 = upright
	IsVertical      bool

	SyntheticBold   bool
	SyntheticItalic bool
}

// DefaultTextStyle returns a default Helvetica 12pt style.
func DefaultTextStyle() TextStyle {
	return TextStyle{
		FontName:      "Helvetica",
		FontSize:      12,
		Color:         ColorBlack,
		LetterSpacing: 0,
	}
}

// Font returns a new style with the given font name.
func (s TextStyle) Font(name string) TextStyle {
	s.FontName = name
	return s
}

// Size returns a new style with the given font size.
func (s TextStyle) Size(size float64) TextStyle {
	s.FontSize = size
	return s
}

// WithColor returns a new style with the given color.
func (s TextStyle) WithColor(c Color) TextStyle {
	s.Color = c
	return s
}

// WithLetterSpacing returns a new style with the given letter spacing.
func (s TextStyle) WithLetterSpacing(spacing float64) TextStyle {
	s.LetterSpacing = spacing
	return s
}

// WithWordSpacing returns a new style with the given word spacing.
func (s TextStyle) WithWordSpacing(spacing float64) TextStyle {
	s.WordSpacing = spacing
	return s
}

// WithHorizontalScale returns a new style with the given horizontal scale (100 = normal).
func (s TextStyle) WithHorizontalScale(scale float64) TextStyle {
	s.HorizontalScale = scale
	return s
}

// WithLeading returns a new style with the given leading (line spacing).
func (s TextStyle) WithLeading(leading float64) TextStyle {
	s.Leading = leading
	return s
}

// WithRotation returns a new style with the given rotation in degrees.
func (s TextStyle) WithRotation(deg float64) TextStyle {
	s.Rotation = deg
	return s
}

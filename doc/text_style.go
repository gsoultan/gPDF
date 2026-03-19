package doc

// TextAlignment defines horizontal alignment for text.
type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

// TextStyle describes the visual appearance of a text run.
type TextStyle struct {
	FontName string
	FontSize float64
	Color    Color
}

// DefaultTextStyle returns a default Helvetica 12pt style.
func DefaultTextStyle() TextStyle {
	return TextStyle{
		FontName: "Helvetica",
		FontSize: 12,
		Color:    ColorBlack,
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

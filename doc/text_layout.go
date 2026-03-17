package doc

// TextAlign describes horizontal alignment for laid-out text.
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
	TextAlignJustify
)

// TextLayoutOptions configures layout for paragraph-like text.
// Width is required; when zero or negative, the call is ignored.
// LineHeight <= 0 falls back to fontSize * 1.2.
// ParagraphSpacing <= 0 means no extra gap after the block.
// When AllowPageBreak is true, text can continue on subsequent pages if they exist.
type TextLayoutOptions struct {
	Width            float64
	Align            TextAlign
	LineHeight       float64
	ParagraphSpacing float64
	AllowPageBreak   bool
}

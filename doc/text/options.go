package text

import "github.com/gsoultan/gpdf/doc/style"

// Align describes horizontal alignment for laid-out text.
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
	AlignJustify
)

// LayoutOptions configures layout for paragraph-like text.
type LayoutOptions struct {
	Width            float64
	Align            Align
	LineHeight       float64
	ParagraphSpacing float64
	AllowPageBreak   bool
	LetterSpacing    float64
	LineRectFn       LineRectFunc
	Color            style.Color
	HasColor         bool
	IsVertical       bool
	SyntheticBold    bool
	SyntheticItalic  bool
}

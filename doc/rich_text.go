package doc

// RichText represents a sequence of styled text segments.
type RichText struct {
	Segments []RichTextSegment
}

// NewRichText creates a new empty RichText.
func NewRichText() *RichText {
	return &RichText{}
}

// Add adds a new styled segment to the rich text.
func (rt *RichText) Add(text string, style TextStyle) *RichText {
	rt.Segments = append(rt.Segments, RichTextSegment{Text: text, Style: style})
	return rt
}

// AddSimple adds a segment with the default font and size.
func (rt *RichText) AddSimple(text string) *RichText {
	return rt.Add(text, DefaultTextStyle())
}

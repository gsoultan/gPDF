package xobject

import "gpdf/model"

// Image wraps a stream that is an Image XObject (/Type /XObject, /Subtype /Image).
// Use IsImageXObject to check before wrapping.
type Image struct {
	Stream *model.Stream
}

// IsImageXObject returns true if the stream dict has /Type /XObject and /Subtype /Image.
func IsImageXObject(s *model.Stream) bool {
	if s == nil || s.Dict == nil {
		return false
	}
	typeVal, _ := s.Dict[model.Name("Type")].(model.Name)
	subVal, _ := s.Dict[model.Name("Subtype")].(model.Name)
	return typeVal == "XObject" && subVal == "Image"
}

// Width returns the image width in samples; 0 if missing.
func (i *Image) Width() int {
	if i == nil || i.Stream == nil {
		return 0
	}
	switch v := i.Stream.Dict[model.Name("Width")].(type) {
	case model.Integer:
		return int(v)
	}
	return 0
}

// Height returns the image height in samples; 0 if missing.
func (i *Image) Height() int {
	if i == nil || i.Stream == nil {
		return 0
	}
	switch v := i.Stream.Dict[model.Name("Height")].(type) {
	case model.Integer:
		return int(v)
	}
	return 0
}

// BitsPerComponent returns bits per component; 0 if missing.
func (i *Image) BitsPerComponent() int {
	if i == nil || i.Stream == nil {
		return 0
	}
	switch v := i.Stream.Dict[model.Name("BitsPerComponent")].(type) {
	case model.Integer:
		return int(v)
	}
	return 0
}

// ColorSpace returns the /ColorSpace entry (Name or Array); nil if missing.
func (i *Image) ColorSpace() model.Object {
	if i == nil || i.Stream == nil {
		return nil
	}
	return i.Stream.Dict[model.Name("ColorSpace")]
}

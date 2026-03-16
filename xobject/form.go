package xobject

import "gpdf/model"

// Form wraps a stream that is a Form XObject (/Type /XObject, /Subtype /Form).
// Use IsFormXObject to check before wrapping.
type Form struct {
	Stream *model.Stream
}

// IsFormXObject returns true if the stream dict has /Type /XObject and /Subtype /Form.
func IsFormXObject(s *model.Stream) bool {
	if s == nil || s.Dict == nil {
		return false
	}
	typeVal, _ := s.Dict[model.Name("Type")].(model.Name)
	subVal, _ := s.Dict[model.Name("Subtype")].(model.Name)
	return typeVal == "XObject" && subVal == "Form"
}

// BBox returns the form’s /BBox entry (array of four numbers [llx lly urx ury]) if present.
func (f *Form) BBox() (model.Array, bool) {
	if f == nil || f.Stream == nil {
		return nil, false
	}
	v, ok := f.Stream.Dict[model.Name("BBox")].(model.Array)
	return v, ok
}

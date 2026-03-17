package model

// Annotation is a PDF annotation (page or form). Common keys: /Type (/Annot), /Subtype, /Rect, /A, /Dest.
type Annotation struct {
	Dict Dict
}

// Subtype returns the annotation subtype (e.g. Link, Highlight, Widget).
func (a *Annotation) Subtype() Name {
	if a == nil || a.Dict == nil {
		return ""
	}
	if v, ok := a.Dict[Name("Subtype")].(Name); ok {
		return v
	}
	return ""
}

// Rect returns the annotation rectangle [llx, lly, urx, ury] if present.
func (a *Annotation) Rect() Array {
	if a == nil || a.Dict == nil {
		return nil
	}
	if v, ok := a.Dict[Name("Rect")].(Array); ok {
		return v
	}
	return nil
}

// Dest returns the destination array for link annotations (when not using /A).
func (a *Annotation) Dest() Array {
	if a == nil || a.Dict == nil {
		return nil
	}
	if v, ok := a.Dict[Name("Dest")].(Array); ok {
		return v
	}
	return nil
}

// ARef returns the reference to the action dictionary (/A) if present.
func (a *Annotation) ARef() *Ref {
	if a == nil || a.Dict == nil {
		return nil
	}
	if v, ok := a.Dict[Name("A")].(Ref); ok {
		return &v
	}
	return nil
}

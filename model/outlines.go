package model

// Outlines is the root of the document outline (bookmarks) dictionary.
type Outlines struct {
	Dict Dict
}

// First returns the reference to the first top-level outline item, or nil.
func (o *Outlines) First() *Ref {
	if o == nil || o.Dict == nil {
		return nil
	}
	if v, ok := o.Dict[Name("First")].(Ref); ok {
		return &v
	}
	return nil
}

// Last returns the reference to the last top-level outline item, or nil.
func (o *Outlines) Last() *Ref {
	if o == nil || o.Dict == nil {
		return nil
	}
	if v, ok := o.Dict[Name("Last")].(Ref); ok {
		return &v
	}
	return nil
}

// Count returns the total number of outline entries (can be negative per spec).
func (o *Outlines) Count() int64 {
	if o == nil || o.Dict == nil {
		return 0
	}
	if v, ok := o.Dict[Name("Count")].(Integer); ok {
		return int64(v)
	}
	return 0
}

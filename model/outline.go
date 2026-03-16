package model

// Outlines is the root of the document outline (bookmarks) dictionary.
// Keys: /Type (/Outlines), /First, /Last, /Count.
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

// OutlineItem is a single outline (bookmark) entry.
// Keys: /Title, /Parent, /Prev, /Next, /First, /Last, /Count, /Dest or /A.
type OutlineItem struct {
	Dict Dict
}

// Title returns the bookmark title string.
func (i *OutlineItem) Title() string {
	if i == nil || i.Dict == nil {
		return ""
	}
	if v, ok := i.Dict[Name("Title")].(String); ok {
		return string(v)
	}
	return ""
}

// Parent returns the reference to the parent outline item or Outlines root.
func (i *OutlineItem) Parent() *Ref {
	if i == nil || i.Dict == nil {
		return nil
	}
	if v, ok := i.Dict[Name("Parent")].(Ref); ok {
		return &v
	}
	return nil
}

// Next returns the reference to the next sibling outline item, or nil.
func (i *OutlineItem) Next() *Ref {
	if i == nil || i.Dict == nil {
		return nil
	}
	if v, ok := i.Dict[Name("Next")].(Ref); ok {
		return &v
	}
	return nil
}

// Dest returns the destination array (e.g. [pageRef /Fit]), or nil if the item uses /A instead.
func (i *OutlineItem) Dest() Array {
	if i == nil || i.Dict == nil {
		return nil
	}
	if v, ok := i.Dict[Name("Dest")].(Array); ok {
		return v
	}
	return nil
}

// ARef returns the reference to the action dictionary (/A) if present (e.g. URI or GoTo action).
func (i *OutlineItem) ARef() *Ref {
	if i == nil || i.Dict == nil {
		return nil
	}
	if v, ok := i.Dict[Name("A")].(Ref); ok {
		return &v
	}
	return nil
}

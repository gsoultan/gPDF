package model

// AssociatedFiles returns the /AF (associated files) array from the page dictionary (PDF 2.0).
// Each element is typically a file specification dictionary reference.
func (p Page) AssociatedFiles() Array {
	if p.Dict == nil {
		return nil
	}
	if v, ok := p.Dict[Name("AF")].(Array); ok {
		return v
	}
	return nil
}

// DPartRef returns the /DPart (document part) reference from the page dictionary (PDF 2.0).
func (p Page) DPartRef() *Ref {
	if p.Dict == nil {
		return nil
	}
	if v, ok := p.Dict[Name("DPart")].(Ref); ok {
		return &v
	}
	return nil
}

// Tabs returns the /Tabs name indicating the tab order for annotations on this page (PDF 1.5+).
func (p Page) Tabs() Name {
	if p.Dict == nil {
		return ""
	}
	if v, ok := p.Dict[Name("Tabs")].(Name); ok {
		return v
	}
	return ""
}

// UserUnit returns the /UserUnit value (points per user unit, PDF 1.6+). Defaults to 1.0.
func (p Page) UserUnit() float64 {
	if p.Dict == nil {
		return 1.0
	}
	switch v := p.Dict[Name("UserUnit")].(type) {
	case Real:
		return float64(v)
	case Integer:
		return float64(v)
	}
	return 1.0
}

// BleedBox returns the /BleedBox array if present.
func (p Page) BleedBox() (Array, bool) {
	v, ok := p.Dict[Name("BleedBox")].(Array)
	return v, ok
}

// TrimBox returns the /TrimBox array if present.
func (p Page) TrimBox() (Array, bool) {
	v, ok := p.Dict[Name("TrimBox")].(Array)
	return v, ok
}

// ArtBox returns the /ArtBox array if present.
func (p Page) ArtBox() (Array, bool) {
	v, ok := p.Dict[Name("ArtBox")].(Array)
	return v, ok
}

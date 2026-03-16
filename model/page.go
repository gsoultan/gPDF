package model

// Page represents a single PDF page (page dictionary).
// Common keys: /Type (/Page), /Parent, /MediaBox, /Contents, /Resources.
type Page struct {
	Dict Dict
}

// MediaBox returns the page media box as [llx, lly, urx, ury] if present.
func (p Page) MediaBox() (arr Array, ok bool) {
	v, ok := p.Dict[Name("MediaBox")].(Array)
	return v, ok
}

// Contents returns the page contents (single Ref or Array of Refs) if present.
func (p Page) Contents() Object {
	return p.Dict[Name("Contents")]
}

// Resources returns the /Resources dictionary if present.
func (p Page) Resources() (Dict, bool) {
	v, ok := p.Dict[Name("Resources")].(Dict)
	return v, ok
}

// XObjects returns the /XObject subdictionary from Resources (name -> Ref) if present.
func (p Page) XObjects() (Dict, bool) {
	res, ok := p.Resources()
	if !ok {
		return nil, false
	}
	v, ok := res[Name("XObject")].(Dict)
	return v, ok
}

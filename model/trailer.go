package model

// Trailer holds trailer dictionary entries: /Root, /Size, /Encrypt, /Info, etc.
type Trailer struct {
	Dict Dict
}

// Root returns the catalog reference from the trailer, or nil if absent.
func (t Trailer) Root() *Ref {
	if v, ok := t.Dict[Name("Root")].(Ref); ok {
		return &v
	}
	return nil
}

// Size returns the total number of entries in the cross-reference table.
func (t Trailer) Size() int {
	if v, ok := t.Dict[Name("Size")].(Integer); ok {
		return int(v)
	}
	return 0
}

// Info returns the document Info dictionary reference from the trailer, or nil if absent.
func (t Trailer) Info() *Ref {
	if v, ok := t.Dict[Name("Info")].(Ref); ok {
		return &v
	}
	return nil
}

// Encrypt returns the Encrypt dictionary reference from the trailer, or nil if absent.
func (t Trailer) Encrypt() *Ref {
	if v, ok := t.Dict[Name("Encrypt")].(Ref); ok {
		return &v
	}
	return nil
}

// ID returns the file identifier array (two byte strings) from the trailer, or nil if absent.
func (t Trailer) ID() Array {
	if v, ok := t.Dict[Name("ID")].(Array); ok {
		return v
	}
	return nil
}

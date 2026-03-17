package model

// Perms is the document permissions dictionary (Catalog /Perms). Used for DocMDP (signature) and other permissions.
// Keys: /DocMDP (reference to signature field or dict).
type Perms struct {
	Dict Dict
}

// DocMDPRef returns the reference to the DocMDP signature (document modification permissions) if present.
func (p *Perms) DocMDPRef() *Ref {
	if p == nil || p.Dict == nil {
		return nil
	}
	if v, ok := p.Dict[Name("DocMDP")].(Ref); ok {
		return &v
	}
	return nil
}

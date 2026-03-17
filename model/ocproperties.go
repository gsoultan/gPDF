package model

// OCProperties is the optional content properties dictionary (Catalog /OCProperties).
// Keys: /OCGs (array of OCG dicts), /D (default config), /Configs.
type OCProperties struct {
	Dict Dict
}

// OCGs returns the array of optional content group dictionaries (or refs).
func (o *OCProperties) OCGs() Array {
	if o == nil || o.Dict == nil {
		return nil
	}
	if v, ok := o.Dict[Name("OCGs")].(Array); ok {
		return v
	}
	return nil
}

// DRef returns the reference to the default optional content configuration dictionary.
func (o *OCProperties) DRef() *Ref {
	if o == nil || o.Dict == nil {
		return nil
	}
	if v, ok := o.Dict[Name("D")].(Ref); ok {
		return &v
	}
	return nil
}

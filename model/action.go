package model

// Action is an action dictionary (e.g. /A on an outline item or annotation).
// Keys: /Type (/Action), /S (action type), /D (GoTo destination), /URI (URI action), etc.
type Action struct {
	Dict Dict
}

// S returns the action type (e.g. GoTo, URI).
func (a *Action) S() Name {
	if a == nil || a.Dict == nil {
		return ""
	}
	if v, ok := a.Dict[Name("S")].(Name); ok {
		return v
	}
	return ""
}

// D returns the destination for GoTo actions: either an Array (explicit) or a Name (named destination).
func (a *Action) D() Object {
	if a == nil || a.Dict == nil {
		return nil
	}
	return a.Dict[Name("D")]
}

// URI returns the URL string for URI actions.
func (a *Action) URI() string {
	if a == nil || a.Dict == nil {
		return ""
	}
	if v, ok := a.Dict[Name("URI")].(String); ok {
		return string(v)
	}
	return ""
}

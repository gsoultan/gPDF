package model

// StructElem represents a structure element in the tagged PDF structure tree.
// Typical roles include /Document, /Sect, /P, /Table, /TR, /TH, /TD, etc.
// Keys: /S (role), /P (parent), /K (kids), /Pg (page), /ID, /A, /Lang, /Alt, /ActualText.
type StructElem struct {
	Dict Dict
}

// S returns the structure type name (/S) such as /Table, /TR, /TH, /TD.
func (e *StructElem) S() Name {
	if e == nil || e.Dict == nil {
		return ""
	}
	if v, ok := e.Dict[Name("S")].(Name); ok {
		return v
	}
	return ""
}

// K returns the /K entry (kids) which may be an Array, a single StructElem dict, or marked-content reference.
func (e *StructElem) K() Object {
	if e == nil || e.Dict == nil {
		return nil
	}
	return e.Dict[Name("K")]
}

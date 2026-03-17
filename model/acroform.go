package model

// AcroForm is the interactive form (AcroForm) dictionary. Keys: /Fields, /SigFlags, /CO, /DA, etc.
type AcroForm struct {
	Dict Dict
}

// Fields returns the array of references to field dictionaries (root fields).
func (a *AcroForm) Fields() Array {
	if a == nil || a.Dict == nil {
		return nil
	}
	if v, ok := a.Dict[Name("Fields")].(Array); ok {
		return v
	}
	return nil
}

// SigFlags returns the signature flags (e.g. 3 = append-only / signatures exist). 0 if absent.
func (a *AcroForm) SigFlags() int64 {
	if a == nil || a.Dict == nil {
		return 0
	}
	if v, ok := a.Dict[Name("SigFlags")].(Integer); ok {
		return int64(v)
	}
	return 0
}

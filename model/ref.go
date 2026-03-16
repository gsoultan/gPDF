package model

// Ref is an indirect reference: object number and generation.
type Ref struct {
	ObjectNumber int
	Generation   int
}

func (Ref) IsIndirect() bool { return true }

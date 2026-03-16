package model

// Null represents the PDF null object.
type Null struct{}

func (Null) IsIndirect() bool { return false }

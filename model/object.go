package model

// Object represents a PDF value type (ISO 32000-2). Implementations are immutable
// value types: Boolean, Integer, Real, String, Name, Array, Dictionary, Stream, Null, Ref.
type Object interface {
	// IsIndirect returns true if this object is an indirect reference.
	IsIndirect() bool
}

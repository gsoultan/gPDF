package content

// Parser parses PDF content stream bytes into a sequence of operators.
type Parser interface {
	// Parse parses the content stream (decoded bytes) and returns operators in order.
	// Callers are responsible for decoding any stream filters (e.g. FlateDecode) first.
	Parse(stream []byte) ([]Op, error)
}

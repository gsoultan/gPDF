package syntax

// Tokenizer produces a sequence of PDF tokens from input.
type Tokenizer interface {
	Next() (Token, error)
	CurrentOffset() int64
}

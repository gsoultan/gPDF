package syntax

// TokenKind identifies the type of a PDF token.
type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenInteger
	TokenReal
	TokenKeyword  // obj, endobj, stream, endstream, xref, trailer, startxref, R
	TokenName     // /Name
	TokenLiteral  // ( ... )
	TokenHex      // < ... >
	TokenLBracket // [
	TokenRBracket // ]
	TokenLDict    // <<
	TokenRDict    // >>
	TokenComment  // % ... (may be skipped by consumer)
	TokenUnknown  // unrecognized keyword-like token
)

// Token is a single lexical element from PDF input.
type Token struct {
	Kind  TokenKind
	Value string // raw or decoded value as appropriate
	Int   int64  // for TokenInteger
	Float float64
}

// Tokenizer produces a sequence of PDF tokens from input.
type Tokenizer interface {
	// Next returns the next token. At EOF it returns a token with Kind TokenEOF.
	Next() (Token, error)
	// CurrentOffset returns the byte offset in the source of the next token (for stream body reads).
	CurrentOffset() int64
}

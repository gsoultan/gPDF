package syntax

// TokenKind identifies the type of a PDF token.
type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenInteger
	TokenReal
	TokenKeyword
	TokenName
	TokenLiteral
	TokenHex
	TokenLBracket
	TokenRBracket
	TokenLDict
	TokenRDict
	TokenComment
	TokenUnknown
)

// Token is a single lexical element from PDF input.
type Token struct {
	Kind  TokenKind
	Value string
	Int   int64
	Float float64
}

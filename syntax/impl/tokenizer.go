package impl

import (
	"bufio"
	"bytes"
	"io"
	"strconv"

	"github.com/gsoultan/gpdf/syntax"
)

// tokenizer produces PDF tokens from a reader. It supports reset to a new position for random access.
type tokenizer struct {
	r   io.ReaderAt
	pos int64
	end int64
	buf *bufio.Reader
}

// NewTokenizer returns a tokenizer that reads from r starting at start, up to size bytes.
func NewTokenizer(r io.ReaderAt, start, size int64) syntax.Tokenizer {
	t := &tokenizer{r: r, pos: start, end: start + size}
	t.buf = bufio.NewReader(io.NewSectionReader(r, start, size))
	return t
}

// SetPosition sets the current read position (for Parser.SetPosition).
func (t *tokenizer) SetPosition(offset int64) {
	t.pos = offset
	t.buf = bufio.NewReader(io.NewSectionReader(t.r, offset, t.end-offset))
}

// CurrentOffset returns the byte offset of the next character to be read.
func (t *tokenizer) CurrentOffset() int64 {
	return t.pos
}

func (t *tokenizer) Next() (syntax.Token, error) {
	// Skip whitespace and comments
	for {
		b, err := t.buf.ReadByte()
		if err == io.EOF {
			return syntax.Token{Kind: syntax.TokenEOF}, nil
		}
		if err != nil {
			return syntax.Token{}, err
		}
		t.pos++
		if b == '%' {
			_, _ = t.buf.ReadBytes('\n')
			continue
		}
		if isWhitespace(b) {
			continue
		}
		t.buf.UnreadByte()
		t.pos--
		break
	}

	// Peek first byte
	peek, err := t.buf.Peek(1)
	if err == io.EOF {
		return syntax.Token{Kind: syntax.TokenEOF}, nil
	}
	if err != nil {
		return syntax.Token{}, err
	}
	first := peek[0]

	switch first {
	case '/':
		return t.readName()
	case '(':
		return t.readLiteralString()
	case '<':
		peek2, _ := t.buf.Peek(2)
		if len(peek2) == 2 && peek2[1] == '<' {
			t.buf.Discard(2)
			t.pos += 2
			return syntax.Token{Kind: syntax.TokenLDict, Value: "<<"}, nil
		}
		return t.readHexString()
	case '>':
		peek2, _ := t.buf.Peek(2)
		if len(peek2) == 2 && peek2[1] == '>' {
			t.buf.Discard(2)
			t.pos += 2
			return syntax.Token{Kind: syntax.TokenRDict, Value: ">>"}, nil
		}
		return syntax.Token{}, nil
	case '[':
		t.buf.Discard(1)
		t.pos++
		return syntax.Token{Kind: syntax.TokenLBracket}, nil
	case ']':
		t.buf.Discard(1)
		t.pos++
		return syntax.Token{Kind: syntax.TokenRBracket}, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', '.':
		return t.readNumber()
	default:
		return t.readKeywordOrRef()
	}
}

func isWhitespace(b byte) bool {
	return b == 0 || b == '\t' || b == '\n' || b == '\r' || b == ' '
}

func (t *tokenizer) readName() (syntax.Token, error) {
	t.buf.Discard(1) // /
	t.pos++
	var buf bytes.Buffer
	for {
		b, err := t.buf.ReadByte()
		if err != nil {
			return syntax.Token{Kind: syntax.TokenName, Value: buf.String()}, nil
		}
		t.pos++
		if isDelimiter(b) || isWhitespace(b) {
			t.buf.UnreadByte()
			t.pos--
			break
		}
		if b == '#' && buf.Len() > 0 {
			hex, err := t.buf.Peek(2)
			if err != nil || len(hex) < 2 {
				buf.WriteByte(b)
				continue
			}
			if isHex(hex[0]) && isHex(hex[1]) {
				t.buf.Discard(2)
				t.pos += 2
				v := hexByte(hex[0], hex[1])
				buf.WriteByte(v)
				continue
			}
		}
		buf.WriteByte(b)
	}
	return syntax.Token{Kind: syntax.TokenName, Value: buf.String()}, nil
}

func isDelimiter(b byte) bool {
	return b == '/' || b == '(' || b == ')' || b == '<' || b == '>' || b == '[' || b == ']' || b == '{' || b == '}'
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

func hexByte(a, b byte) byte {
	return (hexVal(a) << 4) | hexVal(b)
}

func hexVal(b byte) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	}
	if b >= 'A' && b <= 'F' {
		return b - 'A' + 10
	}
	if b >= 'a' && b <= 'f' {
		return b - 'a' + 10
	}
	return 0
}

func (t *tokenizer) readLiteralString() (syntax.Token, error) {
	t.buf.Discard(1) // (
	t.pos++
	var buf bytes.Buffer
	depth := 1
	for depth > 0 {
		b, err := t.buf.ReadByte()
		if err != nil {
			break
		}
		t.pos++
		switch b {
		case '\\':
			n, _ := t.buf.ReadByte()
			t.pos++
			switch n {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(', ')', '\\':
				buf.WriteByte(n)
			case '\n', '\r':
				// line continuation
			default:
				if n >= '0' && n <= '7' {
					oct := []byte{n}
					for i := 0; i < 2; i++ {
						p, _ := t.buf.Peek(1)
						if len(p) > 0 && p[0] >= '0' && p[0] <= '7' {
							c, _ := t.buf.ReadByte()
							t.pos++
							oct = append(oct, c)
						}
					}
					if len(oct) >= 2 {
						var v byte
						for _, c := range oct {
							v = v*8 + (c - '0')
						}
						buf.WriteByte(v)
					}
				}
			}
		case '(':
			depth++
			buf.WriteByte(b)
		case ')':
			depth--
			if depth > 0 {
				buf.WriteByte(b)
			}
		default:
			buf.WriteByte(b)
		}
	}
	return syntax.Token{Kind: syntax.TokenLiteral, Value: buf.String()}, nil
}

func (t *tokenizer) readHexString() (syntax.Token, error) {
	t.buf.Discard(1) // <
	t.pos++
	var buf bytes.Buffer
	for {
		b, err := t.buf.ReadByte()
		if err != nil || b == '>' {
			if b == '>' {
				t.pos++
			}
			break
		}
		t.pos++
		if isWhitespace(b) {
			continue
		}
		t.buf.UnreadByte()
		t.pos--
		peek, _ := t.buf.Peek(2)
		if len(peek) < 2 {
			break
		}
		if isHex(peek[0]) && isHex(peek[1]) {
			t.buf.Discard(2)
			t.pos += 2
			buf.WriteByte(hexByte(peek[0], peek[1]))
		} else {
			t.buf.ReadByte()
			t.pos++
		}
	}
	return syntax.Token{Kind: syntax.TokenHex, Value: buf.String()}, nil
}

func (t *tokenizer) readNumber() (syntax.Token, error) {
	var buf bytes.Buffer
	hasDot := false
	for {
		b, err := t.buf.ReadByte()
		if err != nil {
			break
		}
		t.pos++
		if (b >= '0' && b <= '9') || b == '+' || b == '-' {
			buf.WriteByte(b)
			continue
		}
		if b == '.' {
			buf.WriteByte(b)
			hasDot = true
			continue
		}
		t.buf.UnreadByte()
		t.pos--
		break
	}
	s := buf.String()
	if hasDot {
		f, _ := strconv.ParseFloat(s, 64)
		return syntax.Token{Kind: syntax.TokenReal, Value: s, Float: f}, nil
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return syntax.Token{Kind: syntax.TokenInteger, Value: s, Int: i}, nil
}

var keywords = map[string]syntax.TokenKind{
	"obj":       syntax.TokenKeyword,
	"endobj":    syntax.TokenKeyword,
	"stream":    syntax.TokenKeyword,
	"endstream": syntax.TokenKeyword,
	"xref":      syntax.TokenKeyword,
	"trailer":   syntax.TokenKeyword,
	"startxref": syntax.TokenKeyword,
	"R":         syntax.TokenKeyword,
	"true":      syntax.TokenKeyword,
	"false":     syntax.TokenKeyword,
	"null":      syntax.TokenKeyword,
}

func (t *tokenizer) readKeywordOrRef() (syntax.Token, error) {
	var buf bytes.Buffer
	for {
		b, err := t.buf.ReadByte()
		if err != nil {
			break
		}
		t.pos++
		if isWhitespace(b) || isDelimiter(b) {
			t.buf.UnreadByte()
			t.pos--
			break
		}
		buf.WriteByte(b)
	}
	s := buf.String()
	if k, ok := keywords[s]; ok {
		return syntax.Token{Kind: k, Value: s}, nil
	}
	return syntax.Token{Kind: syntax.TokenUnknown, Value: s}, nil
}

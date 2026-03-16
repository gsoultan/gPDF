package impl

import (
	"bytes"
	"io"
	"strconv"
)

// contentTokenizer tokenizes PDF content stream bytes (numbers, names, strings, arrays, operators).
type contentTokenizer struct {
	r *bytes.Reader
}

func newContentTokenizer(data []byte) *contentTokenizer {
	return &contentTokenizer{r: bytes.NewReader(data)}
}

func isSpace(b byte) bool {
	return b == 0 || b == '\t' || b == '\n' || b == '\r' || b == ' '
}

func (t *contentTokenizer) skipWhitespaceAndComments() error {
	for {
		b, err := t.r.ReadByte()
		if err == io.EOF {
			return io.EOF
		}
		if err != nil {
			return err
		}
		if b == '%' {
			for {
				c, e := t.r.ReadByte()
				if e == io.EOF || c == '\n' || c == '\r' {
					break
				}
			}
			continue
		}
		if !isSpace(b) {
			_ = t.r.UnreadByte()
			return nil
		}
	}
}

func (t *contentTokenizer) Next() (ctoken, error) {
	if err := t.skipWhitespaceAndComments(); err != nil {
		return ctoken{kind: ctEOF}, nil
	}
	first, err := t.r.ReadByte()
	if err == io.EOF {
		return ctoken{kind: ctEOF}, nil
	}
	if err != nil {
		return ctoken{}, err
	}
	switch first {
	case '/':
		return t.readName()
	case '(':
		return t.readLiteralString()
	case '<':
		next, _ := t.r.ReadByte()
		_ = t.r.UnreadByte()
		if next == '<' {
			_, _ = t.r.ReadByte()
			return ctoken{kind: ctOp, value: "<<"}, nil
		}
		return t.readHexString()
	case '[':
		return ctoken{kind: ctLArray, value: "["}, nil
	case ']':
		return ctoken{kind: ctRArray, value: "]"}, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', '.':
		_ = t.r.UnreadByte()
		return t.readNumber()
	default:
		_ = t.r.UnreadByte()
		return t.readOperator()
	}
}

func (t *contentTokenizer) readName() (ctoken, error) {
	var buf bytes.Buffer
	for {
		b, err := t.r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ctoken{}, err
		}
		if isSpace(b) || b == '/' || b == '(' || b == ')' || b == '<' || b == '>' || b == '[' || b == ']' || b == '%' {
			_ = t.r.UnreadByte()
			break
		}
		if b == '#' && buf.Len() > 0 {
			hex := make([]byte, 2)
			if n, _ := t.r.Read(hex); n == 2 && isHexChar(hex[0]) && isHexChar(hex[1]) {
				buf.WriteByte(hexByte(hex[0], hex[1]))
				continue
			}
			_ = t.r.UnreadByte()
		}
		buf.WriteByte(b)
	}
	return ctoken{kind: ctName, value: buf.String()}, nil
}

func (t *contentTokenizer) readLiteralString() (ctoken, error) {
	var buf bytes.Buffer
	depth := 1
	for depth > 0 {
		b, err := t.r.ReadByte()
		if err != nil {
			break
		}
		switch b {
		case '\\':
			n, _ := t.r.ReadByte()
			switch n {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b', 'f':
				buf.WriteByte(byte(n))
			case '(', ')', '\\':
				buf.WriteByte(n)
			case '\n', '\r':
				// line continuation
			default:
				buf.WriteByte(n)
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
	return ctoken{kind: ctString, value: buf.String()}, nil
}

func (t *contentTokenizer) readHexString() (ctoken, error) {
	_, _ = t.r.ReadByte() // consume '<'
	var buf bytes.Buffer
	for {
		b, err := t.r.ReadByte()
		if err != nil || b == '>' {
			break
		}
		if isSpace(b) {
			continue
		}
		_ = t.r.UnreadByte()
		hex := make([]byte, 2)
		if n, _ := t.r.Read(hex); n == 2 && isHexChar(hex[0]) && isHexChar(hex[1]) {
			buf.WriteByte(hexByte(hex[0], hex[1]))
		}
	}
	return ctoken{kind: ctHex, value: buf.String()}, nil
}

func (t *contentTokenizer) readNumber() (ctoken, error) {
	var buf bytes.Buffer
	hasDot := false
	for {
		b, err := t.r.ReadByte()
		if err != nil {
			break
		}
		if (b >= '0' && b <= '9') || b == '+' || b == '-' {
			buf.WriteByte(b)
			continue
		}
		if b == '.' {
			buf.WriteByte(b)
			hasDot = true
			continue
		}
		_ = t.r.UnreadByte()
		break
	}
	s := buf.String()
	if hasDot {
		f, _ := strconv.ParseFloat(s, 64)
		return ctoken{kind: ctReal, value: s, fltVal: f}, nil
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return ctoken{kind: ctInteger, value: s, intVal: i}, nil
}

func (t *contentTokenizer) readOperator() (ctoken, error) {
	var buf bytes.Buffer
	for {
		b, err := t.r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ctoken{}, err
		}
		if isSpace(b) || b == '/' || b == '(' || b == ')' || b == '<' || b == '>' || b == '[' || b == ']' || b == '%' {
			_ = t.r.UnreadByte()
			break
		}
		buf.WriteByte(b)
	}
	return ctoken{kind: ctOp, value: buf.String()}, nil
}

func isHexChar(b byte) bool {
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

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
		next, err := t.r.ReadByte()
		if err != nil {
			return ctoken{kind: ctOp, value: "<"}, nil
		}
		if next == '<' {
			return ctoken{kind: ctLDict, value: "<<"}, nil
		}
		_ = t.r.UnreadByte()
		return t.readHexString()
	case '>':
		next, err := t.r.ReadByte()
		if err != nil || next != '>' {
			if err == nil {
				_ = t.r.UnreadByte()
			}
			return ctoken{kind: ctOp, value: ">"}, nil
		}
		return ctoken{kind: ctRDict, value: ">>"}, nil
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
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(', ')', '\\':
				buf.WriteByte(n)
			case '\n':
				// line continuation — consume optional following \r
				if next, e := t.r.ReadByte(); e == nil && next != '\r' {
					_ = t.r.UnreadByte()
				}
			case '\r':
				// line continuation — consume optional following \n
				if next, e := t.r.ReadByte(); e == nil && next != '\n' {
					_ = t.r.UnreadByte()
				}
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// octal escape \ddd (1–3 octal digits)
				val := int(n - '0')
				for range 2 {
					digit, e := t.r.ReadByte()
					if e != nil || digit < '0' || digit > '7' {
						if e == nil {
							_ = t.r.UnreadByte()
						}
						break
					}
					val = val*8 + int(digit-'0')
				}
				buf.WriteByte(byte(val & 0xFF))
			default:
				// PDF spec: unrecognised escape — ignore the backslash, keep the char
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
	var buf bytes.Buffer
	var pending byte
	hasPending := false
	for {
		b, err := t.r.ReadByte()
		if err != nil || b == '>' {
			break
		}
		if isSpace(b) {
			continue
		}
		if !isHexChar(b) {
			continue
		}
		if !hasPending {
			pending = hexVal(b) << 4
			hasPending = true
			continue
		}
		buf.WriteByte(pending | hexVal(b))
		hasPending = false
	}
	if hasPending {
		buf.WriteByte(pending)
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

// readUntilEI reads raw inline image data bytes up to (but not including) the EI operator.
// Per PDF spec, the ID operator is followed by a single whitespace byte, then raw image data,
// then whitespace + "EI" + whitespace (or EOF). We consume the single whitespace after ID,
// then scan for the EI terminator.
func (t *contentTokenizer) readUntilEI() ([]byte, error) {
	// Consume the single whitespace byte that follows ID.
	if b, err := t.r.ReadByte(); err == nil && !isSpace(b) {
		_ = t.r.UnreadByte()
	}
	// Read all remaining bytes and find the EI marker.
	// EI must be preceded by whitespace and followed by whitespace or EOF.
	var buf []byte
	for {
		b, err := t.r.ReadByte()
		if err != nil {
			// EOF — return what we have (malformed stream, but don't error)
			return buf, nil
		}
		buf = append(buf, b)
		// Check if the last bytes form <ws>EI<ws-or-end>
		n := len(buf)
		if n >= 3 && buf[n-2] == 'E' && buf[n-1] == 'I' && isSpace(buf[n-3]) {
			// Peek next byte to confirm EI is followed by whitespace or EOF
			next, nerr := t.r.ReadByte()
			if nerr != nil || isSpace(next) {
				// Valid EI found — strip the <ws>EI from the data
				return buf[:n-3], nil
			}
			// Not a real EI — put back and continue
			_ = t.r.UnreadByte()
		}
	}
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

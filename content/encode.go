package content

import (
	"bytes"
	"fmt"
	"io"

	"gpdf/model"
)

// Encoder writes content stream operators to PDF content stream syntax.
type Encoder struct {
	w *bytes.Buffer
}

// NewEncoder returns an encoder that writes to an internal buffer.
func NewEncoder() *Encoder {
	return &Encoder{w: &bytes.Buffer{}}
}

// WriteOp appends one operator (args then name) to the stream.
func (e *Encoder) WriteOp(op Op) error {
	for i, arg := range op.Args {
		if i > 0 {
			e.w.WriteByte(' ')
		}
		if err := e.writeObject(arg); err != nil {
			return err
		}
	}
	e.w.WriteByte(' ')
	e.w.WriteString(op.Name)
	e.w.WriteByte('\n')
	return nil
}

// Bytes returns the encoded content stream.
func (e *Encoder) Bytes() []byte {
	return e.w.Bytes()
}

// Reset clears the buffer for reuse.
func (e *Encoder) Reset() {
	e.w.Reset()
}

func (e *Encoder) writeObject(obj model.Object) error {
	switch v := obj.(type) {
	case model.Integer:
		fmt.Fprintf(e.w, "%d", v)
	case model.Real:
		fmt.Fprintf(e.w, "%g", v)
	case model.String:
		writeLiteralString(e.w, string(v))
	case model.HexString:
		writeHexString(e.w, []byte(v))
	case model.Name:
		fmt.Fprintf(e.w, "/%s", escapeContentName(string(v)))
	case model.Array:
		e.w.WriteByte('[')
		for i, el := range v {
			if i > 0 {
				e.w.WriteByte(' ')
			}
			if err := e.writeObject(el); err != nil {
				return err
			}
		}
		e.w.WriteByte(']')
	case model.Dict:
		e.w.WriteString("<<")
		first := true
		for key, val := range v {
			if !first {
				e.w.WriteByte(' ')
			}
			first = false
			if err := e.writeObject(key); err != nil {
				return err
			}
			e.w.WriteByte(' ')
			if err := e.writeObject(val); err != nil {
				return err
			}
		}
		e.w.WriteString(">>")
	default:
		return fmt.Errorf("content encode: unsupported type %T", obj)
	}
	return nil
}

func writeLiteralString(w io.Writer, s string) {
	io.WriteString(w, "(")
	for i := range len(s) {
		c := s[i]
		switch c {
		case '\\', '(', ')':
			fmt.Fprintf(w, "\\%c", c)
		case '\n':
			io.WriteString(w, "\\n")
		case '\r':
			io.WriteString(w, "\\r")
		case '\t':
			io.WriteString(w, "\\t")
		default:
			w.Write([]byte{c})
		}
	}
	io.WriteString(w, ")")
}

func writeHexString(w io.Writer, b []byte) {
	io.WriteString(w, "<")
	for _, c := range b {
		fmt.Fprintf(w, "%02X", c)
	}
	io.WriteString(w, ">")
}

func escapeContentName(s string) string {
	var b bytes.Buffer
	for i := range len(s) {
		c := s[i]
		if c <= ' ' || c >= 127 || c == '#' || c == '/' || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '%' {
			fmt.Fprintf(&b, "#%02x", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// EncodeBytes encodes a sequence of operators to content stream bytes.
func EncodeBytes(ops []Op) ([]byte, error) {
	enc := NewEncoder()
	for _, op := range ops {
		if err := enc.WriteOp(op); err != nil {
			return nil, err
		}
	}
	return enc.Bytes(), nil
}

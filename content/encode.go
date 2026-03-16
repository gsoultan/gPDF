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
	default:
		return fmt.Errorf("content encode: unsupported type %T", obj)
	}
	return nil
}

func writeLiteralString(w io.Writer, s string) {
	io.WriteString(w, "(")
	for _, c := range s {
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
			io.WriteString(w, string(c))
		}
	}
	io.WriteString(w, ")")
}

func escapeContentName(s string) string {
	var b bytes.Buffer
	for _, c := range s {
		if c <= ' ' || c >= 127 || c == '#' || c == '/' || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '%' {
			fmt.Fprintf(&b, "#%02x", c)
		} else {
			b.WriteRune(c)
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

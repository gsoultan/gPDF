package asciihex

import (
	"encoding/hex"
	"io"
	"strings"

	"gpdf/stream"
)

// Filter implements stream.Filter for ASCIIHexDecode.
type Filter struct{}

func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "ASCIIHexDecode" {
		return nil
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	s := strings.Map(func(r rune) rune {
		switch {
		case r == ' ', r == '\t', r == '\r', r == '\n', r == '>':
			return -1
		default:
			return r
		}
	}, string(data))
	if len(s)%2 != 0 {
		s += "0"
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	_, err = dst.Write(decoded)
	return err
}

func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "ASCIIHexDecode" {
		return nil
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	encoded := strings.ToUpper(hex.EncodeToString(data))
	_, err = io.WriteString(dst, encoded+">")
	return err
}

// NewFilter returns an ASCIIHexDecode stream filter.
func NewFilter() stream.Filter { return Filter{} }

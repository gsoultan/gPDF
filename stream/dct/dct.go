package dct

import (
	"io"

	"gpdf/stream"
)

// Filter implements stream.Filter for DCTDecode (JPEG passthrough).
// JPEG data is already encoded; Decode copies as-is, Encode copies as-is.
// This allows the writer to recognize /Filter /DCTDecode without re-encoding.
type Filter struct{}

// Decode copies JPEG data through unchanged (JPEG decoding is left to the consumer).
func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "DCTDecode" {
		return nil
	}
	_, err := io.Copy(dst, src)
	return err
}

// Encode copies JPEG data through unchanged (data is already JPEG-encoded).
func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "DCTDecode" {
		return nil
	}
	_, err := io.Copy(dst, src)
	return err
}

// NewFilter returns a DCTDecode filter for JPEG passthrough.
func NewFilter() stream.Filter {
	return Filter{}
}

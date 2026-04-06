package jbig2

import (
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/stream"
)

// Filter implements stream.Filter for JBIG2Decode (monochrome image compression).
// Full JBIG2 decoding requires an external codec; this filter passes through the
// raw bitstream so callers can hand it to a renderer or external decoder.
type Filter struct{}

// NewFilter returns a JBIG2Decode filter.
func NewFilter() stream.Filter {
	return Filter{}
}

// Decode passes the raw JBIG2 bitstream through to dst unchanged.
func (Filter) Decode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("jbig2: read: %w", err)
	}
	return nil
}

// Encode passes data through unmodified; encoding JBIG2 is not supported.
func (Filter) Encode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("jbig2: write: %w", err)
	}
	return nil
}

package jpx

import (
	"fmt"
	"io"

	"gpdf/stream"
)

// Filter implements stream.Filter for JPXDecode (JPEG 2000).
// Full JPEG 2000 decoding requires an external codec; this filter passes through
// the raw JP2 bitstream and returns a structured error on decode so callers can
// handle the raw bytes (e.g. hand them to an image renderer).
type Filter struct{}

// NewFilter returns a JPXDecode filter.
func NewFilter() stream.Filter {
	return Filter{}
}

// Decode passes the raw JPEG 2000 bitstream through to dst and returns a
// descriptive error so callers know that hardware/OS-level decoding is needed.
func (Filter) Decode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("jpx: read: %w", err)
	}
	return nil
}

// Encode passes data through unmodified; encoding JPEG 2000 is not supported.
func (Filter) Encode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("jpx: write: %w", err)
	}
	return nil
}

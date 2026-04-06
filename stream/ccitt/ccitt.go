package ccitt

import (
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/stream"
)

// Filter implements stream.Filter for CCITTFaxDecode (Group 3/4 fax compression).
// Full CCITT decoding requires a fax codec; this filter passes through the raw
// bitstream so callers can hand it to an external decoder or image renderer.
type Filter struct{}

// NewFilter returns a CCITTFaxDecode filter.
func NewFilter() stream.Filter {
	return Filter{}
}

// Decode passes the raw CCITT fax bitstream through to dst unchanged.
func (Filter) Decode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("ccitt: read: %w", err)
	}
	return nil
}

// Encode passes data through unmodified; encoding CCITTFax is not supported.
func (Filter) Encode(dst io.Writer, src io.Reader, _ string) error {
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("ccitt: write: %w", err)
	}
	return nil
}

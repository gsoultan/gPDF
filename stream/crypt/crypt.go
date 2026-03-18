package crypt

import (
	"fmt"
	"io"

	"gpdf/stream"
)

// Filter implements stream.Filter for the Crypt filter (PDF 1.5+).
// The Crypt filter is a per-stream encryption selector. When the crypt filter
// name is "Identity" (or absent), the data passes through unchanged. Other
// named crypt filters delegate to the document security handler and are
// resolved at a higher level; this implementation handles the Identity case.
type Filter struct{}

// NewFilter returns a Crypt filter.
func NewFilter() stream.Filter {
	return Filter{}
}

// Decode passes data through for the Identity crypt filter.
func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "Crypt" {
		return fmt.Errorf("crypt: unexpected filter name: %s", name)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("crypt: read: %w", err)
	}
	return nil
}

// Encode passes data through for the Identity crypt filter.
func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "Crypt" {
		return fmt.Errorf("crypt: unexpected filter name: %s", name)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("crypt: write: %w", err)
	}
	return nil
}

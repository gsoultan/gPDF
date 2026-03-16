package stream

import "io"

// Filter decodes or encodes stream data (e.g. FlateDecode).
type Filter interface {
	// Decode decodes src and writes to dst. Name is the PDF filter name (e.g. "FlateDecode").
	Decode(dst io.Writer, src io.Reader, name string) error
	// Encode encodes src and writes to dst.
	Encode(dst io.Writer, src io.Reader, name string) error
}

// FilterRegistry maps PDF filter names to Filter implementations.
type FilterRegistry interface {
	// Get returns the filter for the given name, or nil if not supported.
	Get(name string) Filter
	// Register adds a filter for the given name.
	Register(name string, f Filter)
}

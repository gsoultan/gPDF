package flate

import (
	"compress/zlib"
	"io"

	"github.com/gsoultan/gpdf/stream"
)

// Filter implements stream.Filter for FlateDecode (zlib/deflate).
type Filter struct{}

// Decode decodes zlib-compressed data.
func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "FlateDecode" {
		return nil
	}
	r, err := zlib.NewReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(dst, r)
	return err
}

// Encode encodes data with zlib.
func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "FlateDecode" {
		return nil
	}
	w := zlib.NewWriter(dst)
	_, err := io.Copy(w, src)
	if err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// NewFilter returns a FlateDecode filter.
func NewFilter() stream.Filter {
	return Filter{}
}

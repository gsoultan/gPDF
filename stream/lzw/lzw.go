package lzw

import (
	clzw "compress/lzw"
	"io"

	"gpdf/stream"
)

// Filter implements stream.Filter for LZWDecode.
type Filter struct{}

func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "LZWDecode" {
		return nil
	}
	r := clzw.NewReader(src, clzw.MSB, 8)
	defer r.Close()
	_, err := io.Copy(dst, r)
	return err
}

func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "LZWDecode" {
		return nil
	}
	w := clzw.NewWriter(dst, clzw.MSB, 8)
	_, err := io.Copy(w, src)
	if err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// NewFilter returns an LZWDecode stream filter.
func NewFilter() stream.Filter { return Filter{} }

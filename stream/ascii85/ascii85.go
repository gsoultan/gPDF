package ascii85

import (
	a85 "encoding/ascii85"
	"io"

	"github.com/gsoultan/gpdf/stream"
)

// Filter implements stream.Filter for ASCII85Decode.
type Filter struct{}

func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "ASCII85Decode" {
		return nil
	}
	decoder := a85.NewDecoder(src)
	_, err := io.Copy(dst, decoder)
	return err
}

func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "ASCII85Decode" {
		return nil
	}
	encoder := a85.NewEncoder(dst)
	_, err := io.Copy(encoder, src)
	if err != nil {
		encoder.Close()
		return err
	}
	return encoder.Close()
}

// NewFilter returns an ASCII85Decode stream filter.
func NewFilter() stream.Filter { return Filter{} }

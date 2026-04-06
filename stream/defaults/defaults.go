package defaults

import (
	"github.com/gsoultan/gpdf/stream"
	"github.com/gsoultan/gpdf/stream/ascii85"
	"github.com/gsoultan/gpdf/stream/asciihex"
	"github.com/gsoultan/gpdf/stream/ccitt"
	"github.com/gsoultan/gpdf/stream/crypt"
	"github.com/gsoultan/gpdf/stream/dct"
	"github.com/gsoultan/gpdf/stream/flate"
	"github.com/gsoultan/gpdf/stream/jbig2"
	"github.com/gsoultan/gpdf/stream/jpx"
	"github.com/gsoultan/gpdf/stream/lzw"
	"github.com/gsoultan/gpdf/stream/runlength"
)

// RegisterStandardFilters registers all standard PDF filters in the given registry.
func RegisterStandardFilters(reg stream.FilterRegistry) {
	reg.Register("FlateDecode", flate.NewFilter())
	reg.Register("DCTDecode", dct.NewFilter())
	reg.Register("LZWDecode", lzw.NewFilter())
	reg.Register("ASCII85Decode", ascii85.NewFilter())
	reg.Register("ASCIIHexDecode", asciihex.NewFilter())
	reg.Register("RunLengthDecode", runlength.NewFilter())
	reg.Register("CCITTFaxDecode", ccitt.NewFilter())
	reg.Register("JBIG2Decode", jbig2.NewFilter())
	reg.Register("JPXDecode", jpx.NewFilter())
	reg.Register("Crypt", crypt.NewFilter())
}

// NewStandardRegistry returns a new filter registry with all standard filters pre-registered.
func NewStandardRegistry() stream.FilterRegistry {
	reg := stream.NewRegistry()
	RegisterStandardFilters(reg)
	return reg
}

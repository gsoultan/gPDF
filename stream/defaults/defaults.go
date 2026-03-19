package defaults

import (
	"gpdf/stream"
	"gpdf/stream/ascii85"
	"gpdf/stream/asciihex"
	"gpdf/stream/ccitt"
	"gpdf/stream/crypt"
	"gpdf/stream/dct"
	"gpdf/stream/flate"
	"gpdf/stream/jbig2"
	"gpdf/stream/jpx"
	"gpdf/stream/lzw"
	"gpdf/stream/runlength"
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

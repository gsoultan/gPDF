package syntax

import "github.com/gsoultan/gpdf/model"

// Parser parses PDF syntax into model objects and xref/trailer data.
type Parser interface {
	// ParseObject reads the next PDF object (including indirect object header and value).
	// Returns the indirect object if the input was "N M obj ... endobj", or nil and the inline object.
	ParseObject() (indirect *IndirectObject, inline model.Object, err error)
	// ParseXRefTable parses an xref table starting at current position. Call after locating "xref".
	ParseXRefTable() (entries map[int]XRefEntry, err error)
	// ParseTrailer parses a trailer dictionary. Call after "trailer" keyword.
	ParseTrailer() (dict model.Dict, err error)
	// SetPosition sets the read position (e.g. from startxref) for random access.
	SetPosition(offset int64) error
}

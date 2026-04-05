package reader

import "errors"

var (
	ErrMalformedXRef          = errors.New("malformed xref")
	ErrXRefRepairRequired     = errors.New("xref repair required")
	ErrMissingStartXRef       = errors.New("missing startxref")
	ErrUnsupportedFilter      = errors.New("unsupported stream filter")
	ErrStreamDecodeLimit      = errors.New("stream decode limit exceeded")
	ErrParseLimitExceeded     = errors.New("parse limit exceeded")
	ErrValidationFailed       = errors.New("validation failed")
	ErrInvalidDocumentGraph   = errors.New("invalid document object graph")
	ErrInvalidValidationLevel = errors.New("invalid validation level")
)

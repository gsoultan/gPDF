package sign

import (
	"crypto"
	"time"
)

// SignOptions controls PDF-level signature metadata and optional appearance.
type SignOptions struct {
	Reason      string
	Location    string
	Contact     string
	PageIndex   int
	Rect        [4]float64
	Hash        crypto.Hash
	SigningTime time.Time
}

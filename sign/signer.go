package sign

import (
	"crypto"
	"time"
)

// Signer produces a CMS/PKCS#7 (or similar) signature container over a given digest.
type Signer interface {
	Sign(digest []byte, hash crypto.Hash, signingTime time.Time) ([]byte, error)
}

package sign

import (
	"crypto"
	"io"
	"os"
	"time"
)

// Signer produces a CMS/PKCS#7 (or similar) signature container over a given digest.
// Implementations are responsible for constructing a DER-encoded structure that
// PDF viewers understand for /SubFilter values such as adbe.pkcs7.detached.
type Signer interface {
	// Sign takes the digest of the byte ranges, the hash algorithm that was used
	// to compute it, and the desired signing time. It returns a DER-encoded
	// CMS/PKCS#7 signature container.
	Sign(digest []byte, hash crypto.Hash, signingTime time.Time) ([]byte, error)
}

// SignOptions controls PDF-level signature metadata and optional appearance.
type SignOptions struct {
	// Reason is an optional human-readable reason for signing (e.g. "Approved").
	Reason string
	// Location is an optional physical location (e.g. "London, UK").
	Location string
	// Contact is an optional contact info string (e.g. email or phone).
	Contact string

	// PageIndex selects the page for a visible signature widget.
	// Use -1 to create an invisible signature (no widget annotation).
	PageIndex int
	// Rect is the visible signature rectangle [llx, lly, urx, ury] in points.
	// Ignored when PageIndex < 0.
	Rect [4]float64

	// Hash is the hash algorithm used for the byte-range digest.
	// When zero, implementations should default to crypto.SHA256.
	Hash crypto.Hash

	// SigningTime overrides the current time if non-zero.
	SigningTime time.Time
}

// Sign reads a PDF from r (size bytes) and writes a signed PDF to out.
// This is the low-level API; callers control how the original bytes are provided.
//
// NOTE: At this stage Sign acts as a pass-through (it copies the original PDF
// to out without modifying it). A later step will add the actual incremental
// update and signature objects.
func Sign(r io.ReaderAt, size int64, out io.Writer, s Signer, opts SignOptions) error {
	const chunk = 32 * 1024
	offset := int64(0)
	buf := make([]byte, chunk)
	for offset < size {
		n := chunk
		if remaining := size - offset; remaining < int64(chunk) {
			n = int(remaining)
		}
		if n <= 0 {
			break
		}
		if _, err := r.ReadAt(buf[:n], offset); err != nil && err != io.EOF {
			return err
		}
		if _, err := out.Write(buf[:n]); err != nil {
			return err
		}
		offset += int64(n)
	}
	return nil
}

// SignFile opens an existing PDF from path, applies signing via Sign, and writes
// the result to out. It is a convenience wrapper around Sign.
func SignFile(path string, out io.Writer, s Signer, opts SignOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	return Sign(f, info.Size(), out, s, opts)
}

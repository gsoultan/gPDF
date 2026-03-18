package sign

import (
	"io"
	"os"
)

// Sign reads a PDF from r (size bytes) and writes a signed PDF to out.
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
// the result to out.
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

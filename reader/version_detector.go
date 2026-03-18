package reader

import (
	"fmt"
	"io"
)

// PDFVersion holds the major and minor components of a PDF version header.
type PDFVersion struct {
	Major int
	Minor int
}

// String returns the version as a human-readable string, e.g. "2.0".
func (v PDFVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// IsAtLeast reports whether this version is greater than or equal to major.minor.
func (v PDFVersion) IsAtLeast(major, minor int) bool {
	if v.Major != major {
		return v.Major > major
	}
	return v.Minor >= minor
}

// detectPDFVersion reads the PDF header from the start of r and returns the version.
// It accepts headers of the form "%PDF-M.N" where M and N are single digits.
// Falls back to version 1.0 if the header cannot be parsed.
func detectPDFVersion(r io.ReaderAt) (PDFVersion, error) {
	const headerLen = 16
	buf := make([]byte, headerLen)
	if _, err := r.ReadAt(buf, 0); err != nil && err != io.EOF {
		return PDFVersion{}, fmt.Errorf("version detect: read header: %w", err)
	}
	return parseVersionHeader(buf), nil
}

// parseVersionHeader extracts Major and Minor from a raw PDF header buffer.
func parseVersionHeader(buf []byte) PDFVersion {
	const prefix = "%PDF-"
	if len(buf) < len(prefix)+3 {
		return defaultVersion()
	}
	for i, b := range []byte(prefix) {
		if buf[i] != b {
			return defaultVersion()
		}
	}
	offset := len(prefix)
	major, ok1 := digitAt(buf, offset)
	minor, ok2 := digitAt(buf, offset+2)
	if !ok1 || buf[offset+1] != '.' || !ok2 {
		return defaultVersion()
	}
	return PDFVersion{Major: major, Minor: minor}
}

func digitAt(buf []byte, i int) (int, bool) {
	if i >= len(buf) {
		return 0, false
	}
	b := buf[i]
	if b < '0' || b > '9' {
		return 0, false
	}
	return int(b - '0'), true
}

func defaultVersion() PDFVersion {
	return PDFVersion{Major: 1, Minor: 0}
}

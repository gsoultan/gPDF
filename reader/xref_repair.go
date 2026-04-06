package reader

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/gsoultan/gpdf/syntax"
)

const (
	xrefScanChunk   = 65536
	xrefScanOverlap = 64
	maxObjLineLen   = 32
)

// rebuildXRefByScan reconstructs the XRef table by scanning the entire file
// for "obj" keyword occurrences. This is used as a fallback when the startxref
// offset is corrupt or the XRef table/stream is damaged.
func rebuildXRefByScan(r io.ReaderAt, size int64) (map[int]syntax.XRefEntry, error) {
	entries := make(map[int]syntax.XRefEntry)
	if err := scanFileForObjects(r, size, entries); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("xref repair: no objects found during scan")
	}
	return entries, nil
}

// scanFileForObjects reads the file in overlapping chunks and records every
// "N G obj" header it finds, building XRefEntry records keyed by object number.
func scanFileForObjects(r io.ReaderAt, size int64, entries map[int]syntax.XRefEntry) error {
	objMarker := []byte(" obj")
	offset := int64(0)

	for offset < size {
		buf, readLen, err := readScanChunk(r, size, offset)
		if err != nil {
			return err
		}

		scanChunkForObjects(buf, offset, objMarker, entries)

		if readLen < xrefScanChunk {
			break
		}
		offset += int64(readLen) - xrefScanOverlap
	}
	return nil
}

// readScanChunk reads up to xrefScanChunk bytes from r at the given offset.
func readScanChunk(r io.ReaderAt, size, offset int64) ([]byte, int, error) {
	end := offset + xrefScanChunk
	if end > size {
		end = size
	}
	buf := make([]byte, end-offset)
	n, err := r.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, 0, fmt.Errorf("xref repair: read at %d: %w", offset, err)
	}
	return buf[:n], n, nil
}

// scanChunkForObjects finds "N G obj" patterns in buf and records valid entries.
func scanChunkForObjects(buf []byte, chunkOffset int64, objMarker []byte, entries map[int]syntax.XRefEntry) {
	pos := 0
	for pos < len(buf) {
		idx := bytes.Index(buf[pos:], objMarker)
		if idx < 0 {
			break
		}
		markerPos := pos + idx
		objNum, gen, ok := parseObjHeader(buf, markerPos)
		if ok {
			fileOffset := chunkOffset + int64(markerPos) - int64(len(strconv.Itoa(objNum))+1+len(strconv.Itoa(gen))+1)
			if _, exists := entries[objNum]; !exists {
				entries[objNum] = syntax.XRefEntry{
					Offset:     fileOffset,
					Generation: gen,
					InUse:      true,
				}
			}
		}
		pos = markerPos + len(objMarker)
	}
}

// parseObjHeader looks backward from markerPos to find "objNum generation" before " obj".
// Returns (objectNumber, generation, true) on success.
func parseObjHeader(buf []byte, markerPos int) (int, int, bool) {
	if markerPos <= 0 {
		return 0, 0, false
	}
	// scan backward to find the beginning of "objNum gen obj"
	start := markerPos - 1
	for start > 0 && start > markerPos-maxObjLineLen {
		if buf[start] == '\n' || buf[start] == '\r' {
			start++
			break
		}
		start--
	}
	if start < 0 {
		start = 0
	}
	line := string(bytes.TrimSpace(buf[start:markerPos]))
	fields := splitFields(line)
	if len(fields) < 2 {
		return 0, 0, false
	}
	objNum, err1 := strconv.Atoi(fields[len(fields)-2])
	gen, err2 := strconv.Atoi(fields[len(fields)-1])
	if err1 != nil || err2 != nil || objNum < 0 || gen < 0 {
		return 0, 0, false
	}
	return objNum, gen, true
}

// splitFields splits s on ASCII whitespace without allocating a slice of strings per field.
func splitFields(s string) []string {
	var fields []string
	start := -1
	for i := range len(s) {
		isSpace := s[i] == ' ' || s[i] == '\t' || s[i] == '\r' || s[i] == '\n'
		if !isSpace && start < 0 {
			start = i
		} else if isSpace && start >= 0 {
			fields = append(fields, s[start:i])
			start = -1
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}

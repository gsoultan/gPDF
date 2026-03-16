package impl

import (
	"bytes"
	"testing"
)

func TestParseXRefTable(t *testing.T) {
	// Minimal xref section: "xref\n0 4\n" + 4 lines of 20 bytes
	xrefSection := []byte("xref\n0 4\n" +
		"0000000000 65535 f \n" +
		"0000000009 00000 n \n" +
		"0000000065 00000 n \n" +
		"0000000122 00000 n \n" +
		"trailer\n")
	r := bytes.NewReader(xrefSection)
	p := NewParser(r, int64(len(xrefSection)))
	entries, err := p.ParseXRefTable()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
	e, ok := entries[1]
	if !ok {
		t.Fatal("expected entry 1")
	}
	if !e.InUse || e.Offset != 9 {
		t.Errorf("entry 1: InUse=%v Offset=%d", e.InUse, e.Offset)
	}
}

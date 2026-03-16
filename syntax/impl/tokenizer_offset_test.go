package impl

import (
	"bytes"
	"testing"
)

func TestTokenizer_CurrentOffset_AfterTokens(t *testing.T) {
	// "xref\n0 4\n" then 20-byte lines
	data := []byte("xref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n")
	r := bytes.NewReader(data)
	tok := NewTokenizer(r, 0, int64(len(data)))
	// Read "xref"
	_, _ = tok.Next()
	// Read "0"
	_, _ = tok.Next()
	// Read "4"
	_, _ = tok.Next()
	off := tok.CurrentOffset()
	// After "xref", "0", "4" the tokenizer may sit at last digit or delimiter; xref parser handles 5..8
	if off < 5 || off > 8 {
		t.Errorf("CurrentOffset after xref/0/4 = %d, want 5..8", off)
	}
}

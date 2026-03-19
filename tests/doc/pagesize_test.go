package doc_test

import (
	"gpdf/doc"
	"gpdf/doc/pagesize"
	"testing"
)

func TestPageSizeAPI(t *testing.T) {
	// 1. Using standard preset
	b1 := doc.New().A4()
	if b1.Err() != nil {
		t.Fatalf("A4() failed: %v", b1.Err())
	}
	// Note: internal state check would need access to unexported fields or a helper

	// 2. Using ApplyPageSize with preset
	b2 := doc.New().ApplyPageSize(pagesize.Letter)
	if b2.Err() != nil {
		t.Fatalf("ApplyPageSize(Letter) failed: %v", b2.Err())
	}

	// 3. Using ApplyPageSize with Landscape
	b3 := doc.New().ApplyPageSize(pagesize.A3.Landscape())
	if b3.Err() != nil {
		t.Fatalf("ApplyPageSize(A3.Landscape()) failed: %v", b3.Err())
	}

	// 4. Using Custom size
	b4 := doc.New().ApplyPageSize(pagesize.Custom(100, 200))
	if b4.Err() != nil {
		t.Fatalf("ApplyPageSize(Custom(100, 200)) failed: %v", b4.Err())
	}

	// 5. Chain with Landscape() method
	b5 := doc.New().A4().Landscape()
	if b5.Err() != nil {
		t.Fatalf("A4().Landscape() failed: %v", b5.Err())
	}
	w5, h5 := b5.GetDefaultPageSize()
	if w5 < h5 {
		t.Errorf("Expected landscape orientation (w > h), got w=%.2f, h=%.2f", w5, h5)
	}

	// 6. Chain with Portrait() method
	b6 := doc.New().A4().Landscape().Portrait()
	if b6.Err() != nil {
		t.Fatalf("A4().Landscape().Portrait() failed: %v", b6.Err())
	}
	w6, h6 := b6.GetDefaultPageSize()
	if w6 > h6 {
		t.Errorf("Expected portrait orientation (w < h), got w=%.2f, h=%.2f", w6, h6)
	}

	// 7. Idempotency check
	b7 := doc.New().A4().Landscape().Landscape()
	w7, h7 := b7.GetDefaultPageSize()
	if w7 < h7 {
		t.Errorf("Landscape() should be idempotent, got w=%.2f, h=%.2f", w7, h7)
	}

	b8 := doc.New().A4().Portrait().Portrait()
	w8, h8 := b8.GetDefaultPageSize()
	if w8 > h8 {
		t.Errorf("Portrait() should be idempotent, got w=%.2f, h=%.2f", w8, h8)
	}

	// 8. Size struct methods
	s := pagesize.A4
	ls := s.Landscape()
	if ls.Width < ls.Height {
		t.Errorf("Size.Landscape() failed: w=%.2f, h=%.2f", ls.Width, ls.Height)
	}
	ps := ls.Portrait()
	if ps.Width > ps.Height {
		t.Errorf("Size.Portrait() failed: w=%.2f, h=%.2f", ps.Width, ps.Height)
	}

	// 9. AddPage should use current default orientation
	b9 := doc.New().A4().Landscape().AddPage()
	if b9.PageWidth(0) < b9.PageHeight(0) {
		t.Errorf("Expected added page to be landscape, got w=%.2f, h=%.2f", b9.PageWidth(0), b9.PageHeight(0))
	}

	b10 := doc.New().A4().Landscape().Portrait().AddPage()
	if b10.PageWidth(0) > b10.PageHeight(0) {
		t.Errorf("Expected added page to be portrait, got w=%.2f, h=%.2f", b10.PageWidth(0), b10.PageHeight(0))
	}
}

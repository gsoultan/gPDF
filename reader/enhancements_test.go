package reader

import (
	"bytes"
	"errors"
	"testing"

	"gpdf/model"
	"gpdf/stream"
	"gpdf/stream/asciihex"
)

func TestStrictModeRejectsMissingStartXRef(t *testing.T) {
	pdf := minimalPDFBytes()
	pdf = bytes.Replace(pdf, []byte("startxref"), []byte("start____"), 1)

	r := NewPDFReaderWithOptions(ReaderOptions{Mode: ParserModeStrict})
	_, err := r.ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err == nil {
		t.Fatal("expected strict mode to fail when startxref is missing")
	}
	if !errors.Is(err, ErrMissingStartXRef) {
		t.Fatalf("expected ErrMissingStartXRef, got %v", err)
	}
}

func TestTolerantModeRecoversMissingStartXRef(t *testing.T) {
	pdf := minimalPDFBytes()
	pdf = bytes.Replace(pdf, []byte("startxref"), []byte("start____"), 1)

	r := NewPDFReaderWithOptions(ReaderOptions{Mode: ParserModeTolerant})
	doc, err := r.ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("expected tolerant mode to recover, got error: %v", err)
	}
	pages, err := doc.Pages()
	if err != nil {
		t.Fatalf("expected Pages to succeed, got error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
}

func TestApplyFiltersWithLimitsRejectsOversizedDecode(t *testing.T) {
	reg := stream.NewRegistry()
	reg.Register("ASCIIHexDecode", asciihex.NewFilter())

	_, err := applyFiltersWithLimits([]byte("41424344>"), model.Name("ASCIIHexDecode"), reg, 3, 2)
	if err == nil {
		t.Fatal("expected decode limit error")
	}
	if !errors.Is(err, ErrStreamDecodeLimit) {
		t.Fatalf("expected ErrStreamDecodeLimit, got %v", err)
	}
}

func TestValidateDocumentReportsStructuralStatus(t *testing.T) {
	pdf := minimalPDFBytes()
	r := NewPDFReader()
	doc, err := r.ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("ReadDocument failed: %v", err)
	}

	report, err := doc.Validate(ValidationStructural)
	if err != nil {
		t.Fatalf("Validate failed unexpectedly: %v", err)
	}
	if report.Level != ValidationStructural {
		t.Fatalf("expected level %d, got %d", ValidationStructural, report.Level)
	}
}

func TestValidateRejectsInvalidLevel(t *testing.T) {
	pdf := minimalPDFBytes()
	r := NewPDFReader()
	doc, err := r.ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("ReadDocument failed: %v", err)
	}

	_, err = doc.Validate(ValidationLevel(255))
	if err == nil {
		t.Fatal("expected invalid validation level error")
	}
	if !errors.Is(err, ErrInvalidValidationLevel) {
		t.Fatalf("expected ErrInvalidValidationLevel, got %v", err)
	}
}

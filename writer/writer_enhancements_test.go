package writer

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/gsoultan/gpdf/model"
)

type writerTestDoc struct {
	trailer model.Trailer
	objects map[int]model.Object
}

func (d *writerTestDoc) Trailer() model.Trailer {
	return d.trailer
}

func (d *writerTestDoc) Resolve(ref model.Ref) (model.Object, error) {
	obj, ok := d.objects[ref.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("object %d not found", ref.ObjectNumber)
	}
	return obj, nil
}

func (d *writerTestDoc) ObjectNumbers() []int {
	nums := make([]int, 0, len(d.objects))
	for num := range d.objects {
		nums = append(nums, num)
	}
	return nums
}

func newValidWriterTestDoc() *writerTestDoc {
	return &writerTestDoc{
		trailer: model.Trailer{Dict: model.Dict{
			model.Name("Root"): model.Ref{ObjectNumber: 1, Generation: 0},
			model.Name("Size"): model.Integer(2),
		}},
		objects: map[int]model.Object{
			1: model.Dict{
				model.Name("Type"): model.Name("Catalog"),
			},
		},
	}
}

func newDanglingWriterTestDoc() *writerTestDoc {
	return &writerTestDoc{
		trailer: model.Trailer{Dict: model.Dict{
			model.Name("Root"): model.Ref{ObjectNumber: 1, Generation: 0},
			model.Name("Size"): model.Integer(2),
		}},
		objects: map[int]model.Object{
			1: model.Dict{
				model.Name("Type"):  model.Name("Catalog"),
				model.Name("Pages"): model.Ref{ObjectNumber: 99, Generation: 0},
			},
		},
	}
}

func TestWriteRejectsDanglingReference(t *testing.T) {
	doc := newDanglingWriterTestDoc()
	var out bytes.Buffer

	err := NewPDFWriter().Write(&out, doc)
	if err == nil {
		t.Fatal("expected writer to reject dangling reference")
	}
	if !errors.Is(err, ErrInvalidDocumentGraph) {
		t.Fatalf("expected ErrInvalidDocumentGraph, got %v", err)
	}
}

func TestWriteWithPasswordDeterministicWithFixedRandomSource(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, 16)

	out1 := writeEncryptedWithSeed(t, newValidWriterTestDoc(), seed)
	out2 := writeEncryptedWithSeed(t, newValidWriterTestDoc(), seed)

	id1 := extractTrailerIDString(t, out1)
	id2 := extractTrailerIDString(t, out2)
	if id1 != id2 {
		t.Fatalf("expected deterministic trailer ID, got %q and %q", id1, id2)
	}
	if id1 != "BBBBBBBBBBBBBBBB" {
		t.Fatalf("expected fixed trailer ID from seed, got %q", id1)
	}
}

func writeEncryptedWithSeed(t *testing.T, doc Document, seed []byte) []byte {
	t.Helper()
	opts := DefaultWriterOptions()
	opts.RandomSource = bytes.NewReader(seed)
	pw := NewPDFWriterWithOptions(opts)

	var out bytes.Buffer
	if err := pw.WriteWithPassword(&out, doc, "user", "owner"); err != nil {
		t.Fatalf("WriteWithPassword failed: %v", err)
	}
	return out.Bytes()
}

func extractTrailerIDString(t *testing.T, output []byte) string {
	t.Helper()
	start := bytes.Index(output, []byte("/ID [("))
	if start < 0 {
		t.Fatalf("/ID marker not found in output")
	}
	start += len("/ID [(")
	end := bytes.Index(output[start:], []byte(")]"))
	if end < 0 {
		t.Fatalf("/ID closing marker not found in output")
	}
	return string(output[start : start+end])
}

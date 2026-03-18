package impl

import (
	"reflect"
	"testing"

	"gpdf/content"
	"gpdf/model"
)

func TestStreamParser_Parse(t *testing.T) {
	parser := NewStreamParser()

	t.Run("empty", func(t *testing.T) {
		ops, err := parser.Parse([]byte(""))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 0 {
			t.Errorf("expected 0 ops, got %d", len(ops))
		}
	})

	t.Run("single op no args", func(t *testing.T) {
		ops, err := parser.Parse([]byte("q"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 1 {
			t.Fatalf("expected 1 op, got %d", len(ops))
		}
		if ops[0].Name != "q" || len(ops[0].Args) != 0 {
			t.Errorf("op: name=%q args=%d", ops[0].Name, len(ops[0].Args))
		}
	})

	t.Run("cm with six numbers", func(t *testing.T) {
		ops, err := parser.Parse([]byte("1 0 0 1 0 0 cm"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 1 {
			t.Fatalf("expected 1 op, got %d", len(ops))
		}
		if ops[0].Name != "cm" || len(ops[0].Args) != 6 {
			t.Errorf("op: name=%q args=%d", ops[0].Name, len(ops[0].Args))
		}
	})

	t.Run("BT Tf Tj ET", func(t *testing.T) {
		ops, err := parser.Parse([]byte("BT /F1 12 Tf (Hello) Tj ET"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 4 {
			t.Fatalf("expected 4 ops, got %d", len(ops))
		}
		if ops[0].Name != "BT" {
			t.Errorf("op[0]: want BT, got %q", ops[0].Name)
		}
		if ops[1].Name != "Tf" {
			t.Errorf("op[1]: want Tf, got %q", ops[1].Name)
		}
		if len(ops[1].Args) != 2 {
			t.Errorf("Tf args: want 2, got %d", len(ops[1].Args))
		} else {
			if n, ok := ops[1].Args[0].(model.Name); !ok || n != "F1" {
				t.Errorf("Tf arg0: want name F1, got %v", ops[1].Args[0])
			}
			if r, ok := ops[1].Args[1].(model.Integer); !ok || r != 12 {
				t.Errorf("Tf arg1: want 12, got %v", ops[1].Args[1])
			}
		}
		if ops[2].Name != "Tj" || len(ops[2].Args) != 1 {
			t.Errorf("op[2]: want Tj with 1 arg, got %q len(args)=%d", ops[2].Name, len(ops[2].Args))
		} else if s, ok := ops[2].Args[0].(model.String); !ok || s != "Hello" {
			t.Errorf("Tj arg: want Hello, got %v", ops[2].Args[0])
		}
		if ops[3].Name != "ET" {
			t.Errorf("op[3]: want ET, got %q", ops[3].Name)
		}
	})

	t.Run("re and f", func(t *testing.T) {
		ops, err := parser.Parse([]byte("0 0 100 100 re f"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 2 {
			t.Fatalf("expected 2 ops, got %d", len(ops))
		}
		if ops[0].Name != "re" {
			t.Errorf("op[0]: want re, got %q", ops[0].Name)
		}
		if len(ops[0].Args) != 4 {
			t.Errorf("re args: want 4, got %d", len(ops[0].Args))
		}
		if ops[1].Name != "f" {
			t.Errorf("op[1]: want f, got %q", ops[1].Name)
		}
	})

	t.Run("array for TJ", func(t *testing.T) {
		ops, err := parser.Parse([]byte("[ (Hi) -10 ( ) ] TJ"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 1 {
			t.Fatalf("expected 1 op, got %d", len(ops))
		}
		if ops[0].Name != "TJ" {
			t.Errorf("op: want TJ, got %q", ops[0].Name)
		}
		if len(ops[0].Args) != 1 {
			t.Fatalf("TJ args: want 1 (array), got %d", len(ops[0].Args))
		}
		arr, ok := ops[0].Args[0].(model.Array)
		if !ok {
			t.Fatalf("TJ arg: want Array, got %T", ops[0].Args[0])
		}
		if len(arr) != 3 {
			t.Errorf("TJ array len: want 3, got %d", len(arr))
		}
	})

	t.Run("hex string odd nibble is padded", func(t *testing.T) {
		ops, err := parser.Parse([]byte("<414> Tj"))
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 1 {
			t.Fatalf("expected 1 op, got %d", len(ops))
		}
		if ops[0].Name != "Tj" || len(ops[0].Args) != 1 {
			t.Fatalf("op: want Tj with 1 arg, got %q len(args)=%d", ops[0].Name, len(ops[0].Args))
		}
		s, ok := ops[0].Args[0].(model.String)
		if !ok {
			t.Fatalf("Tj arg: want String, got %T", ops[0].Args[0])
		}
		if got, want := string(s), "A@"; got != want {
			t.Fatalf("hex string decode mismatch: got %q, want %q", got, want)
		}
	})
}

func TestStreamParser_ImplementsInterface(t *testing.T) {
	var _ content.Parser = (*StreamParser)(nil)
}

func TestStreamParser_InlineImage(t *testing.T) {
	// BI with a small inline image dict, ID, raw data bytes, then EI
	stream := []byte("q BI /W 2 /H 1 /CS /G /BPC 8 ID \xff\xfe \nEI Q")
	parser := NewStreamParser()
	ops, err := parser.Parse(stream)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	// Expect: q, BI, Q
	if len(ops) != 3 {
		t.Fatalf("expected 3 ops (q, BI, Q), got %d: %v", len(ops), ops)
	}
	if ops[0].Name != "q" {
		t.Errorf("op[0]: want q, got %q", ops[0].Name)
	}
	if ops[1].Name != "BI" {
		t.Errorf("op[1]: want BI, got %q", ops[1].Name)
	}
	if len(ops[1].Args) != 2 {
		t.Fatalf("BI args: want 2 (dict + data), got %d", len(ops[1].Args))
	}
	dict, ok := ops[1].Args[0].(model.Dict)
	if !ok {
		t.Fatalf("BI arg[0]: want Dict, got %T", ops[1].Args[0])
	}
	if dict[model.Name("W")] != model.Integer(2) {
		t.Errorf("BI dict /W: want 2, got %v", dict[model.Name("W")])
	}
	if ops[2].Name != "Q" {
		t.Errorf("op[2]: want Q, got %q", ops[2].Name)
	}
}

func TestStreamParser_OctalEscape(t *testing.T) {
	// \101 = 'A' (65), \012 = '\n' (10), \7 = 7
	stream := []byte("(\\101\\012\\7) Tj")
	parser := NewStreamParser()
	ops, err := parser.Parse(stream)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(ops) != 1 || ops[0].Name != "Tj" {
		t.Fatalf("expected 1 Tj op, got %d ops", len(ops))
	}
	s, ok := ops[0].Args[0].(model.String)
	if !ok {
		t.Fatalf("Tj arg: want String, got %T", ops[0].Args[0])
	}
	if got, want := string(s), "A\n\x07"; got != want {
		t.Errorf("octal escape: got %q, want %q", got, want)
	}
}

func TestStreamParser_LineContinuation(t *testing.T) {
	// \<newline> should be ignored (line continuation)
	stream := []byte("(hel\\\nlo) Tj")
	parser := NewStreamParser()
	ops, err := parser.Parse(stream)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	s, ok := ops[0].Args[0].(model.String)
	if !ok {
		t.Fatalf("Tj arg: want String, got %T", ops[0].Args[0])
	}
	if got, want := string(s), "hello"; got != want {
		t.Errorf("line continuation: got %q, want %q", got, want)
	}
}

// Ensure we can round-trip: parse then re-express (smoke test for operand types).
func TestStreamParser_OperandTypes(t *testing.T) {
	parser := NewStreamParser()
	ops, err := parser.Parse([]byte("/Helvetica 12 Tf (text) Tj"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(ops))
	}
	// Tf: Name, Integer
	if !reflect.DeepEqual(ops[0].Args[0], model.Name("Helvetica")) || !reflect.DeepEqual(ops[0].Args[1], model.Integer(12)) {
		t.Errorf("Tf args: got %v", ops[0].Args)
	}
	// Tj: String
	if !reflect.DeepEqual(ops[1].Args[0], model.String("text")) {
		t.Errorf("Tj arg: got %v", ops[1].Args[0])
	}
}

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
}

func TestStreamParser_ImplementsInterface(t *testing.T) {
	var _ content.Parser = (*StreamParser)(nil)
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

package content

import (
	"strings"
	"testing"

	"gpdf/model"
)

func TestEncodeBytes(t *testing.T) {
	ops := []Op{
		{Name: "BT", Args: nil},
		{Name: "Tf", Args: []model.Object{model.Name("F1"), model.Real(12)}},
		{Name: "Td", Args: []model.Object{model.Real(100), model.Real(700)}},
		{Name: "Tj", Args: []model.Object{model.String("Hello")}},
		{Name: "ET", Args: nil},
	}
	b, err := EncodeBytes(ops)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "BT") || !strings.Contains(s, "ET") {
		t.Errorf("expected BT/ET in content: %q", s)
	}
	if !strings.Contains(s, "/F1") || !strings.Contains(s, "12") {
		t.Errorf("expected /F1 and 12: %q", s)
	}
	if !strings.Contains(s, "(Hello)") {
		t.Errorf("expected (Hello): %q", s)
	}
}

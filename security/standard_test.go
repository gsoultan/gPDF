package security

import (
	"testing"

	"github.com/gsoultan/gpdf/model"
)

func TestNewStandardDecryptor_InvalidFilter(t *testing.T) {
	d := model.Dict{
		model.Name("Filter"): model.Name("Unknown"),
		model.Name("R"):      model.Integer(2),
		model.Name("O"):      model.String(string(make([]byte, 32))),
		model.Name("U"):      model.String(string(make([]byte, 32))),
		model.Name("P"):      model.Integer(-4),
	}
	_, err := NewStandardDecryptor(d, nil, "user")
	if err == nil {
		t.Error("expected error for unsupported Filter")
	}
}

func TestNewStandardDecryptor_UnsupportedR(t *testing.T) {
	// R=7 is not defined by the PDF specification and must return an error.
	d := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(7),
		model.Name("O"):      model.String(string(make([]byte, 32))),
		model.Name("U"):      model.String(string(make([]byte, 32))),
		model.Name("P"):      model.Integer(-4),
	}
	_, err := NewStandardDecryptor(d, nil, "user")
	if err == nil {
		t.Error("expected error for unsupported R")
	}
}

func TestNewStandardDecryptor_MissingO(t *testing.T) {
	d := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(2),
		model.Name("P"):      model.Integer(-4),
	}
	_, err := NewStandardDecryptor(d, nil, "user")
	if err == nil {
		t.Error("expected error for missing O")
	}
}

func TestDecryptor_DecryptString(t *testing.T) {
	o := make([]byte, 32)
	for i := range o {
		o[i] = byte(i)
	}
	d := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(2),
		model.Name("O"):      model.String(string(o)),
		model.Name("U"):      model.String(string(make([]byte, 32))),
		model.Name("P"):      model.Integer(-4),
	}
	dec, err := NewStandardDecryptor(d, nil, "user")
	if err != nil {
		t.Fatal(err)
	}
	ref := model.Ref{ObjectNumber: 1, Generation: 0}
	out, err := dec.DecryptString(ref, []byte("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 {
		t.Errorf("len: got %d", len(out))
	}
}

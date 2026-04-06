package xobject

import (
	"testing"

	"github.com/gsoultan/gpdf/model"
)

func TestIsImageXObject(t *testing.T) {
	s := &model.Stream{
		Dict: model.Dict{
			model.Name("Type"):       model.Name("XObject"),
			model.Name("Subtype"):    model.Name("Image"),
			model.Name("Width"):      model.Integer(100),
			model.Name("Height"):     model.Integer(50),
			model.Name("ColorSpace"): model.Name("DeviceRGB"),
		},
		Content: []byte{},
	}
	if !IsImageXObject(s) {
		t.Error("expected IsImageXObject true")
	}
	img := &Image{Stream: s}
	if img.Width() != 100 || img.Height() != 50 {
		t.Errorf("Width=%d Height=%d", img.Width(), img.Height())
	}
	if img.BitsPerComponent() != 0 {
		t.Errorf("BitsPerComponent=%d", img.BitsPerComponent())
	}
	cs := img.ColorSpace()
	if cs == nil {
		t.Fatal("ColorSpace nil")
	}
	if n, ok := cs.(model.Name); !ok || n != "DeviceRGB" {
		t.Errorf("ColorSpace=%v", cs)
	}
}

func TestIsFormXObject(t *testing.T) {
	s := &model.Stream{
		Dict: model.Dict{
			model.Name("Type"):    model.Name("XObject"),
			model.Name("Subtype"): model.Name("Form"),
			model.Name("BBox"): model.Array{
				model.Integer(0), model.Integer(0),
				model.Integer(595), model.Integer(842),
			},
		},
		Content: []byte("q\nQ\n"),
	}
	if !IsFormXObject(s) {
		t.Error("expected IsFormXObject true")
	}
	f := &Form{Stream: s}
	bbox, ok := f.BBox()
	if !ok || len(bbox) != 4 {
		t.Errorf("BBox ok=%v len=%d", ok, len(bbox))
	}
}

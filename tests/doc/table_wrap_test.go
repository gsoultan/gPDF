package doc_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"gpdf/doc"
	"gpdf/doc/style"
)

func createTestImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestTableImageWrapping(t *testing.T) {
	b := doc.New().A4().AddPage()

	logoData := createTestImage(100, 100)

	text := "This is a long text that should wrap around the image. " +
		"We are testing the Square wrapping feature of the gPDF library. " +
		"The text should be narrower next to the image and then span the full width of the cell after the image height is exceeded. " +
		"This allows for more complex layouts within a single table cell, similar to how Microsoft Word handles images."

	for i := 0; i < 5; i++ {
		text += " More text to ensure we definitely exceed the image height and see the full width wrapping behavior."
	}

	tbl := b.BeginTable(0, 50, 750, 500, 0, 1)

	// Left wrap
	tbl.RowSpec(doc.TableCellSpec{
		Text: "SQUARE WRAP LEFT: " + text,
		Image: &doc.TableCellImageSpec{
			Raw:     logoData,
			WidthPt: 50, HeightPt: 50,
			Wrap:      style.ImageWrapSquare,
			Side:      style.ImageSideLeft,
			PaddingPt: 10,
		},
		Style: doc.CellStyle{FillColor: doc.ColorLightGray, HasFillColor: true},
	})

	// Right wrap
	tbl.RowSpec(doc.TableCellSpec{
		Text: "SQUARE WRAP RIGHT: " + text,
		Image: &doc.TableCellImageSpec{
			Raw:     logoData,
			WidthPt: 50, HeightPt: 50,
			Wrap:      style.ImageWrapSquare,
			Side:      style.ImageSideRight,
			PaddingPt: 10,
		},
	})

	// TopBottom (Default)
	tbl.RowSpec(doc.TableCellSpec{
		Text: "TOP-BOTTOM WRAP (DEFAULT): " + text,
		Image: &doc.TableCellImageSpec{
			Raw:     logoData,
			WidthPt: 50, HeightPt: 50,
			Wrap: style.ImageWrapTopBottom,
		},
		Style: doc.CellStyle{FillColor: doc.ColorLightGray, HasFillColor: true},
	})

	tbl.EndTable()

	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("Empty PDF generated")
	}
}

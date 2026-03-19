package image

import (
	"bytes"
	stdimage "image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
)

// DetectDimensions attempts to get pixel dimensions from PNG/JPEG data without full decode.
func DetectDimensions(data []byte) (w, h int, isJPEG, isPNG bool, err error) {
	config, format, err := stdimage.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, false, false, err
	}
	return config.Width, config.Height, format == "jpeg", format == "png", nil
}

// HasAlpha returns true if any pixel in img has an alpha value less than fully opaque.
func HasAlpha(img stdimage.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xFFFF {
				return true
			}
		}
	}
	return false
}

// DecodePNGToRaw decodes PNG data into raw pixel bytes suitable for a PDF image XObject.
// Returns the raw bytes, pixel dimensions, and the PDF color space name.
func DecodePNGToRaw(pngData []byte) (raw []byte, w, h int, colorSpace string, err error) {
	img, _, decErr := stdimage.Decode(bytes.NewReader(pngData))
	if decErr != nil {
		return nil, 0, 0, "", decErr
	}
	bounds := img.Bounds()
	w, h = bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return nil, w, h, "", nil
	}

	hasAlpha := false
	switch img.(type) {
	case *stdimage.NRGBA, *stdimage.RGBA, *stdimage.RGBA64, *stdimage.NRGBA64:
		hasAlpha = HasAlpha(img)
	}

	_, isGray := img.(*stdimage.Gray)
	colorSpace = "DeviceRGB"
	bytesPerPixel := 3
	if isGray && !hasAlpha {
		colorSpace = "DeviceGray"
		bytesPerPixel = 1
	}

	raw = make([]byte, w*h*bytesPerPixel)
	for py := range h {
		for px := range w {
			c := img.At(bounds.Min.X+px, bounds.Min.Y+py)
			off := (py*w + px) * bytesPerPixel
			if bytesPerPixel == 1 {
				g, _, _, _ := color.GrayModel.Convert(c).(color.Gray).RGBA()
				raw[off] = byte(g >> 8)
			} else {
				r, g, bl, _ := c.RGBA()
				raw[off] = byte(r >> 8)
				raw[off+1] = byte(g >> 8)
				raw[off+2] = byte(bl >> 8)
			}
		}
	}
	return raw, w, h, colorSpace, nil
}

// ImageInfo holds processed image data.
type ImageInfo struct {
	Raw               []byte
	WidthPx, HeightPx int
	ColorSpace        string
	IsJPEG            bool
	BitsPerComponent  int
}

// ProcessImage detects format and decodes if needed.
func ProcessImage(data []byte) (*ImageInfo, error) {
	if bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		raw, w, h, cs, err := DecodePNGToRaw(data)
		if err != nil {
			return nil, err
		}
		return &ImageInfo{
			Raw:              raw,
			WidthPx:          w,
			HeightPx:         h,
			ColorSpace:       cs,
			BitsPerComponent: 8,
			IsJPEG:           false,
		}, nil
	}

	if bytes.HasPrefix(data, []byte("\xff\xd8")) {
		w, h, isJPEG, _, err := DetectDimensions(data)
		if err != nil {
			return nil, err
		}
		if isJPEG {
			return &ImageInfo{
				Raw:              data,
				WidthPx:          w,
				HeightPx:         h,
				ColorSpace:       "DeviceRGB",
				BitsPerComponent: 8,
				IsJPEG:           true,
			}, nil
		}
	}

	// Fallback/unknown
	return &ImageInfo{
		Raw:              data,
		BitsPerComponent: 8,
		ColorSpace:       "DeviceRGB",
	}, nil
}

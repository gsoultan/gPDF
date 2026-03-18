package doc

import (
	"fmt"

	imgpkg "gpdf/doc/image"
	taggedpkg "gpdf/doc/tagged"
)

// DrawImage queues an image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// Raw is the decoded image stream; widthPx/heightPx and bitsPerComponent/colorSpace must match.
// colorSpace should be DeviceGray, DeviceRGB, or DeviceCMYK. Call after AddPage().
func (b *DocumentBuilder) DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
	})
	return b
}

// DrawJPEG queues a JPEG image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// jpegData is the raw JPEG file bytes (not decoded); widthPx/heightPx must match the JPEG dimensions.
// The image is stored with /Filter /DCTDecode (no re-encoding), resulting in much smaller files than DrawImage for photos.
// colorSpace should match the JPEG (typically DeviceRGB or DeviceGray).
func (b *DocumentBuilder) DrawJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string) *DocumentBuilder {
	if len(b.pc.pages) == 0 || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
	})
	return b
}

// DrawPNG decodes a PNG image and draws it on the last added page at (x, y) with display size (widthPt, heightPt).
// The PNG is decoded to raw RGB pixels; for photos consider DrawJPEG which is more space-efficient.
func (b *DocumentBuilder) DrawPNG(x, y, widthPt, heightPt float64, pngData []byte) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNG: %w", err))
		return b
	}
	if w == 0 || h == 0 {
		b.setErr(fmt.Errorf("DrawPNG: invalid dimensions %dx%d", w, h))
		return b
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
	})
	return b
}

// DrawCircularImage queues a raw image clipped to a circle on the last added page.
// The image is placed at (x, y) with display size (widthPt, heightPt) and clipped to a
// circle centered at (cx, cy) with the given radius (all in points).
func (b *DocumentBuilder) DrawCircularImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, cx, cy, radius float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
	return b
}

// DrawCircularJPEG queues a JPEG image clipped to a circle on the last added page.
// The image is placed at (x, y) with display size (widthPt, heightPt) and clipped to a
// circle centered at (cx, cy) with the given radius (all in points).
func (b *DocumentBuilder) DrawCircularJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, cx, cy, radius float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
	return b
}

// DrawCircularPNG decodes a PNG and draws it clipped to a circle on the last added page.
// The image is placed at (x, y) with display size (widthPt, heightPt) and clipped to a
// circle centered at (cx, cy) with the given radius (all in points).
func (b *DocumentBuilder) DrawCircularPNG(x, y, widthPt, heightPt float64, pngData []byte, cx, cy, radius float64) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawCircularPNG: %w", err))
		return b
	}
	if w == 0 || h == 0 {
		b.setErr(fmt.Errorf("DrawCircularPNG: invalid dimensions %dx%d", w, h))
		return b
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
	return b
}

// DrawImageWithOpacity queues a raw image drawn at (x,y) with the given opacity in [0,1].
// opacity=1 is fully opaque; opacity=0 is invisible.
func (b *DocumentBuilder) DrawImageWithOpacity(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, opacity float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
		Opacity: opacity,
	})
	return b
}

// DrawJPEGWithOpacity queues a JPEG image drawn at (x,y) with the given opacity in [0,1].
func (b *DocumentBuilder) DrawJPEGWithOpacity(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, opacity float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		Opacity: opacity,
	})
	return b
}

// DrawPNGWithOpacity decodes a PNG and draws it at (x,y) with the given opacity in [0,1].
func (b *DocumentBuilder) DrawPNGWithOpacity(x, y, widthPt, heightPt float64, pngData []byte, opacity float64) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNGWithOpacity: %w", err))
		return b
	}
	if w == 0 || h == 0 {
		b.setErr(fmt.Errorf("DrawPNGWithOpacity: invalid dimensions %dx%d", w, h))
		return b
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
		Opacity: opacity,
	})
	return b
}

// DrawJPEGRotated queues a JPEG image rotated clockwise by rotateDeg degrees,
// placed at (x,y) with display size (widthPt, heightPt).
func (b *DocumentBuilder) DrawJPEGRotated(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, rotateDeg float64) *DocumentBuilder {
	if len(b.pc.pages) == 0 || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		RotateDeg: rotateDeg,
	})
	return b
}

// DrawPNGRotated decodes a PNG and draws it rotated clockwise by rotateDeg degrees,
// placed at (x,y) with display size (widthPt, heightPt).
func (b *DocumentBuilder) DrawPNGRotated(x, y, widthPt, heightPt float64, pngData []byte, rotateDeg float64) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNGRotated: %w", err))
		return b
	}
	if w == 0 || h == 0 {
		b.setErr(fmt.Errorf("DrawPNGRotated: invalid dimensions %dx%d", w, h))
		return b
	}
	idx := len(b.pc.pages) - 1
	b.pc.pages[idx].ImageRuns = append(b.pc.pages[idx].ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
		RotateDeg: rotateDeg,
	})
	return b
}

// DrawTaggedFigure draws an image as a tagged Figure with alternative text for accessibility.
// alt is used as /Alt on the Figure structure element (required for PDF/UA). Call after AddPage().
func (b *DocumentBuilder) DrawTaggedFigure(pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, alt string) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(raw) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	ps.ImageRuns = append(ps.ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
		MCID: mcid, HasMCID: true,
	})
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)
	return b
}

// DrawTaggedJPEG draws a JPEG image as a tagged Figure with alternative text.
func (b *DocumentBuilder) DrawTaggedJPEG(pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, alt string) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	ps.ImageRuns = append(ps.ImageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		MCID: mcid, HasMCID: true,
	})
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)
	return b
}

// DrawTaggedPNG decodes a PNG and draws it as a tagged Figure with alternative text.
func (b *DocumentBuilder) DrawTaggedPNG(pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if !b.pc.validPageIndex(pageIndex) || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawTaggedPNG: %w", err))
		return b
	}
	if w == 0 || h == 0 {
		b.setErr(fmt.Errorf("DrawTaggedPNG: invalid dimensions %dx%d", w, h))
		return b
	}
	return b.DrawTaggedFigure(pageIndex, x, y, widthPt, heightPt, raw, w, h, 8, colorSpace, alt)
}

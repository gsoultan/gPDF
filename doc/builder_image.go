package doc

import (
	"fmt"

	imgpkg "gpdf/doc/image"
	taggedpkg "gpdf/doc/tagged"
)

func (b *DocumentBuilder) addImageRun(pageIndex int, run imageRun) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(run.Raw) == 0 {
		return b
	}
	if run.ColorSpace == "" {
		run.ColorSpace = "DeviceRGB"
	}
	if run.BitsPerComponent <= 0 {
		run.BitsPerComponent = 8
	}
	b.pc.pages[pageIndex].ImageRuns = append(b.pc.pages[pageIndex].ImageRuns, run)
	return b
}

func (b *DocumentBuilder) lastPageIndex() int {
	return len(b.pc.pages) - 1
}

// DrawImage queues an image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// Raw is the decoded image stream; widthPx/heightPx and bitsPerComponent/colorSpace must match.
// colorSpace should be DeviceGray, DeviceRGB, or DeviceCMYK. Call after AddPage().
func (b *DocumentBuilder) DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
	})
}

// DrawJPEG queues a JPEG image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// jpegData is the raw JPEG file bytes (not decoded); widthPx/heightPx must match the JPEG dimensions.
func (b *DocumentBuilder) DrawJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
	})
}

// DrawPNG decodes a PNG image and draws it on the last added page at (x, y) with display size (widthPt, heightPt).
func (b *DocumentBuilder) DrawPNG(x, y, widthPt, heightPt float64, pngData []byte) *DocumentBuilder {
	if b.err != nil || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNG: %w", err))
		return b
	}
	return b.DrawImage(x, y, widthPt, heightPt, raw, w, h, 8, colorSpace)
}

// DrawCircularImage queues a raw image clipped to a circle on the last added page.
func (b *DocumentBuilder) DrawCircularImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, cx, cy, radius float64) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
}

// DrawCircularJPEG queues a JPEG image clipped to a circle on the last added page.
func (b *DocumentBuilder) DrawCircularJPEG(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, cx, cy, radius float64) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
}

// DrawCircularPNG decodes a PNG and draws it clipped to a circle on the last added page.
func (b *DocumentBuilder) DrawCircularPNG(x, y, widthPt, heightPt float64, pngData []byte, cx, cy, radius float64) *DocumentBuilder {
	if b.err != nil || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawCircularPNG: %w", err))
		return b
	}
	return b.DrawCircularImage(x, y, widthPt, heightPt, raw, w, h, 8, colorSpace, cx, cy, radius)
}

// DrawImageWithOpacity queues a raw image drawn at (x,y) with the given opacity in [0,1].
func (b *DocumentBuilder) DrawImageWithOpacity(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, opacity float64) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
		Opacity: opacity,
	})
}

// DrawJPEGWithOpacity queues a JPEG image drawn at (x,y) with the given opacity in [0,1].
func (b *DocumentBuilder) DrawJPEGWithOpacity(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, opacity float64) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		Opacity: opacity,
	})
}

// DrawPNGWithOpacity decodes a PNG and draws it at (x,y) with the given opacity in [0,1].
func (b *DocumentBuilder) DrawPNGWithOpacity(x, y, widthPt, heightPt float64, pngData []byte, opacity float64) *DocumentBuilder {
	if b.err != nil || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNGWithOpacity: %w", err))
		return b
	}
	return b.DrawImageWithOpacity(x, y, widthPt, heightPt, raw, w, h, 8, colorSpace, opacity)
}

// DrawJPEGRotated queues a JPEG image rotated clockwise by rotateDeg degrees.
func (b *DocumentBuilder) DrawJPEGRotated(x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, rotateDeg float64) *DocumentBuilder {
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
		RotateDeg: rotateDeg,
	})
}

// DrawPNGRotated decodes a PNG and draws it rotated clockwise by rotateDeg degrees.
func (b *DocumentBuilder) DrawPNGRotated(x, y, widthPt, heightPt float64, pngData []byte, rotateDeg float64) *DocumentBuilder {
	if b.err != nil || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawPNGRotated: %w", err))
		return b
	}
	return b.addImageRun(b.lastPageIndex(), imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
		RotateDeg: rotateDeg,
	})
}

func (b *DocumentBuilder) addTaggedFigure(pageIndex int, run imageRun, alt string) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || len(run.Raw) == 0 {
		return b
	}
	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	run.MCID = mcid
	run.HasMCID = true
	b.addImageRun(pageIndex, run)
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)
	return b
}

// DrawTaggedFigure draws an image as a tagged Figure with alternative text for accessibility.
func (b *DocumentBuilder) DrawTaggedFigure(pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, alt string) *DocumentBuilder {
	return b.addTaggedFigure(pageIndex, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: bitsPerComponent, ColorSpace: colorSpace,
	}, alt)
}

// DrawTaggedJPEG draws a JPEG image as a tagged Figure with alternative text.
func (b *DocumentBuilder) DrawTaggedJPEG(pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, alt string) *DocumentBuilder {
	return b.addTaggedFigure(pageIndex, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, IsJPEG: true,
	}, alt)
}

// DrawTaggedPNG decodes a PNG and draws it as a tagged Figure with alternative text.
func (b *DocumentBuilder) DrawTaggedPNG(pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) *DocumentBuilder {
	if b.err != nil || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil {
		b.setErr(fmt.Errorf("DrawTaggedPNG: %w", err))
		return b
	}
	return b.DrawTaggedFigure(pageIndex, x, y, widthPt, heightPt, raw, w, h, 8, colorSpace, alt)
}

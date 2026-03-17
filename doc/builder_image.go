package doc

import (
	imgpkg "gpdf/doc/image"
	taggedpkg "gpdf/doc/tagged"
)

// DrawImage queues an image to be drawn on the last added page at (x, y) with display size (widthPt, heightPt).
// Raw is the decoded image stream; widthPx/heightPx and bitsPerComponent/colorSpace must match.
// colorSpace should be DeviceGray, DeviceRGB, or DeviceCMYK. Call after AddPage().
func (b *DocumentBuilder) DrawImage(x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) *DocumentBuilder {
	if len(b.pages) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	idx := len(b.pages) - 1
	b.pages[idx].imageRuns = append(b.pages[idx].imageRuns, imageRun{
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
	if len(b.pages) == 0 || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	idx := len(b.pages) - 1
	b.pages[idx].imageRuns = append(b.pages[idx].imageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, isJPEG: true,
	})
	return b
}

// DrawPNG decodes a PNG image and draws it on the last added page at (x, y) with display size (widthPt, heightPt).
// The PNG is decoded to raw RGB pixels; for photos consider DrawJPEG which is more space-efficient.
func (b *DocumentBuilder) DrawPNG(x, y, widthPt, heightPt float64, pngData []byte) *DocumentBuilder {
	if len(b.pages) == 0 || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil || w == 0 || h == 0 {
		return b
	}
	idx := len(b.pages) - 1
	b.pages[idx].imageRuns = append(b.pages[idx].imageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: w, HeightPx: h,
		BitsPerComponent: 8, ColorSpace: colorSpace,
	})
	return b
}

// DrawTaggedFigure draws an image as a tagged Figure with alternative text for accessibility.
// alt is used as /Alt on the Figure structure element (required for PDF/UA). Call after AddPage().
func (b *DocumentBuilder) DrawTaggedFigure(pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, alt string) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) || len(raw) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	if bitsPerComponent <= 0 {
		bitsPerComponent = 8
	}
	b.useTagged = true
	ps := &b.pages[pageIndex]
	mcid := ps.nextMCID
	ps.nextMCID++
	ps.imageRuns = append(ps.imageRuns, imageRun{
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
	if pageIndex < 0 || pageIndex >= len(b.pages) || len(jpegData) == 0 {
		return b
	}
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	b.useTagged = true
	ps := &b.pages[pageIndex]
	mcid := ps.nextMCID
	ps.nextMCID++
	ps.imageRuns = append(ps.imageRuns, imageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: colorSpace, isJPEG: true,
		MCID: mcid, HasMCID: true,
	})
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)
	return b
}

// DrawTaggedPNG decodes a PNG and draws it as a tagged Figure with alternative text.
func (b *DocumentBuilder) DrawTaggedPNG(pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) *DocumentBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) || len(pngData) == 0 {
		return b
	}
	raw, w, h, colorSpace, err := imgpkg.DecodePNGToRaw(pngData)
	if err != nil || w == 0 || h == 0 {
		return b
	}
	return b.DrawTaggedFigure(pageIndex, x, y, widthPt, heightPt, raw, w, h, 8, colorSpace, alt)
}

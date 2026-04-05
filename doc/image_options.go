package doc

import (
	imgpkg "gpdf/doc/image"
	taggedpkg "gpdf/doc/tagged"
)

// ImageOptions is a type alias for image.Options.
type ImageOptions = imgpkg.Options

// DrawImageWith queues an image using the options struct.
func (b *DocumentBuilder) DrawImageWith(opts ImageOptions) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 {
		return b
	}

	colorSpace := opts.ColorSpace
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	bpc := opts.BitsPerComponent
	if bpc <= 0 {
		bpc = 8
	}
	pageIndex := opts.PageIndex
	if pageIndex < 0 {
		pageIndex = len(b.pc.pages) - 1
	}
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}

	run := imageRun{
		X: opts.X, Y: opts.Y,
		WidthPt: opts.Width, HeightPt: opts.Height,
		Matrix: opts.Matrix,
		Raw:    opts.Data, WidthPx: opts.PixelWidth, HeightPx: opts.PixelHeight,
		BitsPerComponent: bpc, ColorSpace: colorSpace,
		IsJPEG: opts.IsJPEG, Opacity: opts.Opacity, RotateDeg: opts.RotateDeg,
		ClipCircle: opts.ClipCircle, ClipCX: opts.ClipCX, ClipCY: opts.ClipCY, ClipR: opts.ClipRadius,
		IsArtifact: opts.IsArtifact,
		HasMask:    opts.HasMask, Mask: opts.Mask, MaskWidth: opts.MaskWidth, MaskHeight: opts.MaskHeight,
	}
	b.pc.pages[pageIndex].ImageRuns = append(b.pc.pages[pageIndex].ImageRuns, run)
	return b
}

// DrawTaggedImageWith queues a tagged image using the options struct.
func (b *DocumentBuilder) DrawTaggedImageWith(opts ImageOptions) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	pageIndex := opts.PageIndex
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}

	colorSpace := opts.ColorSpace
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	bpc := opts.BitsPerComponent
	if bpc <= 0 {
		bpc = 8
	}

	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: opts.AltText})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)

	run := imageRun{
		X: opts.X, Y: opts.Y,
		WidthPt: opts.Width, HeightPt: opts.Height,
		Raw: opts.Data, WidthPx: opts.PixelWidth, HeightPx: opts.PixelHeight,
		BitsPerComponent: bpc, ColorSpace: colorSpace,
		IsJPEG: opts.IsJPEG, MCID: mcid, HasMCID: true,
	}
	ps.ImageRuns = append(ps.ImageRuns, run)
	return b
}

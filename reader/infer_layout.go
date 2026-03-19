package reader

import "math"

// InferImageLayouts attempts to determine alignment and text wrapping for images.
func InferImageLayouts(images []ImageInfo, blocks []TextBlock, pageWidth float64) []ImageInfo {
	if len(images) == 0 {
		return images
	}

	for i := range images {
		img := &images[i]

		// 1. Infer Alignment
		center := img.X + img.WidthPt/2
		if center < pageWidth*0.4 {
			img.Alignment = 0 // Left
		} else if center > pageWidth*0.6 {
			img.Alignment = 2 // Right
		} else {
			img.Alignment = 1 // Center
		}

		// 2. Infer Wrapping
		// Check if any text blocks are placed "beside" the image.
		hasTextBeside := false
		for _, b := range blocks {
			// Check vertical overlap
			imgTop := img.Y + img.HeightPt
			imgBottom := img.Y
			blockTop := b.Y + b.Height
			blockBottom := b.Y

			overlap := math.Min(imgTop, blockTop) - math.Max(imgBottom, blockBottom)
			if overlap > 5 { // significant overlap
				// Check if text is to the left or right of the image
				if b.X+b.Width < img.X+5 || b.X > img.X+img.WidthPt-5 {
					hasTextBeside = true
					break
				}
			}
		}

		if hasTextBeside {
			img.Wrap = 2 // Square
		} else {
			// If image takes up most of the width, it's probably TopBottom
			if img.WidthPt > pageWidth*0.7 {
				img.Wrap = 1 // TopBottom
			} else {
				img.Wrap = 0 // None/Inline (often behaves like TopBottom in flow)
			}
		}
	}

	return images
}

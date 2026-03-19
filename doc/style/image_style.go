package style

// ImageAlignment specifies horizontal alignment of an image.
type ImageAlignment int

const (
	ImageAlignLeft ImageAlignment = iota
	ImageAlignCenter
	ImageAlignRight
)

// ImageWrap specifies how text wraps around an image.
type ImageWrap int

const (
	// ImageWrapNone (or TopBottom) is the default: text begins below the image if it takes up full width.
	ImageWrapNone ImageWrap = iota
	ImageWrapTopBottom
	// ImageWrapSquare makes the text wrap around the image's rectangular boundary.
	ImageWrapSquare
	// ImageWrapTight and ImageWrapThrough are similar to Square but follow image contours more closely.
	ImageWrapTight
	ImageWrapThrough
)

// ImageSide defines which side of the content area the image is aligned to for wrapping.
type ImageSide int

const (
	ImageSideLeft ImageSide = iota
	ImageSideRight
)

// ImageStyle defines visual properties of an image.
type ImageStyle struct {
	Opacity    float64
	Rotation   float64
	ClipCircle bool
	ClipCX     float64
	ClipCY     float64
	ClipR      float64
	Border     LineStyle
}

// ImageLayout defines how an image is positioned and how it interacts with other elements.
type ImageLayout struct {
	Width  float64
	Height float64
	Align  ImageAlignment
	Wrap   ImageWrap
	Margin float64 // Margin around the image for text wrapping
}

func DefaultImageStyle() ImageStyle {
	return ImageStyle{
		Opacity: 1.0,
	}
}

func DefaultImageLayout() ImageLayout {
	return ImageLayout{
		Align: ImageAlignLeft,
		Wrap:  ImageWrapNone,
	}
}

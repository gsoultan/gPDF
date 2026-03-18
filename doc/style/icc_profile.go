package style

// ICCProfile holds an ICC color profile for embedding in the PDF.
type ICCProfile struct {
	Data      []byte
	N         int
	Alternate string
}

// SRGBProfile returns a minimal sRGB ICC profile descriptor.
func SRGBProfile() ICCProfile {
	return ICCProfile{
		N:         3,
		Alternate: "DeviceRGB",
	}
}

// NewICCProfile creates an ICCProfile from raw ICC data.
// n is the number of components: 1 for gray, 3 for RGB, 4 for CMYK.
func NewICCProfile(data []byte, n int) ICCProfile {
	alt := "DeviceRGB"
	switch n {
	case 1:
		alt = "DeviceGray"
	case 4:
		alt = "DeviceCMYK"
	}
	return ICCProfile{Data: data, N: n, Alternate: alt}
}

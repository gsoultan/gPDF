package doc

// ICCProfile holds an ICC color profile for embedding in the PDF.
// Use with SetDefaultICCProfile to apply ICC-based color management.
type ICCProfile struct {
	// Data is the raw ICC profile bytes.
	Data []byte
	// N is the number of color components (1=gray, 3=RGB, 4=CMYK).
	N int
	// Alternate is the fallback color space name (e.g. "DeviceRGB").
	Alternate string
}

// SRGBProfile returns a minimal sRGB ICC profile descriptor.
// When used with SetDefaultICCProfile, images and colors are tagged as sRGB.
// Note: this creates a reference to sRGB without embedding a full profile;
// for PDF/A compliance, embed an actual sRGB .icc file via NewICCProfile.
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

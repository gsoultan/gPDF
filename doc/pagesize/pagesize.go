package pagesize

// Size represents page dimensions in points (1/72 inch).
type Size struct {
	Width  float64
	Height float64
}

// Landscape returns the landscape version of the page size.
func (s Size) Landscape() Size {
	if s.Width < s.Height {
		return Size{Width: s.Height, Height: s.Width}
	}
	return s
}

// Portrait returns the portrait version of the page size.
func (s Size) Portrait() Size {
	if s.Width > s.Height {
		return Size{Width: s.Height, Height: s.Width}
	}
	return s
}

// Predefined page sizes (standard sizes used in Microsoft Word).
var (
	// Letter (8.5 x 11 in)
	Letter = Size{612, 792}

	// Legal (8.5 x 14 in)
	Legal = Size{612, 1008}

	// Tabloid (11 x 17 in)
	Tabloid = Size{792, 1224}

	// Ledger (17 x 11 in)
	Ledger = Size{1224, 792}

	// Statement (5.5 x 8.5 in)
	Statement = Size{396, 612}

	// Executive (7.25 x 10.5 in)
	Executive = Size{522, 756}

	// A3 (297 x 420 mm)
	A3 = Size{841.89, 1190.55}

	// A4 (210 x 297 mm)
	A4 = Size{595.28, 841.89}

	// A5 (148 x 210 mm)
	A5 = Size{419.53, 595.28}

	// B4 (JIS) (257 x 364 mm)
	B4JIS = Size{728.5, 1031.81}

	// B4 (ISO) (250 x 353 mm)
	B4ISO = Size{708.66, 1000.63}

	// B5 (JIS) (182 x 257 mm)
	B5JIS = Size{515.91, 728.5}

	// B5 (ISO) (176 x 250 mm)
	B5ISO = Size{498.9, 708.66}

	// Folio (8.5 x 13 in)
	Folio = Size{612, 936}

	// Quarto (215 x 275 mm)
	Quarto = Size{609.45, 779.53}

	// Note (8.5 x 11 in)
	Note = Size{612, 792}

	// Envelope #10 (4.125 x 9.5 in)
	Envelope10 = Size{297, 684}

	// Envelope DL (110 x 220 mm)
	EnvelopeDL = Size{311.81, 623.62}

	// Envelope C5 (162 x 229 mm)
	EnvelopeC5 = Size{459.21, 649.13}

	// Envelope B5 (176 x 250 mm)
	EnvelopeB5 = Size{498.9, 708.66}

	// Envelope Monarch (3.875 x 7.5 in)
	EnvelopeMonarch = Size{279, 540}

	// Envelope Personal (3.625 x 6.5 in)
	EnvelopePersonal = Size{261, 468}
)

// Custom returns a custom page size.
func Custom(width, height float64) Size {
	return Size{Width: width, Height: height}
}

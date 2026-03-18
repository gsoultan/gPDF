package reader

// PageSize captures the effective page dimensions after box, rotation, and user-unit normalization.
type PageSize struct {
	Width    float64
	Height   float64
	Rotation int
	Box      string
	UserUnit float64
}

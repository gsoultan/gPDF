package syntax

// XRefEntry describes one entry in the cross-reference table.
type XRefEntry struct {
	Offset    int64
	Generation int
	InUse     bool
}

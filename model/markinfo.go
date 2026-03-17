package model

// MarkInfo is the mark information dictionary (Catalog /MarkInfo) for tagged PDF.
// Keys: /Marked (bool), /UserProperties (bool).
type MarkInfo struct {
	Dict Dict
}

// Marked returns true if the document is tagged (structure tree present).
func (m *MarkInfo) Marked() bool {
	if m == nil || m.Dict == nil {
		return false
	}
	if v, ok := m.Dict[Name("Marked")].(Boolean); ok {
		return bool(v)
	}
	return false
}

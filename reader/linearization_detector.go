package reader

import (
	"io"

	"gpdf/model"
	"gpdf/syntax/impl"
)

// LinearizationInfo holds the parsed linearization dictionary from a linearized PDF.
// A linearized PDF has this as its first object to enable byte-range page access.
type LinearizationInfo struct {
	// Dict is the raw linearization dictionary (contains /Linearized, /L, /H, /O, /E, /N, /T).
	Dict model.Dict
	// FileLength is the /L entry: the length of the linearized file in bytes.
	FileLength int64
	// FirstPageObjectNumber is the /O entry: object number of the first page.
	FirstPageObjectNumber int
	// NumberOfPages is the /N entry: total page count.
	NumberOfPages int
}

// detectLinearization checks whether the PDF at r is linearized by inspecting
// the first indirect object. If the object's dictionary contains /Linearized,
// it returns the parsed LinearizationInfo; otherwise it returns nil.
func detectLinearization(r io.ReaderAt, size int64) (*LinearizationInfo, error) {
	p := impl.NewParser(r, size)
	if err := p.SetPosition(0); err != nil {
		return nil, nil //nolint:nilerr // non-fatal: treat as non-linearized
	}
	indirect, _, err := p.ParseObject()
	if err != nil || indirect == nil {
		return nil, nil
	}
	dict, ok := indirect.Value.(model.Dict)
	if !ok {
		return nil, nil
	}
	if !isLinearized(dict) {
		return nil, nil
	}
	return buildLinearizationInfo(dict), nil
}

// isLinearized reports whether the dictionary has a /Linearized key.
func isLinearized(dict model.Dict) bool {
	_, ok := dict[model.Name("Linearized")]
	return ok
}

// buildLinearizationInfo extracts well-known fields from the linearization dictionary.
func buildLinearizationInfo(dict model.Dict) *LinearizationInfo {
	info := &LinearizationInfo{Dict: dict}
	info.FileLength = int64(intFromDict(dict, "L"))
	info.FirstPageObjectNumber = intFromDict(dict, "O")
	info.NumberOfPages = intFromDict(dict, "N")
	return info
}

// intFromDict reads an integer value by name from a model.Dict, returning 0 if absent.
func intFromDict(dict model.Dict, name string) int {
	switch v := dict[model.Name(name)].(type) {
	case model.Integer:
		return int(v)
	case model.Real:
		return int(v)
	}
	return 0
}

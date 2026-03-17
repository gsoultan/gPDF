package model

// Boolean represents a PDF boolean.
type Boolean bool

func (Boolean) IsIndirect() bool { return false }

// Integer represents a PDF integer.
type Integer int64

func (Integer) IsIndirect() bool { return false }

// Real represents a PDF real number.
type Real float64

func (Real) IsIndirect() bool { return false }

// String represents a PDF string (literal or hex).
type String string

func (String) IsIndirect() bool { return false }

// Name represents a PDF name object (e.g. /Type, /Pages).
type Name string

func (Name) IsIndirect() bool { return false }

// Array is a PDF array of objects.
type Array []Object

func (Array) IsIndirect() bool { return false }

// Dict is a PDF dictionary (map from Name to Object).
type Dict map[Name]Object

func (Dict) IsIndirect() bool { return false }

// HexString represents raw bytes written as a PDF hex string <...> in content streams.
// Used for CID font text encoding where character codes are 2-byte glyph IDs.
type HexString []byte

func (HexString) IsIndirect() bool { return false }

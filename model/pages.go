package model

// Pages represents a page tree node (pages dictionary).
// Keys: /Type (/Pages), /Kids, /Count, /Parent.
type Pages struct {
	Dict Dict
}

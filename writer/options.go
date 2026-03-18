package writer

import (
	"crypto/rand"
	"io"

	"gpdf/stream"
)

type WriterOptions struct {
	Filters       stream.FilterRegistry
	RandomSource  io.Reader
	ValidateGraph *bool
}

func DefaultWriterOptions() WriterOptions {
	return WriterOptions{
		RandomSource:  rand.Reader,
		ValidateGraph: boolRef(true),
	}
}

func normalizeWriterOptions(options WriterOptions) WriterOptions {
	defaults := DefaultWriterOptions()
	if options.RandomSource == nil {
		options.RandomSource = defaults.RandomSource
	}
	if options.ValidateGraph == nil {
		options.ValidateGraph = defaults.ValidateGraph
	}
	return options
}

func boolRef(v bool) *bool {
	return &v
}

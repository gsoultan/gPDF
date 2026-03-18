package reader

type ParserMode uint8

const (
	ParserModeTolerant ParserMode = iota
	ParserModeStrict
)

const (
	defaultMaxObjectCount       = 1_000_000
	defaultMaxDecodedStreamSize = 256 * 1024 * 1024
	defaultMaxFilterChainLength = 16
)

type ParseLimits struct {
	MaxObjectCount       int
	MaxDecodedStreamSize int
	MaxFilterChainLength int
}

type ReaderOptions struct {
	Mode             ParserMode
	Limits           ParseLimits
	ObjectCacheLimit int
}

func DefaultReaderOptions() ReaderOptions {
	return ReaderOptions{
		Mode: ParserModeTolerant,
		Limits: ParseLimits{
			MaxObjectCount:       defaultMaxObjectCount,
			MaxDecodedStreamSize: defaultMaxDecodedStreamSize,
			MaxFilterChainLength: defaultMaxFilterChainLength,
		},
		ObjectCacheLimit: defaultObjectCacheLimit,
	}
}

func normalizeReaderOptions(opts ReaderOptions) ReaderOptions {
	defaults := DefaultReaderOptions()
	if opts.Mode != ParserModeStrict {
		opts.Mode = defaults.Mode
	}
	if opts.Limits.MaxObjectCount <= 0 {
		opts.Limits.MaxObjectCount = defaults.Limits.MaxObjectCount
	}
	if opts.Limits.MaxDecodedStreamSize <= 0 {
		opts.Limits.MaxDecodedStreamSize = defaults.Limits.MaxDecodedStreamSize
	}
	if opts.Limits.MaxFilterChainLength <= 0 {
		opts.Limits.MaxFilterChainLength = defaults.Limits.MaxFilterChainLength
	}
	if opts.ObjectCacheLimit <= 0 {
		opts.ObjectCacheLimit = defaults.ObjectCacheLimit
	}
	return opts
}

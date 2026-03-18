package reader

// CodeGenOptions controls how GenerateCode reconstructs an existing PDF into Go source.
type CodeGenOptions struct {
	PackageName           string
	FunctionName          string
	EmbedImages           bool
	PreservePageSize      bool
	PreserveTables        bool
	PreserveTextStyles    bool
	PreservePositions     bool
	InlineImageLimit      int
	MaxDecodedStreamBytes int
	MaxImageBytes         int
	MaxOpsPerPage         int
}

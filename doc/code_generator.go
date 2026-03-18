package doc

import (
	"io"

	"gpdf/reader"
)

// CodeGenOptions controls how GenerateCode reconstructs an existing PDF into Go source.
type CodeGenOptions = reader.CodeGenOptions

// GeneratedAsset is an optional binary asset emitted alongside generated Go source.
type GeneratedAsset = reader.GeneratedAsset

// GeneratedCode contains Go source and optional assets emitted by GenerateCode.
type GeneratedCode = reader.GeneratedCode

// CodeGenerator reconstructs an existing PDF into Go builder source.
type CodeGenerator interface {
	GenerateCode(opts CodeGenOptions) (GeneratedCode, error)
	GenerateCodeTo(w io.Writer, opts CodeGenOptions) ([]GeneratedAsset, error)
}

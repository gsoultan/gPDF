package reader

import "io"

// CodeGenerator reconstructs an existing PDF into Go builder source.
type CodeGenerator interface {
	GenerateCode(opts CodeGenOptions) (GeneratedCode, error)
	GenerateCodeTo(w io.Writer, opts CodeGenOptions) ([]GeneratedAsset, error)
}

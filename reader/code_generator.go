package reader

// CodeGenerator reconstructs an existing PDF into Go builder source.
type CodeGenerator interface {
	GenerateCode(opts CodeGenOptions) (GeneratedCode, error)
}

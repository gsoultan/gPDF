package reader

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// ValidateCodeGenOptions validates user-provided code generation options.
func ValidateCodeGenOptions(opts CodeGenOptions) error {
	if opts.PackageName != "" && !isValidGoIdentifier(opts.PackageName) {
		return fmt.Errorf("%w: invalid package name %q", ErrInvalidCodeGenOptions, opts.PackageName)
	}
	if opts.FunctionName != "" && !isValidGoIdentifier(opts.FunctionName) {
		return fmt.Errorf("%w: invalid function name %q", ErrInvalidCodeGenOptions, opts.FunctionName)
	}
	if opts.InlineImageLimit < 0 {
		return fmt.Errorf("%w: InlineImageLimit must be >= 0", ErrInvalidCodeGenOptions)
	}
	if opts.MaxDecodedStreamBytes < 0 {
		return fmt.Errorf("%w: MaxDecodedStreamBytes must be >= 0", ErrInvalidCodeGenOptions)
	}
	if opts.MaxImageBytes < 0 {
		return fmt.Errorf("%w: MaxImageBytes must be >= 0", ErrInvalidCodeGenOptions)
	}
	if opts.MaxOpsPerPage < 0 {
		return fmt.Errorf("%w: MaxOpsPerPage must be >= 0", ErrInvalidCodeGenOptions)
	}
	return nil
}

func isValidGoIdentifier(name string) bool {
	if name == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError || size == 0 {
		return false
	}
	if r != '_' && !unicode.IsLetter(r) {
		return false
	}
	for _, r = range name[size:] {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

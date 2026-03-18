package reader

import "errors"

type ValidationLevel uint8

const (
	ValidationSyntax ValidationLevel = iota
	ValidationStructural
	ValidationSemantic
)

type ValidationSeverity string

const (
	ValidationSeverityInfo  ValidationSeverity = "info"
	ValidationSeverityWarn  ValidationSeverity = "warn"
	ValidationSeverityError ValidationSeverity = "error"
)

type ValidationDiagnostic struct {
	Code         string
	Severity     ValidationSeverity
	Message      string
	ObjectNumber int
}

type ValidationReport struct {
	Level       ValidationLevel
	Diagnostics []ValidationDiagnostic
}

func (r ValidationReport) HasErrors() bool {
	for _, diag := range r.Diagnostics {
		if diag.Severity == ValidationSeverityError {
			return true
		}
	}
	return false
}

func (r ValidationReport) Error() error {
	if r.HasErrors() {
		return ErrValidationFailed
	}
	return nil
}

type ValidationProvider interface {
	Validate(level ValidationLevel) (ValidationReport, error)
}

func isValidationLevel(level ValidationLevel) bool {
	return level == ValidationSyntax || level == ValidationStructural || level == ValidationSemantic
}

func validationError(code string, message string, objectNumber int) ValidationDiagnostic {
	return ValidationDiagnostic{
		Code:         code,
		Severity:     ValidationSeverityError,
		Message:      message,
		ObjectNumber: objectNumber,
	}
}

func validationWarn(code string, message string, objectNumber int) ValidationDiagnostic {
	return ValidationDiagnostic{
		Code:         code,
		Severity:     ValidationSeverityWarn,
		Message:      message,
		ObjectNumber: objectNumber,
	}
}

func appendValidationError(errs []ValidationDiagnostic, code string, message string, objectNumber int) []ValidationDiagnostic {
	return append(errs, validationError(code, message, objectNumber))
}

func appendValidationWarn(errs []ValidationDiagnostic, code string, message string, objectNumber int) []ValidationDiagnostic {
	return append(errs, validationWarn(code, message, objectNumber))
}

func validationErrorOrNil(errs []ValidationDiagnostic) error {
	for _, diag := range errs {
		if diag.Severity == ValidationSeverityError {
			return ErrValidationFailed
		}
	}
	return nil
}

func joinValidationErrors(errs ...error) error {
	return errors.Join(errs...)
}

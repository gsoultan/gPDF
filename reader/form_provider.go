package reader

import "gpdf/model"

// FormProvider exposes the AcroForm interactive form fields.
type FormProvider interface {
	// AcroForm returns the AcroForm dictionary from the catalog, or nil if absent.
	AcroForm() (*model.AcroForm, error)
	// FormFields resolves and returns all root-level AcroForm field dictionaries.
	FormFields() ([]model.Dict, error)
}

package doc

// FormBuilder configures AcroForm and creates fields/widgets.
type FormBuilder interface {
	SetAcroForm() *DocumentBuilder
	SetAcroFormSigFlags(flags int) *DocumentBuilder
	AddTextField(pageIndex int, llx, lly, urx, ury float64, name, value, tooltip string, required bool) *DocumentBuilder
	AddCheckBox(pageIndex int, llx, lly, urx, ury float64, name string, checked bool, tooltip string, required bool) *DocumentBuilder
	AddRadioButton(pageIndex int, llx, lly, urx, ury float64, groupName, value string, checked bool, tooltip string) *DocumentBuilder
	AddSubmitButton(pageIndex int, llx, lly, urx, ury float64, name, label, submitURL, tooltip string) *DocumentBuilder
}

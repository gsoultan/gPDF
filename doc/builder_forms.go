package doc

import "gpdf/doc/form"

// AddTextField adds a text form field (/FT /Tx) with an associated widget annotation on the given page.
// The rectangle is in user space (llx, lly, urx, ury). name is the field name; value is the default value.
func (b *DocumentBuilder) AddTextField(pageIndex int, llx, lly, urx, ury float64, name, value, tooltip string, required bool) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || name == "" {
		return b
	}
	b.forms.UseAcroForm = true
	b.forms.Fields = append(b.forms.Fields, form.FieldSpec{
		PageIndex: pageIndex,
		Rect:      [4]float64{llx, lly, urx, ury},
		Name:      name,
		Kind:      form.FieldText,
		Value:     value,
		Tooltip:   tooltip,
		Required:  required,
	})
	return b
}

// AddCheckBox adds a checkbox form field (/FT /Btn) with a single widget annotation.
func (b *DocumentBuilder) AddCheckBox(pageIndex int, llx, lly, urx, ury float64, name string, checked bool, tooltip string, required bool) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || name == "" {
		return b
	}
	b.forms.UseAcroForm = true
	b.forms.Fields = append(b.forms.Fields, form.FieldSpec{
		PageIndex: pageIndex,
		Rect:      [4]float64{llx, lly, urx, ury},
		Name:      name,
		Kind:      form.FieldCheckbox,
		Checked:   checked,
		Tooltip:   tooltip,
		Required:  required,
	})
	return b
}

// AddRadioButton adds a single radio button belonging to a logical group.
// Radio buttons in the same groupName share one /FT /Btn field with multiple widgets.
func (b *DocumentBuilder) AddRadioButton(pageIndex int, llx, lly, urx, ury float64, groupName, value string, checked bool, tooltip string) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || groupName == "" || value == "" {
		return b
	}
	b.forms.UseAcroForm = true
	b.forms.Fields = append(b.forms.Fields, form.FieldSpec{
		PageIndex: pageIndex,
		Rect:      [4]float64{llx, lly, urx, ury},
		Name:      value,
		Kind:      form.FieldRadio,
		GroupName: groupName,
		Checked:   checked,
		Tooltip:   tooltip,
	})
	return b
}

// AddSubmitButton adds a pushbutton with a simple SubmitForm action to the given URL.
func (b *DocumentBuilder) AddSubmitButton(pageIndex int, llx, lly, urx, ury float64, name, label, submitURL, tooltip string) *DocumentBuilder {
	if !b.pc.validPageIndex(pageIndex) || name == "" || submitURL == "" {
		return b
	}
	b.forms.UseAcroForm = true
	b.forms.Fields = append(b.forms.Fields, form.FieldSpec{
		PageIndex: pageIndex,
		Rect:      [4]float64{llx, lly, urx, ury},
		Name:      name,
		Kind:      form.FieldSubmit,
		Value:     label,
		SubmitURL: submitURL,
		Tooltip:   tooltip,
	})
	return b
}

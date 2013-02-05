/*
wtforms Package
Golang的表单包
提供表单定义,验证,HTML渲染
目前还不是很完善,不久后作为一个包开源
灵感来自Python的WTForms库,http://wtforms.simplecodes.com/
*/

package wtforms

import (
	"html/template"
	"net/http"
	"strings"
)

type Form struct {
	fields map[string]IField
}

func NewForm(fields ...IField) *Form {
	form := Form{}
	form.fields = make(map[string]IField)
	for _, field := range fields {
		form.fields[field.GetName()] = field
	}
	return &form
}

func (form *Form) Validate(r *http.Request) bool {
	result := true
	for name, field := range form.fields {
		field.SetValue(strings.TrimSpace(r.FormValue(name)))
		if !field.Validate() {
			result = false
		}
	}
	return result
}

func (form *Form) Render(name string, attrs ...string) template.HTML {
	field := form.fields[name]

	return field.RenderFull(attrs)
}

func (form *Form) RenderInput(name string, attrs ...string) template.HTML {
	field := form.fields[name]

	return field.RenderInput(attrs)
}

func (form *Form) Value(name string) string {
	return form.fields[name].GetValue()
}

func (form *Form) SetValue(name, value string) {
	form.fields[name].SetValue(value)
}

func (form *Form) AddError(name, err string) {
	field := form.fields[name]
	field.AddError(err)
}

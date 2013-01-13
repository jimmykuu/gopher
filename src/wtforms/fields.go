package wtforms

import (
    "fmt"
    "html/template"
)

func renderFull(field IField, attrs []string) template.HTML {
    errClass := ""
    if field.HasErrors() {
        errClass = " error"
    }
    return template.HTML(fmt.Sprintf(`<div class="control-group%s">%s<div class="controls">%s%s</div></div>`, errClass, field.RenderLabel(), field.RenderInput(attrs), field.RenderErrors()))
}

type IField interface {
    RenderLabel() template.HTML
    RenderInput(attrs []string) template.HTML
    Validate() bool
    GetName() string
    GetValue() string
    SetValue(value string)
    IsName(name string) bool
    RenderFull(attrs []string) template.HTML
    HasErrors() bool
    RenderErrors() string
    AddError(err string)
}

type BaseField struct {
    Name       string
    Label      string
    Value      string
    Errors     []string
    Validators []IValidator
}

func (field *BaseField) RenderLabel() template.HTML {
    return template.HTML(fmt.Sprintf("<label class=\"control-label\" for=\"%s\">%s</label>", field.Name, field.Label))
}

func (field *BaseField) HasErrors() bool {
    return len(field.Errors) > 0
}

func (field *BaseField) RenderInput(attrs []string) template.HTML {
    return template.HTML("")
}

func (field *BaseField) GetName() string {
    return field.Name
}

func (field *BaseField) AddError(err string) {
    field.Errors = append(field.Errors, err)
}

func (field *BaseField) RenderErrors() string {
    result := ""
    for _, err := range field.Errors {
        result += fmt.Sprintf(`<span class="help-block">%s</span>`, err)
    }

    return result
}

func (field *BaseField) Validate() bool {
    // 如果有Required并且输入为空,不在进行其他检查
    for _, validator := range field.Validators {
        if _, ok := validator.(Required); ok {
            if ok, message := validator.CleanData(field.GetValue()); !ok {
                field.Errors = append(field.Errors, message)
                return false
            }
        }
    }

    result := true

    for _, validator := range field.Validators {
        if ok, message := validator.CleanData(field.GetValue()); !ok {
            result = false
            field.Errors = append(field.Errors, message)
        }
    }

    return result
}

func (field *BaseField) GetValue() string {
    return field.Value
}

func (field *BaseField) SetValue(value string) {
    field.Value = value
}

func (field *BaseField) IsName(name string) bool {
    return field.Name == name
}

func (field *BaseField) RenderFull(attrs []string) template.HTML {
    return template.HTML("")
}

type TextField struct {
    BaseField
}

func (field *TextField) RenderInput(attrs []string) template.HTML {
    attrsStr := ""
    if len(attrs) > 0 {
        for _, attr := range attrs {
            attrsStr += " " + attr
        }
    }
    return template.HTML(fmt.Sprintf(`<input type="text" value="%s" name=%q id=%q%s>`, field.Value, field.Name, field.Name, attrsStr))
}

func (field *TextField) RenderFull(attrs []string) template.HTML {
    return renderFull(field, attrs)
}

func NewTextField(name string, label string, value string, validators ...IValidator) *TextField {
    field := TextField{}
    field.Name = name
    field.Label = label
    field.Value = value
    field.Validators = validators

    return &field
}

type PasswordField struct {
    BaseField
}

func (field *PasswordField) RenderInput(attrs []string) template.HTML {
    attrsStr := ""
    if len(attrs) > 0 {
        for _, attr := range attrs {
            attrsStr += " " + attr
        }
    }
    return template.HTML(fmt.Sprintf(`<input type="password" name=%q id=%q%s>`, field.Name, field.Name, attrsStr))
}

func NewPasswordField(name string, label string, validators ...IValidator) *PasswordField {
    field := PasswordField{}
    field.Name = name
    field.Label = label
    field.Validators = validators

    return &field
}

func (field *PasswordField) RenderFull(attrs []string) template.HTML {
    return renderFull(field, attrs)
}

type TextArea struct {
    BaseField
}

func (field *TextArea) RenderInput(attrs []string) template.HTML {
    attrsStr := ""
    if len(attrs) > 0 {
        for _, attr := range attrs {
            attrsStr += " " + attr
        }
    }

    return template.HTML(fmt.Sprintf(`<textarea id=%q name=%q%s>%s</textarea>`, field.Name, field.Name, attrsStr, field.Value))
}

func (field *TextArea) RenderFull(attrs []string) template.HTML {
    return renderFull(field, attrs)
}

func NewTextArea(name string, label string, value string, validators ...IValidator) *TextArea {
    field := TextArea{}
    field.Name = name
    field.Label = label
    field.Value = value
    field.Validators = validators

    return &field
}

type Choice struct {
    Value string
    Label string
}

type SelectField struct {
    BaseField
    Choices []Choice
}

func (field *SelectField) RenderInput(attrs []string) template.HTML {
    attrsStr := ""
    if len(attrs) > 0 {
        for _, attr := range attrs {
            attrsStr += " " + attr
        }
    }
    options := ""
    for _, choice := range field.Choices {
        selected := ""
        if choice.Value == field.Value {
            selected = " selected"
        }
        options += fmt.Sprintf(`<option value=%q%s>%s</option>`, choice.Value, selected, choice.Label)
    }

    return template.HTML(fmt.Sprintf(`<select id=%q name=%q%s>%s</select>`, field.Name, field.Name, attrsStr, options))
}

func (field *SelectField) RenderFull(attrs []string) template.HTML {
    return renderFull(field, attrs)
}

func NewSelectField(name string, label string, choices []Choice, defaultValue string, validators ...IValidator) *SelectField {
    field := SelectField{}
    field.Name = name
    field.Label = label
    field.Value = defaultValue
    field.Choices = choices

    return &field
}

type HiddenField struct {
    BaseField
}

func (field *HiddenField) RenderInput(attrs []string) template.HTML {
    return template.HTML(fmt.Sprintf(`<input type="hidden" value=%q name=%q id=%q>`, field.Value, field.Name, field.Name))
}

func (field *HiddenField) RenderFull(attrs []string) template.HTML {
    return field.RenderInput(attrs)
}

func NewHiddenField(name string, value string) *HiddenField {
    field := HiddenField{}
    field.Name = name
    field.Value = value

    return &field
}

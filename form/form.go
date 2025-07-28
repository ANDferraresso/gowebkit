package form

import (
	"github.com/ANDferraresso/gowebkit/orm"
	"github.com/ANDferraresso/gowebkit/validator"
)

func (form *Form) SetupForm() {
	form.Name = ""
	form.Prefix = ""
	form.Required = []string{}
	form.DontValidate = []string{}
	form.FieldsOrder = []string{}
	form.Fields = make(map[string]*Field)
	form.UIs = make(map[string]*UI)
	form.Validator = validator.Validator{}
	form.Validator.SetupValidator()
}

func (form *Form) AddField(fd *Field) {
	form.FieldsOrder = append(form.FieldsOrder, fd.Name)
	form.Fields[fd.Name] = fd
	form.UIs[fd.Name] = &UI{
		Attrs:      make(map[string]string),
		Default:    "",
		Widget:     "",
		WsUrl:      "",
		WsCallback: "",
		Opts:       make([]map[string]string, 0),
	}
}

func (form *Form) AddCsrfField(value string) {
	form.FieldsOrder = append(form.FieldsOrder, "_csrf")
	form.Fields["_csrf"] = &Field{
		Name:      "_csrf",
		Title:     "",
		MinLength: "",
		MaxLength: "",
		Checks:    []orm.Check{},
	}
	form.UIs["_csrf"] = &UI{
		Attrs:      make(map[string]string),
		Default:    "",
		Widget:     "input-hidden",
		WsUrl:      "",
		WsCallback: "",
		Opts:       make([]map[string]string, 0),
	}

	//form.UIs["_csrf"] = &UI{Attrs: map[string]string{"value": value}}
	form.UIs["_csrf"].Attrs["value"] = value
}

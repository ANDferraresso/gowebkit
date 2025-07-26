package form

import (
	"github.com/ANDferraresso/gowebkit/validator"
)

type FieldImg struct {
	Url string
	Alt string
	W   string
	H   string
}

type Field struct {
	Name      string
	Title     string
	Img       FieldImg
	MinLength string
	MaxLength string
	Checks    []validator.Check
}

type UI struct {
	Attrs      map[string]string
	Default    string
	Widget     string
	WsUrl      string
	WsCallback string
	Opts       []map[string]string
}

type Form struct {
	Name         string
	Prefix       string
	Required     []string
	DontValidate []string
	FieldsOrder  []string
	Fields       map[string]*Field
	UIs          map[string]*UI
	Validator    validator.Validator
}

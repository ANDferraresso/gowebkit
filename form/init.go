package form

import (
	"github.com/ANDferraresso/gowebkit/orm"
	"github.com/ANDferraresso/gowebkit/validator"
)

type FieldImg struct {
	Url string `json:"url"`
	Alt string `json:"alt"`
	W   string `json:"w"`
	H   string `json:"h"`
}

type Field struct {
	Name      string   `json:"name"`
	Title     string   `json:"title"`
	Img       FieldImg `json:"field_img"`
	MinLength string   `json:"min_length"`
	MaxLength string   `json:"max_length"`
	// Checks []orm.Check `json:"-"` // `json:"checks"`
	Checks []orm.Check `json:"checks"`
}

type UI struct {
	Attrs      map[string]string   `json:"attrs"`
	Default    string              `json:"default"`
	Widget     string              `json:"widget"`
	WsUrl      string              `json:"ws_url"`
	WsCallback string              `json:"ws_callback"`
	Opts       []map[string]string `json:"opts"`
}

type Form struct {
	Name         string              `json:"name"`
	Prefix       string              `json:"prefix"`
	Required     []string            `json:"required"`
	DontValidate []string            `json:"dont_validate"`
	FieldsOrder  []string            `json:"fields_order"`
	Fields       map[string]*Field   `json:"fields"`
	UIs          map[string]*UI      `json:"uis"`
	Validator    validator.Validator `json:"-"`
}

package form

import (
	"reflect"
	"slices"
	"strconv"

	"github.com/ANDferraresso/gowebkit/orm"
)

func (form *Form) ImportTable(table *orm.Table, tableT orm.Dictio, extRefs bool,
	fields []string, notRequired []string, prefix string) {
	//
	form.Name = table.Name
	form.Prefix = prefix
	for _, v := range fields {
		// Check if columns exists.
		if _, ok := table.Columns[v]; !ok {
			continue
		}
		form.FieldsOrder = append(form.FieldsOrder, v)
		// Isn't tabletT empty?
		if !reflect.DeepEqual(tableT, orm.Dictio{}) {
			form.Fields[v] = &Field{
				Name:      v,
				Title:     tableT.Title[v],
				MinLength: table.Columns[v].MinLength,
				MaxLength: table.Columns[v].MaxLength,
				Checks:    table.Columns[v].Checks,
			}
			form.UIs[v] = &UI{
				Attrs:      make(map[string]string),
				Default:    "",
				Widget:     "",
				WsUrl:      "",
				WsCallback: "",
				Opts:       tableT.Opts[v],
			}
		} else {
			// tabletT is empty.
			form.Fields[v] = &Field{
				Name:      v,
				Title:     "",
				MinLength: table.Columns[v].MinLength,
				MaxLength: table.Columns[v].MaxLength,
				Checks:    table.Columns[v].Checks,
			}
			form.UIs[v] = &UI{
				Attrs:      make(map[string]string),
				Default:    "",
				Widget:     "",
				WsUrl:      "",
				WsCallback: "",
				Opts:       make([]map[string]string, 0),
			}
		}

		// Default options.
		/*
		   "0": "Not defined",
		   "1": "NULL",
		   "2": "Empty string",
		   "3": "Come definito",
		   "4": "CURRENT_TIMESTAMP"
		*/
		if table.Columns[v].UcDefault == "3" {
			form.UIs[v].Default = table.Columns[v].Default
		}

		/* TO DO
		   if c_key in table_T['opts']:
		       self.form_def['ui'][c_key]['options'] = table_T['opts'][c_key]

		   if column['ui_widget'] == "input-checkbox":
		       if c_key in not_required:
		           pass
		       else:
		           self.form_def['required'].append(c_key)
		           self.form_def['ui'][c_key]['attrs']['required'] = "required"
		   else:
		       if c_key in not_required:
		           pass
		       else:
		           self.form_def['required'].append(c_key)
		           self.form_def['ui'][c_key]['attrs']['required'] = "required"
		*/

		if !slices.Contains(notRequired, v) {
			form.Required = append(form.Required, v)
			mL, err := strconv.Atoi(table.Columns[v].MinLength)
			if err == nil && mL > 0 {
				form.UIs[v].Attrs["required"] = "required"
			}
		}

		if table.Columns[v].UI_Widget != "" {
			form.UIs[v].Widget = table.Columns[v].UI_Widget
		}
		if table.Columns[v].UI_WsUrl != "" {
			form.UIs[v].WsUrl = table.Columns[v].UI_WsUrl
		}
		if table.Columns[v].UI_WsCallback != "" {
			form.UIs[v].WsCallback = table.Columns[v].UI_WsCallback
		}
	}
}

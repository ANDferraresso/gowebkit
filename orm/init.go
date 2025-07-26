package orm

import (
	"fmt"

	"github.com/ANDferraresso/gowebkit/validator"
)

type Dictio struct {
	EntTitle string                         `json:"EntTitle"`
	Title    map[string]string              `json:"Title"`
	Msg      map[string]string              `json:"Msg"`
	Info     map[string]string              `json:"Info"`
	Opts     map[string][]map[string]string `json:"Opts"`
}

type Res struct {
	Err  bool
	Msg  string
	Data []map[string]interface{}
}

type OptionsRes struct {
	Err  bool
	Msg  string
	Data []map[string]string
}

func ManageErr(res *Res, debug string, err error, query string) *Res {
	res.Err = true
	res.Msg = "DBMS ERROR"
	res.Data = []map[string]interface{}{}

	switch debug {
	case "0":
		res.Msg += "."
	case "1", "2":
		res.Msg += ": " + err.Error() + " - " + query
	default:

	}
	fmt.Println(err.Error() + " - " + query)
	return res
}

type Column struct {
	Type          string
	Length        string
	NotNull       bool
	UcDefault     string
	Default       string
	MinLength     string
	MaxLength     string
	Checks        []validator.Check
	UI_Widget     string
	UI_WsUrl      string
	UI_WsCallback string
}

type Table struct {
	Name             string
	Primary          []string
	Uniques          [][]string
	ColumnsInUniques []string
	FKeys            map[string]FKeys
	Indexes          [][]string
	ColumnsOrder     []string
	Columns          map[string]Column
}

type FKeys struct {
	ToTable  string
	ToColumn string
	ToRefs   []string
}

/*
func ManageNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func ManageDbResValue(rb sql.RawBytes) interface{} {
	if rb == nil {
		return nil
	} else {
		return string(rb)
	}
}

func ManageValues(values *[]interface{}, v interface{}) {
	if (reflect.TypeOf(v)) == nil {
		*values = append(*values, ManageNullString(""))
	} else {
		t := v.(interface{})
		switch t.(type) {
		case nil:
			*values = append(*values, ManageNullString(""))
		case bool:
			if v.(bool) == true {
				*values = append(*values, 1)
			} else {
				*values = append(*values, 0)
			}
		case int:
			*values = append(*values, v.(int))
		case int32:
			*values = append(*values, v.(int32))
		case int64:
			*values = append(*values, v.(int64))
		case string:
			*values = append(*values, v.(string))
		default:
			*values = append(*values, "")
		}
	}
}
*/

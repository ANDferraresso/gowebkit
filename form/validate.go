package form

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (form *Form) ValidateField(k string, value string) bool {
	if slices.Contains(form.DontValidate, k) {
		return true
	}

	if form.Fields[k].MinLength != "" {
		length, err := strconv.Atoi(form.Fields[k].MinLength)
		if err != nil {
			return false
		}
		if utf8.RuneCountInString(value) < length {
			return false
		}
	}

	if form.Fields[k].MaxLength != "" {
		length, err := strconv.Atoi(form.Fields[k].MaxLength)
		if err != nil {
			return false
		}
		if utf8.RuneCountInString(value) > length {
			return false
		}
	}

	// Checks
	for _, check := range form.Fields[k].Checks {
		if check.Func != "" {
			if !(form.Validator.Validate(check.Func, value, &check.Pars)) {
				return false
			}
		}
	}

	return true
}

/*
r.PostForm[...] returns a variable of type url.Values.
The url.Values type is defined in the net/url package and is a map
that associates strings with string slices.
I only take the first value [0], I don't process any others.
*/
func (form *Form) ValidateAll(r *http.Request) (bool, map[string]interface{}, []string) {
	err := r.ParseForm()
	if err != nil {
		return false, nil, nil
	}

	fValues := map[string]interface{}{}
	wrongFields := []string{}

	for _, k := range form.FieldsOrder {
		if k == "_csrf" {
			continue
		} else if slices.Contains(form.DontValidate, k) {
			continue
		} else {
			_, ok := r.PostForm[form.Prefix+k]
			if !ok {
				// Parameter not present.
				fValues[k] = ""
				if slices.Contains(form.Required, k) {
					wrongFields = append(wrongFields, k)
				}
			} else {
				value := ValToStr(r.PostForm[form.Prefix+k][0])
				value = strings.Trim(value, " ")
				fValues[k] = value
				// If the minimum allowed length is "0" (or "") and the input has length 0, do not validate it.
				if (form.Fields[k].MinLength == "" || form.Fields[k].MinLength == "0") && utf8.RuneCountInString(value) == 0 {
					//
				} else {
					if !form.ValidateField(k, value) {
						wrongFields = append(wrongFields, k)
					}
				}
			}
		}
	}

	if len(wrongFields) > 0 {
		return true, fValues, wrongFields
	}

	return true, fValues, wrongFields
}

func (form *Form) ValidateAllMap(param map[string]interface{}) (bool, map[string]interface{}, []string) {
	fValues := map[string]interface{}{}
	wrongFields := []string{}

	for _, k := range form.FieldsOrder {
		if k == "_csrf" {
			continue
		} else {
			_, ok := param[form.Prefix+k]
			if !ok {
				// Parameter not present.
				fValues[k] = ""
				if slices.Contains(form.Required, k) {
					wrongFields = append(wrongFields, k)
				}
			} else {
				value := ValToStr(param[form.Prefix+k])
				value = strings.Trim(value, " ")
				fValues[k] = value
				// If the minimum allowed length is "0" (or "") and the input has length 0, do not validate it.
				if (form.Fields[k].MinLength == "" || form.Fields[k].MinLength == "0") && utf8.RuneCountInString(value) == 0 {
					//
				} else {
					if !form.ValidateField(k, value) {
						wrongFields = append(wrongFields, k)
					}
				}
			}
		}
	}

	if len(wrongFields) > 0 {
		return false, fValues, wrongFields
	}

	return true, fValues, wrongFields
}

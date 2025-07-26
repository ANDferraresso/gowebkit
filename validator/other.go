package validator

import (
	"encoding/json"
	"slices"
)

func IsInSet(value string, pars *[]string) bool {
	return slices.Contains(*pars, value)
}

func IsJson(value string, pars *[]string) bool {
	var js interface{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		return false
	}

	switch js.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

package cast

import "strings"

// ValToBool returns true for the numeric value 1 (of any numeric type),
// the strings "1", "t" and "true" (case-insensitive; PostgreSQL returns
// booleans as "t"/"true") and the bool true; false otherwise.
func ValToBool(v interface{}) bool {
	switch v := v.(type) {
	case int:
		return v == 1
	case int8:
		return v == 1
	case int16:
		return v == 1
	case int32:
		return v == 1
	case int64:
		return v == 1
	case uint:
		return v == 1
	case uint8:
		return v == 1
	case uint16:
		return v == 1
	case uint32:
		return v == 1
	case uint64:
		return v == 1
	case float32:
		return v == 1
	case float64:
		return v == 1
	case string:
		return v == "1" || strings.EqualFold(v, "t") || strings.EqualFold(v, "true")
	case bool:
		return v
	default:
		return false
	}
}

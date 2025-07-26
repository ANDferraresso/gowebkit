package cast

func ValToBool(v interface{}) bool {
	switch v := v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if v == 1 {
			return true
		}
		return false
	case string:
		if v == "1" {
			return true
		}
		return false
	case bool:
		return v
	default:
		return false
	}
}

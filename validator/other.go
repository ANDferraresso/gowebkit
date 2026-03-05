package validator

import (
	"encoding/json"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Regex: "lat, lng". Es: -45.12, 12.3
var mapsLatLngRe = regexp.MustCompile(`^\s*([+-]?\d+(?:\.\d+)?)\s*,\s*([+-]?\d+(?:\.\d+)?)\s*$`)

func IsGoogleMapsLatLng(value string, pars *[]string) bool {
	s := strings.TrimSpace(value)

	m := mapsLatLngRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}

	lat, err := strconv.ParseFloat(m[1], 64)
	if err != nil || math.IsNaN(lat) || math.IsInf(lat, 0) {
		return false
	}
	lng, err := strconv.ParseFloat(m[2], 64)
	if err != nil || math.IsNaN(lng) || math.IsInf(lng, 0) {
		return false
	}

	if lat < -90 || lat > 90 {
		return false
	}
	if lng < -180 || lng > 180 {
		return false
	}

	return true
}

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

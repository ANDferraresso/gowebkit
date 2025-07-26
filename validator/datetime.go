package validator

import (
	"regexp"
	"slices"
	"strconv"
)

// (*pars)[0] = mode (it|eng|iso)
func IsDate(value string, pars *[]string) bool {
	return dateValidator(value, (*pars)[0])
}

// (*pars)[0] = mode (it|eng|iso)
func IsDateTime(value string, pars *[]string) bool {
	date := value[0:10]
	time := value[11:19]
	if !dateValidator(date, (*pars)[0]) {
		return false
	}
	if !timeValidator(time) {
		return false
	}
	return true
}

func IsTime(value string, pars *[]string) bool {
	return timeValidator(value)
}

// pars = mode (it|eng|iso), from (date; iso mode), to (date; iso mode)
func IsDateInRange(value string, pars *[]string) bool {
	if !dateValidator(value, (*pars)[0]) {
		return false
	}
	v := convertDate(value, (*pars)[0], "iso")

	var date1 string
	var date2 string
	var dateV string

	date1 = (*pars)[1][0:4] + (*pars)[1][5:7] + (*pars)[1][8:10]
	date2 = (*pars)[2][0:4] + (*pars)[2][5:7] + (*pars)[2][8:10]
	dateV = v[0:4] + v[5:7] + v[8:10]

	if dateV >= date1 && dateV <= date2 {
		return true
	}
	return false
}

// pars = from (time; iso mode), to (time; iso mode)
func IsTimeInRange(value string, pars *[]string) bool {
	var err error

	if !timeValidator(value) {
		return false
	}

	var time1_h int
	var time1_m int
	var time1_s int
	var time2_h int
	var time2_m int
	var time2_s int
	var timeV_h int
	var timeV_m int
	var timeV_s int

	time1_h, err = strconv.Atoi((*pars)[0][0:2])
	if err != nil {
		return false
	}
	time1_m, err = strconv.Atoi((*pars)[0][3:5])
	if err != nil {
		return false
	}
	time1_s, err = strconv.Atoi((*pars)[0][6:8])
	if err != nil {
		return false
	}
	time1 := time1_h*3600 + time1_m*60 + time1_s

	time2_h, err = strconv.Atoi((*pars)[1][0:2])
	if err != nil {
		return false
	}
	time2_m, err = strconv.Atoi((*pars)[1][3:5])
	if err != nil {
		return false
	}
	time2_s, err = strconv.Atoi((*pars)[1][6:8])
	if err != nil {
		return false
	}
	time2 := time2_h*3600 + time2_m*60 + time2_s

	timeV_h, err = strconv.Atoi(value[0:2])
	if err != nil {
		return false
	}
	timeV_m, err = strconv.Atoi(value[3:5])
	if err != nil {
		return false
	}
	timeV_s, err = strconv.Atoi(value[6:8])
	if err != nil {
		return false
	}
	timeV := timeV_h*3600 + timeV_m*60 + timeV_s

	if timeV >= time1 && timeV <= time2 {
		return true
	}
	return false
}

func convertDate(value string, fromMode string, toMode string) string {
	var d string
	var m string
	var y string

	// mode = it|en|iso
	switch fromMode {
	case "it":
		d = value[0:2]
		m = value[3:5]
		y = value[6:10]
	case "en":
		d = value[3:5]
		m = value[0:2]
		y = value[6:10]
	case "iso":
		d = value[8:10]
		m = value[5:7]
		y = value[0:4]
	default:
		return value // Nessuna conversione.
	}

	switch toMode {
	case "it":
		return d + "-" + m + "-" + y
	case "en":
		return m + "-" + d + "-" + y
	case "iso":
		return y + "-" + m + "-" + d
	default:
		return value // Nessuna conversione.
	}
}

func dateValidator(value string, mode string) bool {
	var err error

	if len(value) != 10 {
		return false
	}

	var d int
	var m int
	var y int

	switch mode {
	case "it": // DD-MM-YYYY
		re := regexp.MustCompile(`^[0-9]{2}\-[0-9]{2}\-[0-9]{4}$`)
		if !re.Match([]byte(value)) {
			return false
		}
		d, err = strconv.Atoi(value[0:2])
		if err != nil {
			return false
		}
		m, err = strconv.Atoi(value[3:5])
		if err != nil {
			return false
		}
		y, err = strconv.Atoi(value[6:10])
		if err != nil {
			return false
		}
	case "en": // MM-DD-YYYY
		re := regexp.MustCompile(`^[0-9]{2}\-[0-9]{2}\-[0-9]{4}$`)
		if !re.Match([]byte(value)) {
			return false
		}
		d, err = strconv.Atoi(value[3:5])
		if err != nil {
			return false
		}
		m, err = strconv.Atoi(value[0:2])
		if err != nil {
			return false
		}
		y, err = strconv.Atoi(value[6:10])
		if err != nil {
			return false
		}
	case "iso": // YYYY-MM-DD
		re := regexp.MustCompile(`^[0-9]{4}\-[0-9]{2}\-[0-9]{2}$`)
		if !re.Match([]byte(value)) {
			return false
		}
		d, err = strconv.Atoi(value[8:10])
		if err != nil {
			return false
		}
		m, err = strconv.Atoi(value[5:7])
		if err != nil {
			return false
		}
		y, err = strconv.Atoi(value[0:4])
		if err != nil {
			return false
		}
	default:
		return false
	}

	if d < 1 || d > 31 {
		return false
	}
	if m < 1 || m > 12 {
		return false
	}
	if slices.Contains([]int{4, 6, 9, 11}, m) && d > 30 {
		return false
	}
	if m == 2 {
		if d > 29 {
			return false
		}
		if d == 29 && (y%4 != 0 || (y%100 == 0 && y%400 != 0)) {
			return false
		}
	}

	return true
}

// Formato (tempo): hh:mm:ss
func timeValidator(value string) bool {
	var err error

	re := regexp.MustCompile(`^[0-9]{2}:[0-9]{2}:[0-9]{2}$`)
	if !re.Match([]byte(value)) {
		return false
	}

	var h int
	var m int
	var s int

	h, err = strconv.Atoi(value[0:2])
	if err != nil {
		return false
	}
	m, err = strconv.Atoi(value[3:5])
	if err != nil {
		return false
	}
	s, err = strconv.Atoi(value[6:8])
	if err != nil {
		return false
	}

	if h < 0 || h > 23 {
		return false
	}
	if m < 0 || m > 59 {
		return false
	}
	if s < 0 || s > 59 {
		return false
	}
	return true
}

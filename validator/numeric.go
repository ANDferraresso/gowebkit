package validator

import (
	"regexp"
	"strconv"
)

// NUMERIC

func IsDecimal(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^-?(0\.[0-9]+|[1-9][0-9]*\.[0-9]+)$`) // -1.0, 10.52, ...
	return re.Match([]byte(value))
}

func IsMoney(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^(0\.[0-9]{2}|[1-9][0-9]*\.[0-9]{2})$`) // 0.00, 1.20, 19.50, ...
	return re.Match([]byte(value))
}

func IsInteger(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`) // 0, 1, -2, -123, 456, ...
	return re.Match([]byte(value))
}

func IsNegativeInt(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^-[1-9][0-9]*$`) // -1, -10, -123, ...
	return re.Match([]byte(value))
}

func IsPositiveInt(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^[1-9][0-9]*$`) // 1, 10, 123, ...
	return re.Match([]byte(value))
}

func IsZeroNegativeInt(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^(0|-[1-9][0-9]*)$`) // 0, -1, -10, -123, ...
	return re.Match([]byte(value))
}

func IsZeroPositiveInt(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^(0|[1-9][0-9]*)$`) // 0, 1, 10, 123, ...
	return re.Match([]byte(value))
}

// COMPARING

func IsIntGreater(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p, err := strconv.ParseInt((*pars)[0], 10, 64)
	if err != nil {
		return false
	}
	if x > p {
		return true
	}
	return false
}

func IsIntGreaterEqual(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p, err := strconv.ParseInt((*pars)[0], 10, 64)
	if err != nil {
		return false
	}
	if x >= p {
		return true
	}
	return false
}

func IsIntInRange(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p1, err := strconv.ParseInt((*pars)[0], 10, 64)
	if err != nil {
		return false
	}
	p2, err := strconv.ParseInt((*pars)[1], 10, 64)
	if err != nil {
		return false
	}
	if x >= p1 && x <= p2 {
		return true
	}
	return false
}

func IsIntLower(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p, err := strconv.ParseInt((*pars)[0], 10, 64)
	if err != nil {
		return false
	}
	if x < p {
		return true
	}
	return false
}

func IsIntLowerEqual(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p, err := strconv.ParseInt((*pars)[0], 10, 64)
	if err != nil {
		return false
	}
	if x <= p {
		return true
	}
	return false
}

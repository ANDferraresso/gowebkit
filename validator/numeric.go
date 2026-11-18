package validator

import (
	"regexp"
	"strconv"
)

// NUMERIC

var (
	decimalRe         = regexp.MustCompile(`^-?(0\.[0-9]+|[1-9][0-9]*\.[0-9]+)$`)   // -1.0, 10.52, ...
	moneyRe           = regexp.MustCompile(`^(0\.[0-9]{2}|[1-9][0-9]*\.[0-9]{2})$`) // 0.00, 1.20, 19.50, ...
	integerRe         = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)                   // 0, 1, -2, -123, 456, ...
	negativeIntRe     = regexp.MustCompile(`^-[1-9][0-9]*$`)                        // -1, -10, -123, ...
	positiveIntRe     = regexp.MustCompile(`^[1-9][0-9]*$`)                         // 1, 10, 123, ...
	zeroNegativeIntRe = regexp.MustCompile(`^(0|-[1-9][0-9]*)$`)                    // 0, -1, -10, -123, ...
	zeroPositiveIntRe = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)                     // 0, 1, 10, 123, ...
)

func IsDecimal(value string, pars *[]string) bool {
	return decimalRe.MatchString(value)
}

func IsMoney(value string, pars *[]string) bool {
	return moneyRe.MatchString(value)
}

func IsInteger(value string, pars *[]string) bool {
	return integerRe.MatchString(value)
}

func IsNegativeInt(value string, pars *[]string) bool {
	return negativeIntRe.MatchString(value)
}

func IsPositiveInt(value string, pars *[]string) bool {
	return positiveIntRe.MatchString(value)
}

func IsZeroNegativeInt(value string, pars *[]string) bool {
	return zeroNegativeIntRe.MatchString(value)
}

func IsZeroPositiveInt(value string, pars *[]string) bool {
	return zeroPositiveIntRe.MatchString(value)
}

// COMPARING

// intPar parses the i-th parameter as int64.
func intPar(pars *[]string, i int) (int64, bool) {
	s, ok := par(pars, i)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// compareInt parses value and the first parameter, then applies cmp.
func compareInt(value string, pars *[]string, cmp func(x, p int64) bool) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p, ok := intPar(pars, 0)
	if !ok {
		return false
	}
	return cmp(x, p)
}

func IsIntGreater(value string, pars *[]string) bool {
	return compareInt(value, pars, func(x, p int64) bool { return x > p })
}

func IsIntGreaterEqual(value string, pars *[]string) bool {
	return compareInt(value, pars, func(x, p int64) bool { return x >= p })
}

func IsIntInRange(value string, pars *[]string) bool {
	x, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	p1, ok1 := intPar(pars, 0)
	p2, ok2 := intPar(pars, 1)
	if !ok1 || !ok2 {
		return false
	}
	return x >= p1 && x <= p2
}

func IsIntLower(value string, pars *[]string) bool {
	return compareInt(value, pars, func(x, p int64) bool { return x < p })
}

func IsIntLowerEqual(value string, pars *[]string) bool {
	return compareInt(value, pars, func(x, p int64) bool { return x <= p })
}

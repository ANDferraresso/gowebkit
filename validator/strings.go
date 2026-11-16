package validator

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// passwordSpecialChars is the set of characters that count as "special"
// for the password levels B and D.
const passwordSpecialChars = `@$!%*#?&`

// IsPassword checks the password strength; (*pars)[0] is the level:
//
//	A: minimum eight characters, at least one letter and one number.
//	B: minimum eight characters, at least one letter, one number and one special character.
//	C: minimum eight characters, at least one uppercase letter, one lowercase letter and one number.
//	D: minimum eight characters, at least one uppercase letter, one lowercase letter, one number and one special character.
//
// The whole value is checked (not just a substring); characters other than
// the required ones are allowed.
func IsPassword(value string, pars *[]string) bool {
	level, ok := par(pars, 0)
	if !ok {
		return false
	}
	if utf8.RuneCountInString(value) < 8 {
		return false
	}

	var hasLetter, hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasLetter, hasUpper = true, true
		case unicode.IsLower(r):
			hasLetter, hasLower = true, true
		case unicode.IsLetter(r):
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
		if strings.ContainsRune(passwordSpecialChars, r) {
			hasSpecial = true
		}
	}

	switch level {
	case "A":
		return hasLetter && hasDigit
	case "B":
		return hasLetter && hasDigit && hasSpecial
	case "C":
		return hasUpper && hasLower && hasDigit
	case "D":
		return hasUpper && hasLower && hasDigit && hasSpecial
	}
	return false
}

func AllowedChars(value string, pars *[]string) bool {
	if utf8.RuneCountInString(value) == 0 {
		return true
	}
	chars, ok := par(pars, 0)
	if !ok {
		return false
	}
	re, err := regexp.Compile("^[" + chars + "]*$")
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

func ForbiddenChars(value string, pars *[]string) bool {
	if utf8.RuneCountInString(value) == 0 {
		return true
	}
	chars, ok := par(pars, 0)
	if !ok {
		return false
	}
	re, err := regexp.Compile("^[^" + chars + "]*$")
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

func IsStringEqual(value string, pars *[]string) bool {
	s, ok := par(pars, 0)
	return ok && value == s
}

// lengthPar parses the i-th parameter as int.
func lengthPar(pars *[]string, i int) (int, bool) {
	s, ok := par(pars, i)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

func IsLength(value string, pars *[]string) bool {
	length, ok := lengthPar(pars, 0)
	return ok && utf8.RuneCountInString(value) == length
}

func IsLengthInRange(value string, pars *[]string) bool {
	lengthA, okA := lengthPar(pars, 0)
	lengthB, okB := lengthPar(pars, 1)
	if !okA || !okB {
		return false
	}
	n := utf8.RuneCountInString(value)
	return n >= lengthA && n <= lengthB
}

func IsMaxLength(value string, pars *[]string) bool {
	length, ok := lengthPar(pars, 0)
	return ok && utf8.RuneCountInString(value) <= length
}

func IsMinLength(value string, pars *[]string) bool {
	length, ok := lengthPar(pars, 0)
	return ok && utf8.RuneCountInString(value) >= length
}

// IsRegex matches value against the pattern in (*pars)[0].
// An invalid pattern makes the validation fail instead of panicking.
func IsRegex(value string, pars *[]string) bool {
	pattern, ok := par(pars, 0)
	if !ok {
		return false
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

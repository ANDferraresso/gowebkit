package validator

import (
	"regexp"
	"strconv"
	"unicode/utf8"
)

func IsPassword(value string, pars *[]string) bool {
	switch (*pars)[0] {
	case "A":
		// Minimum eight characters, at least one letter and one number.
		re := regexp.MustCompile(`[A-Za-z0-9]{8,}`)
		if !re.Match([]byte(value)) {
			return false
		} else {
			re1 := regexp.MustCompile(`[A-Za-z]+`)
			re2 := regexp.MustCompile(`[0-9]+`)
			if re1.Match([]byte(value)) && re2.Match([]byte(value)) {
				return true
			} else {
				return false
			}
		}
	case "B":
		// Minimum eight characters, at least one letter, one number and one special character.
		re := regexp.MustCompile(`[A-Za-z0-9@\$!%\*#\?&]{8,}`)
		if !re.Match([]byte(value)) {
			return false
		} else {
			re1 := regexp.MustCompile(`[A-Za-z]+`)
			re2 := regexp.MustCompile(`[0-9]+`)
			re3 := regexp.MustCompile(`[@\$!%\*#\?&]+`)
			if re1.Match([]byte(value)) && re2.Match([]byte(value)) && re3.Match([]byte(value)) {
				return true
			} else {
				return false
			}
		}
	case "C":
		// Minimum eight characters, at least one uppercase letter, one lowercase letter and one number.
		re := regexp.MustCompile(`[A-Za-z0-9]{8,}`)
		if !re.Match([]byte(value)) {
			return false
		} else {
			re1 := regexp.MustCompile(`[A-Z]+`)
			re2 := regexp.MustCompile(`[a-z]+`)
			re3 := regexp.MustCompile(`[0-9]+`)
			if re1.Match([]byte(value)) && re2.Match([]byte(value)) && re3.Match([]byte(value)) {
				return true
			} else {
				return false
			}
		}
	case "D":
		// Minimum eight characters, at least one uppercase letter, one lowercase letter, one number and one special character.
		re := regexp.MustCompile(`[A-Za-z0-9@\$!%\*#\?&]{8,}`)
		if !re.Match([]byte(value)) {
			return false
		} else {
			re1 := regexp.MustCompile(`[A-Z]+`)
			re2 := regexp.MustCompile(`[a-z]+`)
			re3 := regexp.MustCompile(`[0-9]+`)
			re4 := regexp.MustCompile(`[@\$!%\*#\?&]+`)
			if re1.Match([]byte(value)) && re2.Match([]byte(value)) && re3.Match([]byte(value)) && re4.Match([]byte(value)) {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func AllowedChars(value string, pars *[]string) bool {
	if utf8.RuneCountInString(value) == 0 {
		return true
	}
	re := regexp.MustCompile("^[" + (*pars)[0] + "]*$")
	return re.Match([]byte(value))
}

func ForbiddenChars(value string, pars *[]string) bool {
	if utf8.RuneCountInString(value) == 0 {
		return true
	}
	re := regexp.MustCompile("^[^" + (*pars)[0] + "]*$")
	return re.Match([]byte(value))
}

func IsStringEqual(value string, pars *[]string) bool {
	return value == (*pars)[0]
}

func IsLength(value string, pars *[]string) bool {
	length, err := strconv.Atoi((*pars)[0])
	if err != nil {
		return false
	}
	if utf8.RuneCountInString(value) == length {
		return true
	}
	return false
}

func IsLengthInRange(value string, pars *[]string) bool {
	lengthA, err := strconv.Atoi((*pars)[0])
	if err != nil {
		return false
	}
	lengthB, err := strconv.Atoi((*pars)[1])
	if err != nil {
		return false
	}
	if utf8.RuneCountInString(value) >= lengthA && utf8.RuneCountInString(value) <= lengthB {
		return true
	}
	return false
}

func IsMaxLength(value string, pars *[]string) bool {
	length, err := strconv.Atoi((*pars)[0])
	if err != nil {
		return false
	}
	if utf8.RuneCountInString(value) <= length {
		return true
	}
	return false
}

func IsMinLength(value string, pars *[]string) bool {
	length, err := strconv.Atoi((*pars)[0])
	if err != nil {
		return false
	}
	if utf8.RuneCountInString(value) >= length {
		return true
	}
	return false
}

func IsRegex(value string, pars *[]string) bool {
	re := regexp.MustCompile((*pars)[0])
	return re.Match([]byte(value))
}

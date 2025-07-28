package validator

type Check struct {
	Func string   `json:"func"`
	Pars []string `json:"pars"`
}

type Validator struct {
	Funcs map[string]func(value string, pars *[]string) bool
}

func (val *Validator) SetupValidator() {
	val.Funcs = map[string]func(value string, pars *[]string) bool{
		// Date & Time
		"isDate":        IsDate,
		"isDateTime":    IsDateTime,
		"isTime":        IsTime,
		"isDateInRange": IsDateInRange,
		"isTimeInRange": IsTimeInRange,
		// Internet
		"isDomain": IsDomain,
		"isEmail":  IsEmail,
		"isIPV4":   IsIPV4,
		"isURL":    IsURL,
		// Numeric
		"isDecimal":         IsDecimal,
		"isMoney":           IsMoney,
		"isInteger":         IsInteger,
		"isNegativeInt":     IsNegativeInt,
		"isPositiveInt":     IsPositiveInt,
		"isZeroNegativeInt": IsZeroNegativeInt,
		"isZeroPositiveInt": IsZeroPositiveInt,
		"isIntGreater":      IsIntGreater,
		"isIntGreaterEqual": IsIntGreaterEqual,
		"isIntInRange":      IsIntInRange,
		"isIntLower":        IsIntLower,
		"isIntLowerEqual":   IsIntLowerEqual,
		// String
		"isPassword":      IsPassword,
		"allowedChars":    AllowedChars,
		"forbiddenChars":  ForbiddenChars,
		"isStringEqual":   IsStringEqual,
		"isLength":        IsLength,
		"isLengthInRange": IsLengthInRange,
		"isMaxLength":     IsMaxLength,
		"isMinLength":     IsMinLength,
		"isRegex":         IsRegex,
		// Other
		"isInSet": IsInSet,
		"isJson":  IsJson,
	}
}

func (val *Validator) Validate(f string, value string, pars *[]string) bool {
	fn, exists := val.Funcs[f]
	if !exists {
		return false
	}
	return fn(value, pars)
	// return val.Funcs[f](value, pars)
}

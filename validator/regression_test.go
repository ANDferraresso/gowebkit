package validator

import "testing"

func TestIsDateTime(t *testing.T) {
	tests := []struct {
		value, mode string
		want        bool
	}{
		{"2024-12-31 23:59:59", "iso", true},
		{"2024-12-31T23:59:59", "iso", true},
		{"31-12-2024 10:00:00", "it", true},
		{"2024", "iso", false},
		{"", "iso", false},
		{"2024-12-31 23:59", "iso", false},
		{"2024-12-31X23:59:59", "iso", false},
		{"2024-12-31 23:59:59<script>", "iso", false},
		{"2024-02-30 10:00:00", "iso", false},
	}
	for _, tt := range tests {
		pars := []string{tt.mode}
		if got := IsDateTime(tt.value, &pars); got != tt.want {
			t.Errorf("IsDateTime(%q, %s) = %v, want %v", tt.value, tt.mode, got, tt.want)
		}
	}
}

func TestIsPasswordWholeValue(t *testing.T) {
	tests := []struct {
		value, level string
		want         bool
	}{
		{"abcd1234", "A", true},
		{"abc 12 <>", "A", true}, // other characters are allowed
		{"abcdefgh", "A", false}, // no digit
		{"a1", "A", false},       // too short
		{"abc123!x", "B", true},
		{"abc123xx", "B", false}, // no special char
		{"Abcdef12", "C", true},
		{"abcdef12", "C", false}, // no uppercase
		{"Abcdef1!", "D", true},
		{"Abcdef12", "D", false}, // no special char
		{"Abcdef1!", "Z", false}, // unknown level
	}
	for _, tt := range tests {
		pars := []string{tt.level}
		if got := IsPassword(tt.value, &pars); got != tt.want {
			t.Errorf("IsPassword(%q, %s) = %v, want %v", tt.value, tt.level, got, tt.want)
		}
	}
}

// Missing parameters or invalid patterns must make validation fail, never panic.
func TestValidatorsDoNotPanic(t *testing.T) {
	var v Validator
	v.SetupValidator()

	inputs := []string{"", "x", "2024", "12:00:00", "2024-01-01", "abc123!X"}
	parSets := [][]string{nil, {}, {"("}, {"["}, {"bad", "bad", "bad"}}

	for name := range v.Funcs {
		for _, in := range inputs {
			for _, ps := range parSets {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("%s(%q, %v) panicked: %v", name, in, ps, r)
						}
					}()
					if ps == nil {
						v.Funcs[name](in, nil)
						return
					}
					p := append([]string(nil), ps...)
					v.Validate(name, in, &p)
				}()
			}
		}
	}
}

func TestRangeValidators(t *testing.T) {
	if !IsDateInRange("15-06-2024", &[]string{"it", "2024-01-01", "2024-12-31"}) {
		t.Error("IsDateInRange: expected true")
	}
	if IsDateInRange("15-06-2025", &[]string{"it", "2024-01-01", "2024-12-31"}) {
		t.Error("IsDateInRange: expected false")
	}
	if !IsTimeInRange("12:00:00", &[]string{"08:00:00", "18:00:00"}) {
		t.Error("IsTimeInRange: expected true")
	}
	if IsTimeInRange("19:00:00", &[]string{"08:00:00", "18:00:00"}) {
		t.Error("IsTimeInRange: expected false")
	}
	if !IsIntInRange("5", &[]string{"1", "10"}) || IsIntInRange("11", &[]string{"1", "10"}) {
		t.Error("IsIntInRange: wrong result")
	}
}

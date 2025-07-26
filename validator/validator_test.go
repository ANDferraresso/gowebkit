package validator

import "testing"

func TestIsDate(t *testing.T) {
	tests := []struct {
		value string
		mode  string
		want  bool
	}{
		{"31-12-2024", "it", true},
		{"12-31-2024", "en", true},
		{"2024-12-31", "iso", true},
		{"31-02-2024", "it", false},
		{"29-02-1900", "it", false},
		{"29-02-2000", "it", true},
	}

	for _, tt := range tests {
		pars := []string{tt.mode}
		if got := IsDate(tt.value, &pars); got != tt.want {
			t.Errorf("IsDate(%s, %s) = %v, want %v", tt.value, tt.mode, got, tt.want)
		}
	}
}

func TestIsPassword(t *testing.T) {
	tests := []struct {
		value string
		level string
		want  bool
	}{
		{"abcd1234", "A", true},
		{"abcdefg!", "A", false},
		{"Abc123!@", "B", true},
		{"abc123", "C", false},
		{"Abcdef1!", "D", true},
	}

	for _, tt := range tests {
		pars := []string{tt.level}
		if got := IsPassword(tt.value, &pars); got != tt.want {
			t.Errorf("IsPassword(%s, %s) = %v, want %v", tt.value, tt.level, got, tt.want)
		}
	}
}

func TestIsInteger(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"123", true},
		{"-123", true},
		{"0", true},
		{"01", false},
		{"--1", false},
	}

	for _, tt := range tests {
		pars := []string{}
		if got := IsInteger(tt.value, &pars); got != tt.want {
			t.Errorf("IsInteger(%s) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"user@example.com", true},
		{"user@.com", false},
		{"@domain.com", false},
		{"user@domain", false},
		{"user@domain.org", true},
	}

	for _, tt := range tests {
		pars := []string{}
		if got := IsEmail(tt.value, &pars); got != tt.want {
			t.Errorf("IsEmail(%s) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"http://example.com", true},
		{"https://example.com/path", true},
		{"example.com", true},
		{"ftp://example.com", false},
		{"http://", false},
	}

	for _, tt := range tests {
		pars := []string{}
		if got := IsURL(tt.value, &pars); got != tt.want {
			t.Errorf("IsURL(%s) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

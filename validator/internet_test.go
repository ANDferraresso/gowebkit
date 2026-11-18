package validator

import "testing"

func TestInternetValidators(t *testing.T) {
	tests := []struct {
		name  string
		fn    func(string, *[]string) bool
		value string
		want  bool
	}{
		{"IsDomain", IsDomain, "example.com", true},
		{"IsDomain", IsDomain, "xn--bcher-kva.example", true},
		{"IsDomain", IsDomain, "sub.my-site.it", true},
		{"IsDomain", IsDomain, "-bad.com", false},
		{"IsDomain", IsDomain, "localhost", false},
		{"IsEmail", IsEmail, "andrea.f+tag@example.co.uk", true},
		{"IsEmail", IsEmail, "no-at.example.com", false},
		{"IsEmail", IsEmail, "a@b", false},
		{"IsIPV4", IsIPV4, "192.168.1.1", true},
		{"IsIPV4", IsIPV4, "256.1.1.1", false},
		{"IsIPV4", IsIPV4, "::1", false},
		{"IsIPV4", IsIPV4, "1.2.3", false},
		{"IsURL", IsURL, "https://example.com/path?q=1", true},
		{"IsURL", IsURL, "example.org", true},
		{"IsURL", IsURL, "ftp://example.com", false},
		{"IsURL", IsURL, "https://exa mple.com", false},
	}
	for _, tt := range tests {
		if got := tt.fn(tt.value, nil); got != tt.want {
			t.Errorf("%s(%q) = %v, want %v", tt.name, tt.value, got, tt.want)
		}
	}
}

func TestTimeValidatorModes(t *testing.T) {
	tests := []struct {
		value, mode string
		want        bool
	}{
		{"23", "H", true},
		{"24", "H", false},
		{"23:59", "HM", true},
		{"23:60", "HM", false},
		{"00:00:00", "HMS", true},
		{"12:30:60", "HMS", false},
		{"1:30:00", "HMS", false},
	}
	for _, tt := range tests {
		pars := []string{tt.mode}
		if got := IsTime(tt.value, &pars); got != tt.want {
			t.Errorf("IsTime(%q, %s) = %v, want %v", tt.value, tt.mode, got, tt.want)
		}
	}
}

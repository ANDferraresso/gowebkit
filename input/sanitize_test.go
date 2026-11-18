package input

import "testing"

func TestNormalizeSingleLine(t *testing.T) {
	tests := map[string]string{
		"  Mario Rossi \t": "Mario Rossi",
		"\n a  b \r\n":     "a  b",
		"":                 "",
		"   ":              "",
		"O'Brien":          "O'Brien",
	}
	for in, want := range tests {
		if got := NormalizeSingleLine(in); got != want {
			t.Errorf("NormalizeSingleLine(%q) = %q, want %q", in, got, want)
		}
	}
}

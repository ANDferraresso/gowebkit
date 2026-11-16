package cast

import "testing"

func TestValToBool(t *testing.T) {
	tests := []struct {
		in   interface{}
		want bool
	}{
		{1, true},
		{int8(1), true},
		{int64(1), true},
		{uint32(1), true},
		{float32(1), true},
		{float64(1), true},
		{0, false},
		{int64(2), false},
		{float64(1.5), false},
		{"1", true},
		{"0", false},
		{"t", true},
		{"true", true},
		{"TRUE", true},
		{"f", false},
		{"false", false},
		{"yes", false},
		{true, true},
		{false, false},
		{nil, false},
	}
	for _, tt := range tests {
		if got := ValToBool(tt.in); got != tt.want {
			t.Errorf("ValToBool(%#v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

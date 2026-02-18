package core

import (
	"testing"
)

func TestCalcSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "json string",
			input:    `{"type":"transaction","data":{}}`,
			expected: "8c2a767d2d2139f1386dc17b5c256234623b1ffe87ff33c1e5dec904973a3b0b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcSHA256(tt.input)
			if result != tt.expected {
				t.Errorf("CalcSHA256(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalcSHA256_Deterministic(t *testing.T) {
	input := "test data"
	result1 := CalcSHA256(input)
	result2 := CalcSHA256(input)

	if result1 != result2 {
		t.Errorf("CalcSHA256 is not deterministic: %q != %q", result1, result2)
	}
}

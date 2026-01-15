package redirect

import (
	"testing"
)

func TestIsValidCode(t *testing.T) {
	s := &Service{codeLength: 8}

	tests := []struct {
		code string
		want bool
	}{
		// Valid codes
		{"abcdef", true},
		{"ABC123xyz", true},
		{"123456789012", true}, // max 12

		// Invalid: too short
		{"abc", false},
		{"abcde", false},

		// Invalid: too long
		{"abcdefghijklm", false},

		// Invalid: bad chars
		{"abc-def", false},
		{"abc_123", false},
		{"abc 123", false},
	}

	for _, tt := range tests {
		if got := s.isValidCode(tt.code); got != tt.want {
			t.Errorf("isValidCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

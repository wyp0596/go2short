package link

import (
	"testing"
)

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		// Valid codes
		{"abcdef", true},      // min length 6
		{"AbCdEf123456", true}, // max length 12
		{"abc123XYZ", true},

		// Invalid: too short
		{"abc", false},
		{"abcde", false},

		// Invalid: too long
		{"abcdefghijklm", false},

		// Invalid: bad chars
		{"abc-def", false},
		{"abc_def", false},
		{"abc def", false},
		{"abc!@#", false},
	}

	for _, tt := range tests {
		if got := isValidCode(tt.code); got != tt.want {
			t.Errorf("isValidCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestGenerateCode(t *testing.T) {
	lengths := []int{6, 8, 12}
	for _, length := range lengths {
		code := generateCode(length)
		if len(code) != length {
			t.Errorf("generateCode(%d) returned len %d", length, len(code))
		}
		if !isValidCode(code) {
			t.Errorf("generateCode(%d) returned invalid code: %s", length, code)
		}
	}
}

func TestIsPrivateHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		// Private IPv4
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"169.254.1.1", true},

		// Public IPv4
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"203.0.113.1", false},

		// Private IPv6
		{"::1", true},
		{"fc00::1", true},
		{"fe80::1", true},

		// Public IPv6
		{"2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		if got := isPrivateHost(tt.host); got != tt.want {
			t.Errorf("isPrivateHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

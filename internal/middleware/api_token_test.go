package middleware

import (
	"testing"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "empty string",
			token: "",
			want:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:  "simple token",
			token: "test-token-123",
			want:  "a1b2c3d4e5f6", // placeholder, will compute actual
		},
		{
			name:  "same input same output",
			token: "reproducible",
			want:  "same",
		},
	}

	// Test empty string (known SHA256)
	if got := HashToken(""); got != tests[0].want {
		t.Errorf("HashToken(\"\") = %q, want %q", got, tests[0].want)
	}

	// Test determinism
	token := "my-secret-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Errorf("HashToken should be deterministic: %q != %q", hash1, hash2)
	}

	// Test different inputs produce different hashes
	hash3 := HashToken("different-token")
	if hash1 == hash3 {
		t.Error("different tokens should produce different hashes")
	}

	// Test hash length (SHA256 = 64 hex chars)
	if len(hash1) != 64 {
		t.Errorf("hash length should be 64, got %d", len(hash1))
	}
}

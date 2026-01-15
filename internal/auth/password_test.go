package auth

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == password {
		t.Error("hash should not equal plaintext password")
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	if CheckPassword("wrongPassword", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "samePassword"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("same password should produce different hashes (bcrypt salt)")
	}

	// Both should still verify
	if !CheckPassword(password, hash1) || !CheckPassword(password, hash2) {
		t.Error("both hashes should verify correctly")
	}
}

package security_test

import (
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
)

func TestHashAndVerifyPassword_RoundTrip(t *testing.T) {
	hash, err := security.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}

	ok, err := security.VerifyPassword("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword() = false, want true for correct password")
	}
}

func TestVerifyPassword_WrongPasswordFails(t *testing.T) {
	hash, err := security.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}

	ok, err := security.VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() unexpected error: %v", err)
	}
	if ok {
		t.Fatal("VerifyPassword() = true, want false for wrong password")
	}
}

func TestVerifyPassword_MalformedHashReturnsError(t *testing.T) {
	_, err := security.VerifyPassword("anything", "not-a-valid-hash")
	if err == nil {
		t.Fatal("VerifyPassword() expected error for malformed hash, got nil")
	}
}

func TestHashPassword_ProducesDifferentSaltsEachCall(t *testing.T) {
	hash1, err := security.HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	hash2, err := security.HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Fatal("HashPassword() produced identical hashes for two calls — salt is not randomized")
	}
}

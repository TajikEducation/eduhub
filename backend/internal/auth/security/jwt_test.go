package security_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
)

func TestIssueAndParseAccessToken_RoundTrip(t *testing.T) {
	userID := uuid.New()
	token, err := security.IssueAccessToken("test-secret", userID, "user", time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken() unexpected error: %v", err)
	}

	claims, err := security.ParseAccessToken("test-secret", token)
	if err != nil {
		t.Fatalf("ParseAccessToken() unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("claims.UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Role != "user" {
		t.Errorf("claims.Role = %q, want %q", claims.Role, "user")
	}
}

func TestParseAccessToken_ExpiredTokenFails(t *testing.T) {
	token, err := security.IssueAccessToken("test-secret", uuid.New(), "user", -time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken() unexpected error: %v", err)
	}

	if _, err := security.ParseAccessToken("test-secret", token); err == nil {
		t.Fatal("ParseAccessToken() expected error for expired token, got nil")
	}
}

func TestParseAccessToken_WrongSecretFails(t *testing.T) {
	token, err := security.IssueAccessToken("secret-a", uuid.New(), "user", time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken() unexpected error: %v", err)
	}

	if _, err := security.ParseAccessToken("secret-b", token); err == nil {
		t.Fatal("ParseAccessToken() expected error for wrong secret, got nil")
	}
}

func TestNewRefreshToken_HashIsDeterministicFromRaw(t *testing.T) {
	raw, hash, err := security.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() unexpected error: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatal("NewRefreshToken() returned empty raw or hash")
	}
	if got := security.HashRefreshToken(raw); got != hash {
		t.Errorf("HashRefreshToken(raw) = %q, want %q (must match hash returned by NewRefreshToken)", got, hash)
	}
}

func TestNewRefreshToken_ProducesUniqueTokens(t *testing.T) {
	raw1, _, err := security.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() unexpected error: %v", err)
	}
	raw2, _, err := security.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() unexpected error: %v", err)
	}
	if raw1 == raw2 {
		t.Fatal("NewRefreshToken() produced identical tokens across two calls")
	}
}

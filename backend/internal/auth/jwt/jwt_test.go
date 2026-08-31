package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

func TestIssueParse_RoundTrip_RecoversUserIDAndRole(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC))
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	userID := uuid.New()

	token, err := issuer.Issue(userID, "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("Parse() вернул ошибку: %v", err)
	}

	gotID, err := claims.UserID()
	if err != nil {
		t.Fatalf("claims.UserID() вернул ошибку: %v", err)
	}
	if gotID != userID {
		t.Errorf("UserID() = %v, want %v", gotID, userID)
	}
	if claims.Role != "user" {
		t.Errorf("Role = %q, want %q", claims.Role, "user")
	}
}

func TestParse_ExpiredToken_ReturnsError(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC))
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)

	token, err := issuer.Issue(uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	clk.Advance(16 * time.Minute)

	_, err = issuer.Parse(token)
	if err == nil {
		t.Fatal("Parse() истёкшего токена вернул nil-ошибку")
	}
}

func TestParse_TamperedSignature_ReturnsError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)

	token, err := issuer.Issue(uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	// Токен, подписанный ДРУГИМ секретом, не должен проходить проверку исходным Issuer.
	otherIssuer := jwt.NewIssuer([]byte("другой-секрет"), 15*time.Minute, clk)
	otherToken, err := otherIssuer.Issue(uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() (другой issuer) вернул ошибку: %v", err)
	}

	if _, err := issuer.Parse(otherToken); err == nil {
		t.Error("Parse() токена с чужой подписью вернул nil-ошибку")
	}

	// Явная порча подписи исходного токена тоже должна быть отклонена.
	tampered := token[:len(token)-4] + "AAAA"
	if _, err := issuer.Parse(tampered); err == nil {
		t.Error("Parse() испорченного токена вернул nil-ошибку")
	}
}

func TestParse_MalformedToken_ReturnsErrorNotPanic(t *testing.T) {
	clk := clock.NewFake(time.Now())
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)

	cases := []string{"", "не.токен.вообще", "a.b.c.d"}
	for _, c := range cases {
		if _, err := issuer.Parse(c); err == nil {
			t.Errorf("Parse(%q) вернул nil-ошибку, ожидали ошибку", c)
		}
	}
}

func TestIssue_ExpiresAtMatchesTTLFromClock(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)

	token, err := issuer.Issue(uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("Parse() вернул ошибку: %v", err)
	}

	want := start.Add(15 * time.Minute)
	if !claims.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", claims.ExpiresAt.Time, want)
	}
}

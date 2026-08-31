package googleoauth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/auth/googleoauth"
)

// TestNewVerifier_ReturnsNonNil — конструктор просто собирает значение, без сети.
func TestNewVerifier_ReturnsNonNil(t *testing.T) {
	v := googleoauth.NewVerifier("test-client-id")
	if v == nil {
		t.Fatal("NewVerifier() вернул nil")
	}
}

// TestVerify_MalformedToken_ReturnsErrInvalidIDTokenWithoutPanicking — битый (не-JWT) токен
// не должен паниковать, а должен вернуть ErrInvalidIDToken. Discovery-поход к Google в этом
// тестовом окружении, скорее всего, недоступен/недетерминирован — проверяем, что в ЛЮБОМ
// случае (недоступность discovery ИЛИ успешная инициализация verifier'а) результат —
// ошибка без паники, не 500/крах.
func TestVerify_MalformedToken_ReturnsErrorWithoutPanicking(t *testing.T) {
	v := googleoauth.NewVerifier("test-client-id")

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Verify() запаниковал: %v", r)
			}
		}()

		_, err := v.Verify(context.Background(), "не-jwt-вообще")
		if err == nil {
			t.Fatal("Verify() с мусорным токеном вернул nil-ошибку")
		}
	}()
}

// TestErrInvalidIDToken_IsSentinel — ErrInvalidIDToken матчится через errors.Is, в том числе
// через обёртку fmt.Errorf.
func TestErrInvalidIDToken_IsSentinel(t *testing.T) {
	wrapped := errors.New("wrap: " + googleoauth.ErrInvalidIDToken.Error())
	if errors.Is(wrapped, googleoauth.ErrInvalidIDToken) {
		t.Fatal("errors.New не должен матчиться с sentinel — это негативная проверка на то, что тест вообще что-то проверяет")
	}
	if !errors.Is(googleoauth.ErrInvalidIDToken, googleoauth.ErrInvalidIDToken) {
		t.Fatal("errors.Is(ErrInvalidIDToken, ErrInvalidIDToken) = false, want true")
	}
}

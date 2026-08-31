// Package googleoauth — тонкая обёртка над go-oidc для верификации Google ID-токенов
// (подпись/audience/issuer/expiry через discovery-документ Google и его JWKS).
package googleoauth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ErrInvalidIDToken — предъявленный rawIDToken не прошёл верификацию (подпись/issuer/
// audience/expiry) или не распарсился как валидный Google ID-токен.
var ErrInvalidIDToken = errors.New("googleoauth: invalid id token")

// Claims — то, что нужно от Google ID-токена: кто это, подтверждён ли email самим Google.
type Claims struct {
	Subject       string // sub — стабильный уникальный id пользователя у Google
	Email         string
	EmailVerified bool
}

// Verifier проверяет Google ID-токены. Discovery (поход к accounts.google.com за JWKS) —
// ленивая: недоступность Google при старте сервиса не должна ронять запуск ВСЕГО API —
// только этот один эндпоинт временно недоступен.
type Verifier struct {
	clientID string

	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier // nil до первого успешного Verify — ленивая инициализация
}

// NewVerifier создаёт Verifier с client_id вашего Google OAuth-приложения (аудитория токена).
func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID}
}

// Verify проверяет rawIDToken (подпись/issuer/audience/expiry) и возвращает разобранные Claims.
func (v *Verifier) Verify(ctx context.Context, rawIDToken string) (Claims, error) {
	verifier, err := v.getVerifier(ctx)
	if err != nil {
		return Claims{}, fmt.Errorf("googleoauth: init verifier: %w", err)
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return Claims{}, ErrInvalidIDToken
	}

	var raw struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&raw); err != nil {
		return Claims{}, ErrInvalidIDToken
	}

	return Claims{Subject: idToken.Subject, Email: raw.Email, EmailVerified: raw.EmailVerified}, nil
}

func (v *Verifier) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.verifier != nil {
		return v.verifier, nil
	}
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	v.verifier = provider.Verifier(&oidc.Config{ClientID: v.clientID})
	return v.verifier, nil
}

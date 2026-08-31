package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// Principal — аутентифицированный вызывающий, извлечённый RequireAuth из access-токена.
type Principal struct {
	UserID uuid.UUID
	Role   string
}

// principalContextKey — свой тип ключа контекста, чтобы не коллизировать с другими пакетами.
type principalContextKey struct{}

// FromContext достаёт Principal из контекста; второй возврат — false, если его там нет
// (не паникуем на «голом» контексте).
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(Principal)
	return p, ok
}

// RequireAuth — middleware: парсит "Authorization: Bearer <token>", проверяет access-токен
// через issuer и кладёт Principal в контекст. Отсутствует/невалиден/просрочен →
// apperr.Unauthorized через httpx.WriteError, хендлер дальше не вызывается.
func RequireAuth(issuer *jwt.Issuer, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("отсутствует или некорректен заголовок Authorization"))
				return
			}

			claims, err := issuer.Parse(token)
			if err != nil {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("невалидный или истёкший токен"))
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("невалидный или истёкший токен"))
				return
			}

			principal := Principal{UserID: userID, Role: claims.Role}
			ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// bearerToken извлекает токен из заголовка "Authorization: Bearer <token>".
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimPrefix(header, prefix)
	if token == "" {
		return "", false
	}
	return token, true
}

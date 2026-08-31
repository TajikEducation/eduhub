// Package rbac — Principal (аутентифицированный пользователь текущего запроса) и middleware
// RequireAuth/RequireRole. Отдельный пакет от usecase/transport модуля auth, потому что
// на Principal завязываются и другие модули (catalog write, moderation, ... — веха 3+),
// которым не нужно тянуть весь internal/auth/usecase.
package rbac

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// Principal — аутентифицированный пользователь текущего запроса.
type Principal struct {
	UserID uuid.UUID
	Role   string
}

type contextKey struct{}

// NewContext кладёт Principal в контекст запроса.
func NewContext(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// FromContext достаёт Principal из контекста. ok=false — запрос прошёл без RequireAuth
// (публичный эндпоинт) или middleware не сработал; вызывающий обязан проверять ok, не
// полагаться на нулевое значение Principal.
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}

// RequireAuth — middleware: парсит Bearer-токен из Authorization, кладёт Principal в контекст.
// Отсутствие/невалидность токена → 401 через httpx.WriteError, обработчик дальше не вызывается.
func RequireAuth(secret string, logger *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("missing bearer token"))
				return
			}

			claims, err := security.ParseAccessToken(secret, strings.TrimPrefix(header, prefix))
			if err != nil {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("invalid or expired token"))
				return
			}

			ctx := NewContext(r.Context(), Principal{UserID: claims.UserID, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole — middleware: должен идти после RequireAuth. Principal отсутствует в контексте →
// 401 (RequireAuth не был применён — конфигурационная ошибка роутинга, не должна маскироваться
// под 403), роль не входит в allowed → 403.
func RequireRole(logger *slog.Logger, allowed ...string) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := FromContext(r.Context())
			if !ok {
				httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
				return
			}

			for _, role := range allowed {
				if p.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, r, logger, apperr.Forbidden("insufficient role"))
		})
	}
}

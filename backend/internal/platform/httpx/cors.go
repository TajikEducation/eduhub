package httpx

import (
	"net/http"
	"slices"
)

const (
	corsAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowedHeaders = "Content-Type, Authorization, Idempotency-Key"
	corsMaxAge         = "600"
)

// CORS ограничивает cross-origin доступ allowedOrigins: preflight-запросы (OPTIONS с
// Access-Control-Request-Method) обрабатывает сама, обычные запросы — проставляет заголовок
// и передаёт дальше next.
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := slices.Contains(allowedOrigins, origin)

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
					w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
					w.Header().Set("Access-Control-Max-Age", corsMaxAge)
				}
				// Origin не в allowlist — тоже 204, но без CORS-заголовков: браузер сам
				// заблокирует ответ на стороне клиента из-за отсутствия Allow-Origin.
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			next.ServeHTTP(w, r)
		})
	}
}

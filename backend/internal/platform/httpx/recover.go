package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover — middleware, ловит панику в хендлере: без неё net/http не гарантирует чистое закрытие ответа.
// Клиенту уходит обобщённый 500, полный стек — в лог с request_id.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() { //nolint:contextcheck // замыкание использует r.Context() из внешней области, новый контекст не создаёт
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
						slog.String("request_id", RequestID(r.Context())),
					)
					writeInternalServerError(w, logger)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// writeInternalServerError — минимальный JSON 500 при панике (не httpx.WriteError, задача 8 отдельно).
func writeInternalServerError(w http.ResponseWriter, logger *slog.Logger) {
	const body = `{"error":{"code":"internal","message":"internal server error"}}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	// Ошибку не игнорируем: повторная паника тут хуже исходной, поэтому логируем на Warn, не пробрасываем.
	if _, err := w.Write([]byte(body)); err != nil {
		logger.Warn("failed to write recovery response body", slog.Any("error", err))
	}
}

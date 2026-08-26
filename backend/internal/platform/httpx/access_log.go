package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog — middleware, пишет одну запись Info на запрос: method, path, status, duration_ms, request_id.
// Логируется r.URL.Path, не query — query может содержать PII (например, ?q=<имя ребёнка>).
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("request_id", RequestID(r.Context())),
			)
		})
	}
}

// statusWriter перехватывает итоговый HTTP-статус — net/http не даёт доступа к нему после ServeHTTP.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

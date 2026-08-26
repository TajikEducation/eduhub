package httpx

import (
	"encoding/json"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// rateLimitEntry — счётчик запросов одного IP в текущем окне.
type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// RateLimit — in-memory лимитер запросов по IP, модель full-refill-per-window: limit запросов
// разрешено за window, после чего окно полностью сбрасывается. Per-instance счётчик осознанно
// для MVP — при переходе на несколько инстансов сервиса лимитер должен стать Redis-based (веха 5).
func RateLimit(logger *slog.Logger, clk clock.Clock, limit int, window time.Duration) Middleware {
	var (
		mu      sync.Mutex
		entries = make(map[string]*rateLimitEntry)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := clk.Now()

			mu.Lock()

			entry, ok := entries[ip]
			// Ленивая TTL-очистка: запись старше 2*window удаляется при обращении к тому же ключу,
			// полной фоновой чистки всей мапы не делаем (осознанное упрощение MVP).
			if !ok || now.Sub(entry.windowStart) >= 2*window {
				entry = &rateLimitEntry{windowStart: now}
				entries[ip] = entry
			} else if now.Sub(entry.windowStart) >= window {
				entry.count = 0
				entry.windowStart = now
			}

			if entry.count >= limit {
				retryAfter := retryAfterSeconds(now, entry.windowStart, window)
				mu.Unlock()
				writeRateLimitedError(w, r, logger, retryAfter)
				return
			}

			entry.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// retryAfterSeconds считает секунды до конца текущего окна, округляя вверх, минимум 1.
func retryAfterSeconds(now, windowStart time.Time, window time.Duration) int {
	remaining := window - now.Sub(windowStart)
	seconds := int(math.Ceil(remaining.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}

// clientIP достаёт IP клиента из RemoteAddr; если host:port распарсить не удалось, использует
// RemoteAddr как есть.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// writeRateLimitedError пишет 429 в том же JSON-контракте, что и остальные ошибки пакета
// (errorResponse/errorPayload из write_error.go).
func writeRateLimitedError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, retryAfterSecs int) {
	body := errorResponse{
		Error: errorPayload{
			Code:    "rate_limited",
			Message: "too many requests",
		},
		RequestID: RequestID(r.Context()),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		// Маршалинг фиксированной структуры не должен падать — но если случилось,
		// это тоже инцидент, а не бизнес-ошибка.
		logger.Error("failed to marshal rate limit response", slog.Any("error", err))
		payload = []byte(`{"error":{"code":"rate_limited","message":"too many requests"}}`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSecs))
	w.WriteHeader(http.StatusTooManyRequests)

	if _, writeErr := w.Write(payload); writeErr != nil {
		logger.Warn("failed to write rate limit response body", slog.Any("error", writeErr))
	}
}

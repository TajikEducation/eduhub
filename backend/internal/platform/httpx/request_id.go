// Package httpx — транспортный слой: роутер, middleware, JSON in/out, маппинг apperr→HTTP.
package httpx

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

const headerRequestID = "X-Request-ID"

// requestIDKey — свой тип ключа контекста, чтобы не коллизировать с другими пакетами.
type requestIDKey struct{}

// WithRequestID берёт request_id из заголовка или генерирует новый, кладёт в контекст и в ответ.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerRequestID)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(headerRequestID, id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID возвращает request_id из контекста; пустая строка, если его там нет.
func RequestID(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

// newRequestID — свой UUID v4 (16 случайных байт + фиксация версии/варианта), без внешней либы.
func newRequestID() string {
	var b [16]byte
	// crypto/rand.Read падает только при сломанном системном энтропийном источнике —
	// неисправимо, поэтому паника вместо проброса ошибки через весь стек middleware.
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("httpx: crypto/rand unavailable: %w", err))
	}

	b[6] = (b[6] & 0x0f) | 0x40 // версия 4
	b[8] = (b[8] & 0x3f) | 0x80 // вариант RFC 4122

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

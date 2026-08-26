package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Dependency — зависимость, которую Readyz проверяет вызовом Ping перед тем, как считать сервис готовым.
type Dependency struct {
	Name string
	Ping func(ctx context.Context) error
}

// Healthz сообщает, что процесс жив, без обращения к каким-либо зависимостям (liveness, не readiness).
func Healthz(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = WriteJSON(w, logger, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// Readyz проверяет каждую Dependency в пределах timeout и отвечает 200, если все успешны,
// иначе 503 со списком имён упавших/не успевших зависимостей.
func Readyz(logger *slog.Logger, timeout time.Duration, deps ...Dependency) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var failed []string
		for _, dep := range deps {
			if !pingWithTimeout(r.Context(), dep.Ping, timeout) {
				failed = append(failed, dep.Name)
			}
		}

		if len(failed) > 0 {
			_ = WriteJSON(w, logger, http.StatusServiceUnavailable, map[string]any{
				"status": "unavailable",
				"failed": failed,
			})
			return
		}

		_ = WriteJSON(w, logger, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// pingWithTimeout принудительно ограничивает ожидание timeout'ом, даже если сам ping игнорирует
// контекст (например, использует time.Sleep без проверки ctx.Done()): горутина с ping'ом
// не отменяется, а просто "бросается" — канал с буфером 1 не даёт ей заблокироваться навсегда.
func pingWithTimeout(ctx context.Context, ping func(context.Context) error, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- ping(ctx)
	}()

	select {
	case err := <-result:
		return err == nil
	case <-ctx.Done():
		return false
	}
}

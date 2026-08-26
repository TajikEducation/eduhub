package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
)

const (
	readHeaderTimeout = 5 * time.Second // защита от Slowloris — медленных заголовков
	writeTimeout      = 15 * time.Second
)

// Pinger — то, что нужно run от пула БД: проверка доступности и закрытие при остановке.
// Отделён от *pgxpool.Pool специально не через прямую зависимость — тестам нужен дублёр.
type Pinger interface {
	Ping(ctx context.Context) error
	Close()
}

// Deps — всё, что run получает извне: тестам это позволяет подставить фейковый Pinger
// и урезанный handler (например с искусственно медленным роутом), не трогая main().
type Deps struct {
	Logger  *slog.Logger
	Pool    Pinger
	Handler http.Handler
	// Ready, если задан, вызывается один раз сразу после того, как слушатель занял порт —
	// единственный способ узнать реальный адрес, когда HTTPAddr==":0" (тесты просят ОС
	// выбрать свободный порт, чтобы не конфликтовать друг с другом).
	Ready func(addr string)
}

// run поднимает HTTP-сервер на deps.Handler и блокируется до отмены ctx, затем делает
// graceful shutdown в пределах cfg.ShutdownTimeout и закрывает deps.Pool.
func run(ctx context.Context, cfg config.Config, deps Deps) error {
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	if deps.Ready != nil {
		deps.Ready(ln.Addr().String())
	}

	srv := &http.Server{
		Handler:           deps.Handler,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}

	deps.Logger.Info("listening", slog.String("addr", ln.Addr().String()))

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	deps.Pool.Close()
	deps.Logger.Info("shutdown complete")
	return nil
}

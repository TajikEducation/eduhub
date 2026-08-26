// Command api — точка входа HTTP API бэкенда EduHub.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// readyzTimeout — сколько ждём ответа от каждой зависимости в /readyz.
const readyzTimeout = 2 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Логгера ещё нет — конфигурация обязательна для его настройки (уровень).
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, "eduhub-api", os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pg.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to open db pool", slog.Any("error", err))
		os.Exit(1)
	}

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))
	router.Handle("GET /readyz", httpx.Readyz(log, readyzTimeout, httpx.Dependency{Name: "db", Ping: pool.Ping}))

	handler := httpx.Chain(httpx.WithRequestID, httpx.AccessLog(log), httpx.Recover(log))(router)

	deps := Deps{Logger: log, Pool: pool, Handler: handler}
	if err := run(ctx, cfg, deps); err != nil {
		log.Error("server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

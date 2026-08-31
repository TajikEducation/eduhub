// Command api — точка входа HTTP API бэкенда EduHub.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	authpg "github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	catalogpg "github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/rediscache"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	cache := rediscache.New(rdb)

	argonParams := password.DefaultParams
	if cfg.ArgonMemoryKiB != 0 {
		argonParams.MemoryKiB = cfg.ArgonMemoryKiB
	}
	if cfg.ArgonIterations != 0 {
		argonParams.Iterations = cfg.ArgonIterations
	}
	if cfg.ArgonParallelism != 0 {
		argonParams.Parallelism = cfg.ArgonParallelism
	}
	hasher := password.New(argonParams)

	userRepo := authpg.NewUserRepo(pool)
	refreshTokenRepo := authpg.NewRefreshTokenRepo(pool)
	oauthRepo := authpg.NewOAuthRepo(pool)

	handler := newHandler(
		log, cfg.CORSAllowedOrigins, []httpx.Dependency{{Name: "db", Ping: pool.Ping}},
		catalogpg.New(pool), cache,
		userRepo, refreshTokenRepo, oauthRepo, hasher, []byte(cfg.JWTSecret), clock.New(), cfg.GoogleClientID,
	)

	deps := Deps{Logger: log, Pool: pool, Handler: handler}
	if err := run(ctx, cfg, deps); err != nil {
		log.Error("server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

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

	applicantspg "github.com/abdulhalim/eduhub/backend/internal/applicants/repo/postgres"
	authpg "github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	authusecase "github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	catalogpg "github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/rediscache"
	chatpg "github.com/abdulhalim/eduhub/backend/internal/communications/repo/postgres"
	moderationpg "github.com/abdulhalim/eduhub/backend/internal/moderation/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
	reviewspg "github.com/abdulhalim/eduhub/backend/internal/reviews/repo/postgres"
	vacanciespg "github.com/abdulhalim/eduhub/backend/internal/vacancies/repo/postgres"
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

	authSvc := authusecase.New(authpg.New(pool), authpg.NewRefreshTokenRepo(pool), clock.New(), authusecase.Config{
		JWTSecret:  cfg.JWTSecret,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
	})

	auditRecorder := moderationpg.New(pool)
	reviewsRepo := reviewspg.New(pool)
	vacanciesRepo := vacanciespg.New(pool)
	conversationRepo := chatpg.NewConversationRepo(pool)
	messageRepo := chatpg.NewMessageRepo(pool)
	applicantRepo := applicantspg.NewApplicantRepo(pool)
	applicantAchievementRepo := applicantspg.NewAchievementRepo(pool)
	applicationRepo := applicantspg.NewApplicationRepo(pool)

	handler := newHandler(log, cfg.CORSAllowedOrigins, []httpx.Dependency{{Name: "db", Ping: pool.Ping}}, catalogpg.New(pool), cache, authSvc, cfg.JWTSecret, auditRecorder, reviewsRepo, vacanciesRepo, conversationRepo, messageRepo, applicantRepo, applicantAchievementRepo, applicationRepo)

	deps := Deps{Logger: log, Pool: pool, Handler: handler}
	if err := run(ctx, cfg, deps); err != nil {
		log.Error("server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

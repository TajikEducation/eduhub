// Command devseed засеивает 9 демо-институций (перенесённых из web/lib/data.ts) в
// PostgreSQL для локальной разработки. Отказывает, если APP_ENV != dev — ДО открытия
// соединения с БД и начала транзакции (docs/EduHub_Backend_Development_Plan.md, задача 30).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "devseed: config: %v\n", err)
		os.Exit(1)
	}

	if err := Seed(context.Background(), cfg.AppEnv, cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "devseed: %v\n", err)
		os.Exit(1)
	}
}

// Seed проверяет, что appEnv=="dev", и только затем открывает пул подключений к БД и
// запускает SeedAll. Вынесена из main для тестируемости (см. cmd/devseed/seed_integration_test.go,
// RED-кейс (в): при appEnv!="dev" не должно быть ни соединения с БД, ни транзакции).
func Seed(ctx context.Context, appEnv, databaseURL string) error {
	if appEnv != "dev" {
		return fmt.Errorf("APP_ENV=%q, devseed допустим только при APP_ENV=dev", appEnv)
	}

	pool, err := pg.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open db pool: %w", err)
	}
	defer pool.Close()

	return SeedAll(ctx, pool)
}

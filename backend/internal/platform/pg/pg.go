// Package pg отвечает за создание пула подключений к PostgreSQL.
package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectTimeout — недостижимый хост не должен вешать вызывающего дольше этого срока.
const connectTimeout = 5 * time.Second

// Open создаёт пул подключений к PostgreSQL и проверяет его пингом перед возвратом.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pg: parse config: %w", err)
	}

	cfg.ConnConfig.ConnectTimeout = connectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pg: new pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg: ping failed: %w", err)
	}

	return pool, nil
}

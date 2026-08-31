package usecase

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss — сигнатура промаха кэша, отличается от реальной ошибки инфраструктуры
// (сеть/таймаут Redis) — только промах не считается деградацией, это ожидаемый путь.
var ErrCacheMiss = errors.New("usecase: cache miss")

// CacheClient — порт кэша листинга каталога. Реализация — internal/catalog/repo/rediscache.
// Любая ошибка КРОМЕ ErrCacheMiss трактуется вызывающим (CachedService) как деградация:
// запрос идёт напрямую в InstitutionRepo, кэш не является точкой отказа.
type CacheClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Version(ctx context.Context) (int64, error)
}

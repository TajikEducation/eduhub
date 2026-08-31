package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// CachedService — декоратор над Service, добавляющий кэш листинга каталога (Redis, задача 31).
// Кэш — не точка отказа: любая ошибка инфраструктуры кэша (в отличие от ErrCacheMiss)
// деградирует к прямому вызову inner.List, ответ пользователю не портится.
type CachedService struct {
	inner *Service
	cache CacheClient
	ttl   time.Duration
	sf    singleflight.Group
	log   *slog.Logger
}

// NewCachedService создаёт CachedService поверх inner Service с заданным CacheClient и TTL.
func NewCachedService(inner *Service, cache CacheClient, ttl time.Duration, log *slog.Logger) *CachedService {
	return &CachedService{inner: inner, cache: cache, ttl: ttl, log: log}
}

// List возвращает страницу институций, кэшируя результат в Redis по нормализованному
// фильтру. Схлопывает параллельные идентичные запросы через singleflight.
func (c *CachedService) List(ctx context.Context, f domain.Filter) (domain.ListResult, error) {
	f.Statuses = approvedOnly
	f.Normalize()

	version, err := c.cache.Version(ctx)
	if err != nil {
		c.log.Warn("catalog cache: version unavailable, degrading to repo", slog.Any("error", err))
		return c.inner.List(ctx, f)
	}

	key, err := listCacheKey(version, f)
	if err != nil {
		c.log.Warn("catalog cache: failed to build cache key, degrading to repo", slog.Any("error", err))
		return c.inner.List(ctx, f)
	}

	v, err, _ := c.sf.Do(key, func() (any, error) {
		return c.listWithCache(ctx, key, f)
	})
	if err != nil {
		return domain.ListResult{}, err
	}
	return v.(domain.ListResult), nil
}

// listWithCache — тело singleflight-группы: читает кэш, при промахе/ошибке кэша вызывает
// inner.List и best-effort записывает результат обратно в кэш.
func (c *CachedService) listWithCache(ctx context.Context, key string, f domain.Filter) (domain.ListResult, error) {
	raw, err := c.cache.Get(ctx, key)
	switch {
	case err == nil:
		var cached domain.ListResult
		if unmarshalErr := json.Unmarshal(raw, &cached); unmarshalErr == nil {
			return cached, nil
		}
		c.log.Warn("catalog cache: corrupted cache entry, treating as miss", slog.String("key", key))
	case errors.Is(err, ErrCacheMiss):
		// ожидаемый путь — идём в репо ниже.
	default:
		c.log.Warn("catalog cache: get failed, degrading to repo", slog.Any("error", err))
	}

	result, err := c.inner.List(ctx, f)
	if err != nil {
		return domain.ListResult{}, err
	}

	if data, marshalErr := json.Marshal(result); marshalErr != nil {
		c.log.Warn("catalog cache: failed to marshal result for caching", slog.Any("error", marshalErr))
	} else if setErr := c.cache.Set(ctx, key, data, c.ttl); setErr != nil {
		c.log.Warn("catalog cache: set failed", slog.Any("error", setErr))
	}

	return result, nil
}

// Get проксирует напрямую в inner — листинг кэшируется этой задачей, карточка институции
// уже имеет собственный ETag-механизм (задача 28), отдельное кэширование было бы избыточным.
func (c *CachedService) Get(ctx context.Context, id uuid.UUID) (domain.Institution, error) {
	return c.inner.Get(ctx, id)
}

// listCacheKey строит ключ кэша листинга: catalog:list:v{version}:{sha256(filter)}.
func listCacheKey(version int64, f domain.Filter) (string, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("usecase: marshal filter for cache key: %w", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("catalog:list:v%d:%s", version, hex.EncodeToString(sum[:])), nil
}

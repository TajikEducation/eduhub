// Package rediscache — реализация usecase.CacheClient поверх Redis (задача 31).
package rediscache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
)

// versionKey — ключ Redis, хранящий текущую версию каталога (инкрементируется в вехе 3
// при мутациях; эта задача только читает).
const versionKey = "catalog:version"

// Client — адаптер usecase.CacheClient поверх *redis.Client.
type Client struct {
	rdb *redis.Client
}

// New создаёт Client поверх переданного *redis.Client.
func New(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// Get читает значение по ключу. Отсутствие ключа маппится в usecase.ErrCacheMiss,
// любая другая ошибка Redis — оборачивается и возвращается как есть (не ErrCacheMiss),
// чтобы CachedService мог отличить промах от сбоя инфраструктуры.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, usecase.ErrCacheMiss
		}
		return nil, fmt.Errorf("rediscache: get %s: %w", key, err)
	}
	return val, nil
}

// Set записывает значение с TTL.
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("rediscache: set %s: %w", key, err)
	}
	return nil
}

// Version читает текущую версию каталога. Отсутствие ключа — не ошибка, а ожидаемое
// начальное состояние (0) до первой записи версии в вехе 3.
func (c *Client) Version(ctx context.Context) (int64, error) {
	v, err := c.rdb.Get(ctx, versionKey).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, fmt.Errorf("rediscache: version: %w", err)
	}
	return v, nil
}

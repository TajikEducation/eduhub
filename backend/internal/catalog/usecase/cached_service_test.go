package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
)

// countingRepo — fakeRepo с подсчётом вызовов List, потокобезопасный (нужен для
// параллельного RED-кейса на singleflight-схлопывание).
type countingRepo struct {
	mu     sync.Mutex
	calls  int
	result domain.ListResult
	err    error
	delay  time.Duration
}

func (r *countingRepo) List(ctx context.Context, filter domain.Filter) (domain.ListResult, error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	return r.result, r.err
}

func (r *countingRepo) GetByID(_ context.Context, _ uuid.UUID) (domain.Institution, error) {
	return domain.Institution{}, nil
}

// IsApproved не используется тестами в этом файле — реализован только чтобы countingRepo
// продолжал удовлетворять usecase.InstitutionRepo после добавления метода в интерфейс (E2.6).
func (r *countingRepo) IsApproved(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (r *countingRepo) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// fakeCache — in-memory реализация usecase.CacheClient для юнит-тестов (без реального Redis).
type fakeCache struct {
	mu      sync.Mutex
	data    map[string][]byte
	version int64
	getErr  error // если не nil — Get всегда возвращает эту ошибку (симуляция сбоя Redis)
	verErr  error // если не nil — Version всегда возвращает эту ошибку
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: make(map[string][]byte)}
}

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.getErr != nil {
		return nil, c.getErr
	}
	v, ok := c.data[key]
	if !ok {
		return nil, usecase.ErrCacheMiss
	}
	return v, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *fakeCache) Version(_ context.Context) (int64, error) {
	if c.verErr != nil {
		return 0, c.verErr
	}
	return c.version, nil
}

func (c *fakeCache) keyCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCache_SecondIdenticalRequest_DoesNotReachRepo(t *testing.T) {
	repo := &countingRepo{result: domain.ListResult{Items: []domain.Institution{{Region: "dushanbe"}}}}
	inner := usecase.New(repo)
	cache := newFakeCache()
	svc := usecase.NewCachedService(inner, cache, time.Minute, discardLogger())

	filter := domain.Filter{Region: strPtr("dushanbe")}

	if _, err := svc.List(context.Background(), filter); err != nil {
		t.Fatalf("List() 1st call unexpected error: %v", err)
	}
	if _, err := svc.List(context.Background(), filter); err != nil {
		t.Fatalf("List() 2nd call unexpected error: %v", err)
	}

	if got := repo.callCount(); got != 1 {
		t.Fatalf("repo.calls = %d, want 1", got)
	}
}

func TestCache_DifferentFilters_DifferentKeys(t *testing.T) {
	repo := &countingRepo{result: domain.ListResult{}}
	inner := usecase.New(repo)
	cache := newFakeCache()
	svc := usecase.NewCachedService(inner, cache, time.Minute, discardLogger())

	if _, err := svc.List(context.Background(), domain.Filter{Region: strPtr("dushanbe")}); err != nil {
		t.Fatalf("List() dushanbe unexpected error: %v", err)
	}
	if _, err := svc.List(context.Background(), domain.Filter{Region: strPtr("sughd")}); err != nil {
		t.Fatalf("List() sughd unexpected error: %v", err)
	}

	if got := repo.callCount(); got != 2 {
		t.Fatalf("repo.calls = %d, want 2", got)
	}
	if got := cache.keyCount(); got != 2 {
		t.Fatalf("cache keys = %d, want 2", got)
	}
}

func TestCache_50ParallelIdenticalRequests_RepoCalledOnce(t *testing.T) {
	repo := &countingRepo{result: domain.ListResult{}, delay: 20 * time.Millisecond}
	inner := usecase.New(repo)
	cache := newFakeCache()
	svc := usecase.NewCachedService(inner, cache, time.Minute, discardLogger())

	filter := domain.Filter{Region: strPtr("dushanbe")}

	const n = 50
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			if _, err := svc.List(context.Background(), filter); err != nil {
				t.Errorf("List() unexpected error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if got := repo.callCount(); got != 1 {
		t.Fatalf("repo.calls = %d, want 1", got)
	}
}

func TestCache_CacheGetError_DegradesToRepo(t *testing.T) {
	repo := &countingRepo{result: domain.ListResult{Items: []domain.Institution{{Region: "dushanbe"}}}}
	inner := usecase.New(repo)
	cache := &fakeCache{data: make(map[string][]byte), getErr: errors.New("redis: connection refused")}
	svc := usecase.NewCachedService(inner, cache, time.Minute, discardLogger())

	result, err := svc.List(context.Background(), domain.Filter{Region: strPtr("dushanbe")})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Region != "dushanbe" {
		t.Fatalf("result = %+v, want repo result passthrough", result)
	}
	if got := repo.callCount(); got != 1 {
		t.Fatalf("repo.calls = %d, want 1", got)
	}
}

func TestCache_VersionError_DegradesToRepo(t *testing.T) {
	repo := &countingRepo{result: domain.ListResult{Items: []domain.Institution{{Region: "dushanbe"}}}}
	inner := usecase.New(repo)
	cache := &fakeCache{data: make(map[string][]byte), verErr: errors.New("redis: timeout")}
	svc := usecase.NewCachedService(inner, cache, time.Minute, discardLogger())

	result, err := svc.List(context.Background(), domain.Filter{Region: strPtr("dushanbe")})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("result.Items = %v, want repo result passthrough", result.Items)
	}
	if got := repo.callCount(); got != 1 {
		t.Fatalf("repo.calls = %d, want 1", got)
	}
}

func strPtr(s string) *string { return &s }

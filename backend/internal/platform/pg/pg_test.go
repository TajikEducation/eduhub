//go:build integration

package pg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testDatabaseURL возвращает TEST_DATABASE_URL из окружения или пропускает тест, если она пуста.
func testDatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, пропускаем интеграционный тест")
	}
	return url
}

func TestOpen_ConnectsAndPings(t *testing.T) {
	url := testDatabaseURL(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping() вернул ошибку: %v", err)
	}
}

func TestOpen_UnreachableHostTimesOut(t *testing.T) {
	// Хост из TEST-NET-1 (RFC 5737) не роутится наружу — соединение зависнет до таймаута.
	const unreachableURL = "postgres://eduhub:eduhub@10.255.255.1:5432/eduhub?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	done := make(chan struct{})
	var pool *pgxpool.Pool
	var err error

	go func() {
		defer close(done)
		p, e := Open(ctx, unreachableURL)
		pool, err = p, e
	}()

	select {
	case <-done:
	case <-time.After(7 * time.Second):
		t.Fatal("Open() не завершился за 7 секунд — недостижимый хост должен вызывать таймаут внутри connectTimeout")
	}

	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatal("Open() с недостижимым хостом должен вернуть ошибку, получен nil")
	}
}

func TestOpen_SkipsWithoutTestDatabaseURL(t *testing.T) {
	// t.Skip прерывает горутину теста (runtime.Goexit), поэтому проверяем поведение
	// через subtest: если testDatabaseURL не вызвал Skip, subtest падает на Fatal ниже.
	t.Run("skip", func(t *testing.T) {
		t.Setenv("TEST_DATABASE_URL", "")

		_ = testDatabaseURL(t)
		t.Fatal("testDatabaseURL должен был вызвать t.Skip раньше этой строки")
	})
}

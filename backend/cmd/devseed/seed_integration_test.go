//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// testDatabaseURL возвращает TEST_DATABASE_URL из окружения или пропускает тест, если она
// пуста (паттерн — internal/catalog/repo/postgres/*_test.go). Миграции 00001-00005
// ожидаются уже применёнными (см. Makefile migrate-up / CI job backend).
func testDatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, пропускаем интеграционный тест")
	}
	return url
}

// countApprovedInstitutions — число институций catalog.institutions со moderation_status='approved'.
func countApprovedInstitutions(ctx context.Context, t *testing.T, url string) int {
	t.Helper()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM catalog.institutions WHERE moderation_status = 'approved'`).Scan(&count); err != nil {
		t.Fatalf("count institutions: %v", err)
	}
	return count
}

// countSeedRefs — число записей platform.seed_refs с entity_type='institution'.
func countInstitutionSeedRefs(ctx context.Context, t *testing.T, url string) int {
	t.Helper()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM platform.seed_refs WHERE entity_type = 'institution'`).Scan(&count); err != nil {
		t.Fatalf("count seed_refs: %v", err)
	}
	return count
}

// resetSeedState очищает всё, что мог оставить сидер — catalog.institutions (каскадно все
// сателлиты) и platform.seed_refs. В отличие от internal/catalog/repo/postgres, эти тесты
// не оборачиваются в rollback-транзакцию (Seed сам коммитит по институции), поэтому без
// явной уборки прогон `go test -tags=integration ./...` на общей TEST_DATABASE_URL оставляет
// 9 реальных approved-институций для ЛЮБОГО пакета, который выполнится после (порядок пакетов
// в go test — по имени пути, cmd/devseed идёт раньше internal/catalog/repo/postgres) — эти
// институции ломают фикстуры repo/postgres, которые предполагают только свои 3-25 строк.
// Вызывается и до, и после каждого теста — не зависит от того, что оставил предыдущий прогон.
func resetSeedState(t *testing.T, url string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, "TRUNCATE catalog.institutions CASCADE"); err != nil {
		t.Fatalf("truncate catalog.institutions: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE platform.seed_refs"); err != nil {
		t.Fatalf("truncate platform.seed_refs: %v", err)
	}
}

func TestSeed_FirstRun_Creates9ApprovedInstitutions(t *testing.T) {
	url := testDatabaseURL(t)
	resetSeedState(t, url)
	t.Cleanup(func() { resetSeedState(t, url) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := Seed(ctx, "dev", url); err != nil {
		t.Fatalf("Seed() вернул ошибку: %v", err)
	}

	if got := countApprovedInstitutions(ctx, t, url); got != 9 {
		t.Fatalf("после первого прогона ожидали 9 институций со status=approved, получили %d", got)
	}
}

func TestSeed_SecondRun_IsIdempotent(t *testing.T) {
	url := testDatabaseURL(t)
	resetSeedState(t, url)
	t.Cleanup(func() { resetSeedState(t, url) })

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := Seed(ctx, "dev", url); err != nil {
		t.Fatalf("Seed() (1-й прогон) вернул ошибку: %v", err)
	}
	firstCount := countApprovedInstitutions(ctx, t, url)
	firstRefs := countInstitutionSeedRefs(ctx, t, url)

	if err := Seed(ctx, "dev", url); err != nil {
		t.Fatalf("Seed() (2-й прогон) вернул ошибку: %v", err)
	}
	secondCount := countApprovedInstitutions(ctx, t, url)
	secondRefs := countInstitutionSeedRefs(ctx, t, url)

	if secondCount != 9 {
		t.Fatalf("после повторного прогона ожидали по-прежнему 9 институций, получили %d", secondCount)
	}
	if secondCount != firstCount {
		t.Fatalf("повторный прогон изменил число институций: было %d, стало %d", firstCount, secondCount)
	}
	if secondRefs != firstRefs {
		t.Fatalf("повторный прогон изменил число seed_refs (institution): было %d, стало %d", firstRefs, secondRefs)
	}
}

func TestSeed_NonDevAppEnv_RefusesWithoutTouchingDB(t *testing.T) {
	url := testDatabaseURL(t)
	resetSeedState(t, url)
	t.Cleanup(func() { resetSeedState(t, url) })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	before := countApprovedInstitutions(ctx, t, url)

	err := Seed(ctx, "prod", url)
	if err == nil {
		t.Fatal("Seed() с APP_ENV=prod должен вернуть ошибку, получили nil")
	}

	after := countApprovedInstitutions(ctx, t, url)
	if after != before {
		t.Fatalf("Seed() с APP_ENV=prod не должен менять данные: до %d, после %d", before, after)
	}
}

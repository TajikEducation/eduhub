//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// p — тестовый хелпер: адрес переданного значения.
func p[T any](v T) *T {
	return &v
}

// testDatabaseURL возвращает TEST_DATABASE_URL из окружения или пропускает тест, если она пуста.
func testDatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, пропускаем интеграционный тест")
	}
	return url
}

// nameSet собирает множество RU-названий институций для сравнения без учёта порядка.
func nameSet(insts []domain.Institution) map[string]bool {
	set := make(map[string]bool, len(insts))
	for _, inst := range insts {
		set[inst.Name.RU] = true
	}
	return set
}

func TestListInstitutions(t *testing.T) {
	url := testDatabaseURL(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() вернул ошибку: %v", err)
	}
	defer tx.Rollback(ctx)

	// Фикстуры — 5 институций, согласованных под все 7 subtests ниже.
	fixtures := []string{
		// 1. Гулистон — approved, dushanbe, school, bilingual, rating 4.8, discount=true.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, curriculum, rating_avg, discount_available, moderation_status)
		 VALUES
			('{"ru":"Гулистон","tg":"Гулистон"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.78,38.56)::geography, '{bilingual}', 4.8, true, 'approved')`,
		// 2. Лицей №5 — approved, sughd, school, stem, rating NULL, discount=false.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, curriculum, rating_avg, discount_available, moderation_status)
		 VALUES
			('{"ru":"Лицей №5","tg":"Литсейи №5"}', '{school}', 'sughd',
			 ST_MakePoint(69.62,40.28)::geography, '{stem}', NULL, false, 'approved')`,
		// 3. Детский сад Регар — approved, dushanbe, kindergarten, curriculum пустой, rating 4.2.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, curriculum, rating_avg, discount_available, moderation_status)
		 VALUES
			('{"ru":"Детский сад Регар","tg":"Боғчаи Регар"}', '{kindergarten}', 'dushanbe',
			 ST_MakePoint(68.78,38.56)::geography, '{}', 4.2, false, 'approved')`,
		// 4. На модерации — pending, не должна попадать ни в один результат ниже.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, moderation_status)
		 VALUES
			('{"ru":"На модерации","tg":"Дар назорат"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.78,38.56)::geography, 'pending')`,
		// 5. Отклонённая заявка — rejected, не должна попадать ни в один результат ниже.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, moderation_status)
		 VALUES
			('{"ru":"Отклонённая заявка","tg":"Дархости рад"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.78,38.56)::geography, 'rejected')`,
	}
	for _, insertSQL := range fixtures {
		if _, err := tx.Exec(ctx, insertSQL); err != nil {
			t.Fatalf("вставка фикстуры не удалась: %v\nSQL: %s", err, insertSQL)
		}
	}

	repo := postgres.New(tx)

	t.Run("а) базовый фильтр approved возвращает только 1,2,3", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true, "Лицей №5": true, "Детский сад Регар": true}
		if got := nameSet(got); len(got) != len(want) || !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("б) Region=sughd возвращает только Лицей №5", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Region: p("sughd")})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Лицей №5": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("в) Types=[school] исключает kindergarten", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Types: []string{"school"}})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true, "Лицей №5": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("г) Q=гулис находит по подстроке в ru-названии", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Q: p("гулис")})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("д) MinRating=4.5 исключает NULL и более низкий рейтинг", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, MinRating: p(4.5)})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("е) Curriculum=[bilingual,stem] пересечение", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Curriculum: []string{"bilingual", "stem"}})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true, "Лицей №5": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("ж) Discount=true возвращает только Гулистон", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Discount: p(true)})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гулистон": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})
}

// TestListInstitutions_NullableArrayFields проверяет, что nullable TEXT[]-колонки
// (curriculum/languages/program_level/discount_type), реально хранящие SQL NULL (не '{}'),
// возвращаются в domain.Institution как non-nil пустые слайсы — контракт domain.NewInstitution
// (задача 20): "curriculum":[] в JSON, никогда null.
func TestListInstitutions_NullableArrayFields(t *testing.T) {
	url := testDatabaseURL(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() вернул ошибку: %v", err)
	}
	defer tx.Rollback(ctx)

	// Институция без явных curriculum/languages/program_level/discount_type —
	// реальный SQL NULL (nullable-колонки без DEFAULT), не пустой массив '{}'.
	const insertSQL = `INSERT INTO catalog.institutions
			(name, types, region, geo, moderation_status)
		 VALUES
			('{"ru":"Без массивов","tg":"Бе массивхо"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.78,38.56)::geography, 'approved')`
	if _, err := tx.Exec(ctx, insertSQL); err != nil {
		t.Fatalf("вставка фикстуры не удалась: %v", err)
	}

	repo := postgres.New(tx)

	got, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Q: p("Без массивов")})
	if err != nil {
		t.Fatalf("List() вернул ошибку: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ожидали ровно 1 институцию, получили %d", len(got))
	}

	inst := got[0]
	if inst.Curriculum == nil {
		t.Error("Curriculum должен быть non-nil при SQL NULL в БД")
	}
	if inst.Languages == nil {
		t.Error("Languages должен быть non-nil при SQL NULL в БД")
	}
	if inst.ProgramLevel == nil {
		t.Error("ProgramLevel должен быть non-nil при SQL NULL в БД")
	}
	if inst.DiscountType == nil {
		t.Error("DiscountType должен быть non-nil при SQL NULL в БД")
	}
}

func equalSets(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

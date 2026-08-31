//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

func TestListPagination(t *testing.T) {
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

	// 25 институций, все approved, различаются только ценой: 100, 200, ..., 2500.
	for i := 1; i <= 25; i++ {
		price := i * 100
		name := fmt.Sprintf("Институция %02d", i)
		// rating_avg=4.0 константа на всех строках: subtest (д) нужен непустой NextCursor
		// для sort=score (иначе все rating_avg NULL → нет last.RatingAvg → NextCursor не строится),
		// сама сортировка по score в этом файле не проверяется (это задача 22).
		_, err := tx.Exec(ctx, `
			INSERT INTO catalog.institutions (name, types, region, geo, price, rating_avg, moderation_status)
			VALUES ($1, '{school}', 'dushanbe', ST_MakePoint(68.78,38.56)::geography, $2, 4.0, 'approved')`,
			fmt.Sprintf(`{"ru":"%s","tg":"%s"}`, name, name), price,
		)
		if err != nil {
			t.Fatalf("вставка фикстуры #%d не удалась: %v", i, err)
		}
	}

	repo := postgres.New(tx)

	prices := func(insts []domain.Institution) []int {
		out := make([]int, len(insts))
		for i, inst := range insts {
			if inst.Price == nil {
				t.Fatalf("институция %q без цены — фикстуры некорректны", inst.Name.RU)
			}
			out[i] = *inst.Price
		}
		return out
	}

	var page1Cursor, page2Cursor *string

	t.Run("а) первая страница — 10 записей, цены 100..1000, next_cursor задан", func(t *testing.T) {
		result, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "price_asc", Limit: 10,
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 10 {
			t.Fatalf("len(Items) = %d, want 10", len(result.Items))
		}
		if result.NextCursor == nil {
			t.Fatal("NextCursor == nil, ожидали курсор на вторую страницу")
		}
		got := prices(result.Items)
		for i, want := 0, 100; i < 10; i, want = i+1, want+100 {
			if got[i] != want {
				t.Errorf("Items[%d].Price = %d, want %d", i, got[i], want)
			}
		}
		page1Cursor = result.NextCursor
	})

	t.Run("б) вторая страница — следующие 10, без пропусков/пересечений", func(t *testing.T) {
		if page1Cursor == nil {
			t.Fatal("page1Cursor не задан — subtest (а) должен выполниться первым")
		}
		result, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "price_asc", Limit: 10, Cursor: page1Cursor,
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 10 {
			t.Fatalf("len(Items) = %d, want 10", len(result.Items))
		}
		if result.NextCursor == nil {
			t.Fatal("NextCursor == nil, ожидали курсор на третью страницу")
		}
		got := prices(result.Items)
		for i, want := 0, 1100; i < 10; i, want = i+1, want+100 {
			if got[i] != want {
				t.Errorf("Items[%d].Price = %d, want %d", i, got[i], want)
			}
		}
		if got[0] != 1100 {
			t.Errorf("первая цена второй страницы = %d, want 1100 (без пропусков/пересечений с первой страницей)", got[0])
		}
		page2Cursor = result.NextCursor
	})

	t.Run("в) третья страница — остаток 5 записей, next_cursor пуст", func(t *testing.T) {
		if page2Cursor == nil {
			t.Fatal("page2Cursor не задан — subtest (б) должен выполниться первым")
		}
		result, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "price_asc", Limit: 10, Cursor: page2Cursor,
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 5 {
			t.Fatalf("len(Items) = %d, want 5", len(result.Items))
		}
		if result.NextCursor != nil {
			t.Errorf("NextCursor = %v, want nil (последняя страница)", *result.NextCursor)
		}
		got := prices(result.Items)
		for i, want := 0, 2100; i < 5; i, want = i+1, want+100 {
			if got[i] != want {
				t.Errorf("Items[%d].Price = %d, want %d", i, got[i], want)
			}
		}
	})

	t.Run("г) битый курсор — Invalid, не паника", func(t *testing.T) {
		_, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "price_asc", Limit: 10,
			Cursor: p("это-не-валидный-base64!!!"),
		})
		if err == nil {
			t.Fatal("List() с битым курсором вернул nil-ошибку, ожидали apperr.ErrInvalid")
		}
		if !errors.Is(err, apperr.ErrInvalid) {
			t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
		}
	})

	t.Run("д) курсор с sort=score отвергается при sort=price_asc", func(t *testing.T) {
		scoreResult, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "score", Limit: 10,
		})
		if err != nil {
			t.Fatalf("List() (sort=score) вернул ошибку: %v", err)
		}
		if scoreResult.NextCursor == nil {
			t.Fatal("NextCursor (sort=score) == nil, ожидали непустой курсор для проверки несовпадения sort")
		}

		_, err = repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"}, Sort: "price_asc", Limit: 10,
			Cursor: scoreResult.NextCursor,
		})
		if err == nil {
			t.Fatal("List() с курсором от sort=score при sort=price_asc вернул nil-ошибку, ожидали apperr.ErrInvalid")
		}
		if !errors.Is(err, apperr.ErrInvalid) {
			t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
		}
	})
}

// TestListPagination_DefaultSort — баг, найденный при ручном тестировании через Postman:
// при пустом Filter.Sort (обычный листинг каталога без явного sort=) курсор keyset-пагинации
// молча не строился, хотя строк больше чем Limit (switch f.Sort в List() не имел ветки для "").
// Обычный просмотр каталога без сортировки — самый частый сценарий, поэтому это не косметика.
func TestListPagination_DefaultSort(t *testing.T) {
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

	// 25 институций с РАЗЛИЧНЫМИ created_at (now() в одной транзакции даёт одно и то же
	// значение на все строки — без явного смещения created_at все 25 фикстур оказались бы
	// на одной временной метке, и тест не проверял бы порядок, только tie-break по id).
	// Строка #25 — самая новая (offset=0), #1 — самая старая (offset=24с).
	names := make([]string, 25)
	for i := 1; i <= 25; i++ {
		offsetSeconds := 25 - i
		name := fmt.Sprintf("Дефолт-институция %02d", i)
		names[i-1] = name
		_, err := tx.Exec(ctx, `
			INSERT INTO catalog.institutions (name, types, region, geo, moderation_status, created_at)
			VALUES ($1, '{school}', 'dushanbe', ST_MakePoint(68.78,38.56)::geography, 'approved', now() - ($2 * interval '1 second'))`,
			fmt.Sprintf(`{"ru":"%s","tg":"%s"}`, name, name), offsetSeconds,
		)
		if err != nil {
			t.Fatalf("вставка фикстуры #%d не удалась: %v", i, err)
		}
	}

	repo := postgres.New(tx)

	ruNames := func(insts []domain.Institution) []string {
		out := make([]string, len(insts))
		for i, inst := range insts {
			out[i] = inst.Name.RU
		}
		return out
	}

	var page1Cursor, page2Cursor *string
	var seen []string

	t.Run("а) первая страница — 10 самых новых записей, next_cursor задан", func(t *testing.T) {
		result, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Limit: 10})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 10 {
			t.Fatalf("len(Items) = %d, want 10", len(result.Items))
		}
		if result.NextCursor == nil {
			t.Fatal("NextCursor == nil при пустом Sort — баг: пагинация каталога без явной сортировки не работает")
		}
		got := ruNames(result.Items)
		for i, want := 0, 25; i < 10; i, want = i+1, want-1 {
			wantName := fmt.Sprintf("Дефолт-институция %02d", want)
			if got[i] != wantName {
				t.Errorf("Items[%d].Name.RU = %q, want %q (порядок: новые сначала)", i, got[i], wantName)
			}
		}
		seen = append(seen, got...)
		page1Cursor = result.NextCursor
	})

	t.Run("б) вторая страница — следующие 10, без пропусков/пересечений", func(t *testing.T) {
		if page1Cursor == nil {
			t.Fatal("page1Cursor не задан — subtest (а) должен выполниться первым")
		}
		result, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Limit: 10, Cursor: page1Cursor})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 10 {
			t.Fatalf("len(Items) = %d, want 10", len(result.Items))
		}
		if result.NextCursor == nil {
			t.Fatal("NextCursor == nil, ожидали курсор на третью страницу")
		}
		got := ruNames(result.Items)
		for i, want := 0, 15; i < 10; i, want = i+1, want-1 {
			wantName := fmt.Sprintf("Дефолт-институция %02d", want)
			if got[i] != wantName {
				t.Errorf("Items[%d].Name.RU = %q, want %q", i, got[i], wantName)
			}
		}
		for _, name := range got {
			for _, s := range seen {
				if s == name {
					t.Errorf("институция %q встретилась и на первой, и на второй странице", name)
				}
			}
		}
		seen = append(seen, got...)
		page2Cursor = result.NextCursor
	})

	t.Run("в) третья страница — остаток 5 записей, next_cursor пуст", func(t *testing.T) {
		if page2Cursor == nil {
			t.Fatal("page2Cursor не задан — subtest (б) должен выполниться первым")
		}
		result, err := repo.List(ctx, domain.Filter{Statuses: []string{"approved"}, Limit: 10, Cursor: page2Cursor})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(result.Items) != 5 {
			t.Fatalf("len(Items) = %d, want 5", len(result.Items))
		}
		if result.NextCursor != nil {
			t.Errorf("NextCursor = %v, want nil (последняя страница)", *result.NextCursor)
		}
		got := ruNames(result.Items)
		for i, want := 0, 5; i < 5; i, want = i+1, want-1 {
			wantName := fmt.Sprintf("Дефолт-институция %02d", want)
			if got[i] != wantName {
				t.Errorf("Items[%d].Name.RU = %q, want %q", i, got[i], wantName)
			}
		}
	})
}

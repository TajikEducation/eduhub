//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// TestListGeo проверяет гео-фильтр (ST_DWithin), вычисление DistanceM и сортировку
// (price_asc/score/дистанция по умолчанию) метода InstitutionRepo.List.
func TestListGeo(t *testing.T) {
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

	// Фикстуры — 4 институции с реальными координатами (web/lib/data.ts).
	fixtures := []string{
		// 1. Гимназия Сино — Душанбе (кластер А).
		`INSERT INTO catalog.institutions
			(name, types, region, geo, price, rating_avg, moderation_status)
		 VALUES
			('{"ru":"Гимназия Сино","tg":"Гимназияи Сино"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.7760,38.5735)::geography, 800, 4.6, 'approved')`,
		// 2. Лицей Рудаки — Душанбе (кластер Б), rating NULL.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, price, rating_avg, moderation_status)
		 VALUES
			('{"ru":"Лицей Рудаки","tg":"Литсейи Рудакӣ"}', '{school}', 'dushanbe',
			 ST_MakePoint(68.7610,38.5470)::geography, 500, NULL, 'approved')`,
		// 3. Школа Ганджи Худжанд — Sughd.
		`INSERT INTO catalog.institutions
			(name, types, region, geo, price, rating_avg, moderation_status)
		 VALUES
			('{"ru":"Школа Ганджи Худжанд","tg":"Мактаби Ганҷии Хуҷанд"}', '{school}', 'sughd',
			 ST_MakePoint(69.6333,40.2833)::geography, 600, 4.5, 'approved')`,
		// 4. Центр Ояндасоз — Khatlon (Бохтар).
		`INSERT INTO catalog.institutions
			(name, types, region, geo, price, rating_avg, moderation_status)
		 VALUES
			('{"ru":"Центр Ояндасоз","tg":"Маркази Ояндасоз"}', '{school}', 'khatlon',
			 ST_MakePoint(68.7803,37.8322)::geography, 300, 4.2, 'approved')`,
	}
	for _, insertSQL := range fixtures {
		if _, err := tx.Exec(ctx, insertSQL); err != nil {
			t.Fatalf("вставка фикстуры не удалась: %v\nSQL: %s", err, insertSQL)
		}
	}

	repo := postgres.New(tx)

	t.Run("а) RadiusKm=10 возвращает только Душанбе-кластер (1,2)", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Lat:      p(38.56), Lng: p(68.78), RadiusKm: p(10.0),
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{"Гимназия Сино": true, "Лицей Рудаки": true}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("б) RadiusKm=500 возвращает все 4", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Lat:      p(38.56), Lng: p(68.78), RadiusKm: p(500.0),
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		want := map[string]bool{
			"Гимназия Сино": true, "Лицей Рудаки": true,
			"Школа Ганджи Худжанд": true, "Центр Ояндасоз": true,
		}
		if got := nameSet(got); !equalSets(got, want) {
			t.Fatalf("ожидали %v, получили %v", want, got)
		}
	})

	t.Run("в) DistanceM заполнен и монотонно не убывает (сортировка по дистанции по умолчанию)", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Lat:      p(38.56), Lng: p(68.78), RadiusKm: p(500.0),
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(got) != 4 {
			t.Fatalf("ожидали 4 институции, получили %d", len(got))
		}
		for i, inst := range got {
			if inst.DistanceM == nil {
				t.Fatalf("DistanceM[%d] (%s) == nil, ожидали заполненное значение", i, inst.Name.RU)
			}
		}
		for i := 0; i < len(got)-1; i++ {
			if *got[i].DistanceM > *got[i+1].DistanceM {
				t.Fatalf("DistanceM не монотонно не убывает: [%d]=%v > [%d]=%v",
					i, *got[i].DistanceM, i+1, *got[i+1].DistanceM)
			}
		}
	})

	t.Run("г) без Lat/Lng DistanceM == nil у всех", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Region:   p("dushanbe"),
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		if len(got) == 0 {
			t.Fatalf("ожидали хотя бы одну институцию")
		}
		for _, inst := range got {
			if inst.DistanceM != nil {
				t.Fatalf("DistanceM(%s) = %v, ожидали nil (гео не запрашивалось)", inst.Name.RU, *inst.DistanceM)
			}
		}
	})

	t.Run("д1) Sort=price_asc — по возрастанию цены", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Sort:     "price_asc",
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		wantOrder := []string{"Центр Ояндасоз", "Лицей Рудаки", "Школа Ганджи Худжанд", "Гимназия Сино"}
		assertNameOrder(t, got, wantOrder)
	})

	t.Run("д2) Sort=score — по убыванию рейтинга, NULL последним", func(t *testing.T) {
		got, err := repo.List(ctx, domain.Filter{
			Statuses: []string{"approved"},
			Sort:     "score",
		})
		if err != nil {
			t.Fatalf("List() вернул ошибку: %v", err)
		}
		wantOrder := []string{"Гимназия Сино", "Школа Ганджи Худжанд", "Центр Ояндасоз", "Лицей Рудаки"}
		assertNameOrder(t, got, wantOrder)
	})
}

// assertNameOrder проверяет, что RU-названия институций идут в точности в переданном порядке.
func assertNameOrder(t *testing.T, insts []domain.Institution, wantOrder []string) {
	t.Helper()

	if len(insts) != len(wantOrder) {
		t.Fatalf("ожидали %d институций, получили %d", len(wantOrder), len(insts))
	}
	for i, inst := range insts {
		if inst.Name.RU != wantOrder[i] {
			gotOrder := make([]string, len(insts))
			for j, ins := range insts {
				gotOrder[j] = ins.Name.RU
			}
			t.Fatalf("порядок не совпал на позиции %d: ожидали %v, получили %v", i, wantOrder, gotOrder)
		}
	}
}

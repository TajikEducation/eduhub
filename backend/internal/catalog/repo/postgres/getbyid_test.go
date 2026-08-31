//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// queryCountTracer — тестовый pgx.QueryTracer, считающий количество выполненных SQL-запросов.
type queryCountTracer struct {
	count atomic.Int64
}

func (t *queryCountTracer) TraceQueryStart(
	ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData,
) context.Context {
	t.count.Add(1)
	return ctx
}

func (t *queryCountTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// insertFullInstitution вставляет институцию со всеми 6 сателлитными коллекциями и возвращает её id.
// license_no и price намеренно не указаны — остаются SQL NULL (RED-кейс "г").
func insertFullInstitution(ctx context.Context, t *testing.T, tx pgx.Tx) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO catalog.institutions (name, types, region, geo, moderation_status)
		VALUES ('{"ru":"Полная карточка","tg":"Корти пурра"}', '{school}', 'dushanbe',
			ST_MakePoint(68.78,38.56)::geography, 'approved')
		RETURNING id`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("вставка институции не удалась: %v", err)
	}

	fixtures := []struct {
		name string
		sql  string
		args []any
	}{
		{
			name: "staff #1",
			sql: `INSERT INTO catalog.institution_staff (institution_id, name, role_type, role_label)
				VALUES ($1, '{"ru":"Иванова А.","tg":"Иванова А."}', 'teacher', '{"ru":"Учитель","tg":"Муаллим"}')`,
			args: []any{id},
		},
		{
			name: "staff #2",
			sql: `INSERT INTO catalog.institution_staff (institution_id, name, role_type, role_label)
				VALUES ($1, '{"ru":"Петров Б.","tg":"Петров Б."}', 'director', '{"ru":"Директор","tg":"Директор"}')`,
			args: []any{id},
		},
		{
			name: "achievement",
			sql: `INSERT INTO catalog.achievements (owner_type, owner_id, title, year, category, description)
				VALUES ('institution', $1, '{"ru":"Победа в олимпиаде","tg":"Ғалаба дар олимпиада"}', 2024,
					'academic', '{"ru":"Описание","tg":"Тавсиф"}')`,
			args: []any{id},
		},
		{
			name: "gallery #1 (sort_order=1)",
			sql: `INSERT INTO catalog.institution_gallery (institution_id, s3_key, sort_order)
				VALUES ($1, 'gallery/first.jpg', 1)`,
			args: []any{id},
		},
		{
			name: "gallery #2 (sort_order=0)",
			sql: `INSERT INTO catalog.institution_gallery (institution_id, s3_key, sort_order)
				VALUES ($1, 'gallery/second.jpg', 0)`,
			args: []any{id},
		},
		{
			name: "alumnus",
			sql: `INSERT INTO catalog.institution_alumni (institution_id, name, grad_year)
				VALUES ($1, '{"ru":"Выпускник","tg":"Хатмкарда"}', 2020)`,
			args: []any{id},
		},
		{
			name: "transport route",
			sql: `INSERT INTO catalog.institution_transport_routes (institution_id, type, cost_period)
				VALUES ($1, 'own_bus', 'month')`,
			args: []any{id},
		},
		{
			name: "meal plan",
			sql: `INSERT INTO catalog.institution_meal_plans (institution_id, meal_type, cost_period)
				VALUES ($1, 'hot', 'month')`,
			args: []any{id},
		},
	}
	for _, f := range fixtures {
		if _, err := tx.Exec(ctx, f.sql, f.args...); err != nil {
			t.Fatalf("вставка фикстуры %q не удалась: %v", f.name, err)
		}
	}

	return id
}

// insertBareInstitution вставляет институцию без единой сателлитной записи (RED-кейс "в").
func insertBareInstitution(ctx context.Context, t *testing.T, tx pgx.Tx) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO catalog.institutions (name, types, region, geo, moderation_status)
		VALUES ('{"ru":"Без сотрудников","tg":"Бе кормандон"}', '{school}', 'dushanbe',
			ST_MakePoint(68.78,38.56)::geography, 'approved')
		RETURNING id`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("вставка институции не удалась: %v", err)
	}
	return id
}

func TestGetByID(t *testing.T) {
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

	fullID := insertFullInstitution(ctx, t, tx)
	bareID := insertBareInstitution(ctx, t, tx)

	repo := postgres.New(tx)

	t.Run("а) все 6 коллекций заполнены ожидаемым количеством", func(t *testing.T) {
		inst, err := repo.GetByID(ctx, fullID)
		if err != nil {
			t.Fatalf("GetByID() вернул ошибку: %v", err)
		}
		if len(inst.Staff) != 2 {
			t.Errorf("Staff: ожидали 2, получили %d", len(inst.Staff))
		}
		if len(inst.Achievements) != 1 {
			t.Errorf("Achievements: ожидали 1, получили %d", len(inst.Achievements))
		}
		if len(inst.Gallery) != 2 {
			t.Errorf("Gallery: ожидали 2, получили %d", len(inst.Gallery))
		}
		if len(inst.Alumni) != 1 {
			t.Errorf("Alumni: ожидали 1, получили %d", len(inst.Alumni))
		}
		if len(inst.TransportRoutes) != 1 {
			t.Errorf("TransportRoutes: ожидали 1, получили %d", len(inst.TransportRoutes))
		}
		if len(inst.MealPlans) != 1 {
			t.Errorf("MealPlans: ожидали 1, получили %d", len(inst.MealPlans))
		}
	})

	t.Run("б) несуществующий id возвращает apperr.ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("ожидали apperr.ErrNotFound, получили: %v", err)
		}
	})

	t.Run("в) институция без сотрудников — Staff non-nil пустой слайс", func(t *testing.T) {
		inst, err := repo.GetByID(ctx, bareID)
		if err != nil {
			t.Fatalf("GetByID() вернул ошибку: %v", err)
		}
		if inst.Staff == nil {
			t.Error("Staff должен быть non-nil при отсутствии записей")
		}
		if len(inst.Staff) != 0 {
			t.Errorf("Staff: ожидали 0, получили %d", len(inst.Staff))
		}
	})

	t.Run("г) LicenseNo и Price — nil, а не zero-value", func(t *testing.T) {
		inst, err := repo.GetByID(ctx, fullID)
		if err != nil {
			t.Fatalf("GetByID() вернул ошибку: %v", err)
		}
		if inst.LicenseNo != nil {
			t.Errorf("LicenseNo: ожидали nil, получили %q", *inst.LicenseNo)
		}
		if inst.Price != nil {
			t.Errorf("Price: ожидали nil, получили %d", *inst.Price)
		}
	})

	t.Run("д) один SQL-запрос на карточку", func(t *testing.T) {
		cfg, err := pgxpool.ParseConfig(url)
		if err != nil {
			t.Fatalf("ParseConfig() вернул ошибку: %v", err)
		}
		tracer := &queryCountTracer{}
		cfg.ConnConfig.Tracer = tracer

		tracedPool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatalf("NewWithConfig() вернул ошибку: %v", err)
		}
		defer tracedPool.Close()

		tracedTx, err := tracedPool.Begin(ctx)
		if err != nil {
			t.Fatalf("Begin() вернул ошибку: %v", err)
		}
		defer tracedTx.Rollback(ctx)

		id := insertFullInstitution(ctx, t, tracedTx)

		tracer.count.Store(0)

		tracedRepo := postgres.New(tracedTx)
		if _, err := tracedRepo.GetByID(ctx, id); err != nil {
			t.Fatalf("GetByID() вернул ошибку: %v", err)
		}

		if got := tracer.count.Load(); got != 1 {
			t.Errorf("ожидали ровно 1 SQL-запрос, получили %d", got)
		}
	})

	t.Run("е) порядок коллекций стабилен между вызовами", func(t *testing.T) {
		first, err := repo.GetByID(ctx, fullID)
		if err != nil {
			t.Fatalf("GetByID() (1) вернул ошибку: %v", err)
		}
		second, err := repo.GetByID(ctx, fullID)
		if err != nil {
			t.Fatalf("GetByID() (2) вернул ошибку: %v", err)
		}

		if len(first.Gallery) != len(second.Gallery) {
			t.Fatalf("разное количество элементов Gallery между вызовами: %d vs %d",
				len(first.Gallery), len(second.Gallery))
		}
		for i := range first.Gallery {
			if first.Gallery[i].S3Key != second.Gallery[i].S3Key {
				t.Errorf("Gallery[%d]: порядок нестабилен: %q vs %q",
					i, first.Gallery[i].S3Key, second.Gallery[i].S3Key)
			}
		}
		// sort_order=0 должен идти раньше sort_order=1.
		if len(first.Gallery) == 2 && first.Gallery[0].S3Key != "gallery/second.jpg" {
			t.Errorf("Gallery не отсортирован по sort_order: первый элемент %q", first.Gallery[0].S3Key)
		}
	})
}

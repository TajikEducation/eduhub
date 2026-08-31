package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// bilingualJSONOf сериализует domain.Bilingual в JSONB-байты для INSERT/UPDATE (обратная
// операция к scanBilingual в postgres.go).
func bilingualJSONOf(b domain.Bilingual) []byte {
	raw, _ := json.Marshal(bilingualJSON{RU: b.RU, TG: b.TG})
	return raw
}

// bilingualJSONPtrOf — то же для nullable-поля; nil → nil (NULL-параметр), не JSON null.
func bilingualJSONPtrOf(b *domain.Bilingual) []byte {
	if b == nil {
		return nil
	}
	return bilingualJSONOf(*b)
}

// Create вставляет новую институцию и запись владельца. Не обёрнуто в одну транзакцию
// (см. SRS E3.3 — «намеренная простая версия», отсутствие атомарности здесь не создаёт
// проблем безопасности: осиротевшая институция без строки владельца — редкий и
// восстановимый вручную случай, не денежная/RBAC-брешь).
func (r *InstitutionRepo) Create(ctx context.Context, inst domain.Institution, ownerID uuid.UUID) (domain.Institution, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.institutions (
			id, name, types, region, city, district, description,
			geo, phone, email, website, price,
			discount_available, verified, moderation_status, plan, review_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			ST_SetSRID(ST_MakePoint($9, $8), 4326)::geography, $10, $11, $12, $13,
			false, false, $14, $15, 0
		)
		RETURNING id, created_at, updated_at
	`,
		inst.ID, bilingualJSONOf(inst.Name), inst.Types, inst.Region, bilingualJSONPtrOf(inst.City), inst.District, bilingualJSONPtrOf(inst.Description),
		inst.Lat, inst.Lng, inst.Phone, inst.Email, inst.Website, inst.Price,
		inst.ModerationStatus, inst.Plan,
	)

	if err := row.Scan(&inst.ID, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
		return domain.Institution{}, fmt.Errorf("postgres: insert institution: %w", err)
	}

	if _, err := r.db.Exec(ctx, `
		INSERT INTO catalog.institution_owners (institution_id, user_id) VALUES ($1, $2)
	`, inst.ID, ownerID); err != nil {
		return domain.Institution{}, fmt.Errorf("postgres: insert institution owner: %w", err)
	}

	return inst, nil
}

// Update частично обновляет скалярные поля профиля институции (E3.4, урезанная версия —
// см. domain.UpdateInstitutionInput). Каждый nil-указатель оставляет колонку нетронутой
// через COALESCE — не различает "не передано" от явного null (полная семантика отложена).
func (r *InstitutionRepo) Update(ctx context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE catalog.institutions SET
			description = COALESCE($2, description),
			phone = COALESCE($3, phone),
			email = COALESCE($4, email),
			website = COALESCE($5, website),
			cover_photo_s3_key = COALESCE($6, cover_photo_s3_key),
			price = COALESCE($7, price),
			age_range = COALESCE($8, age_range),
			updated_at = now()
		WHERE id = $1
	`, id, bilingualJSONPtrOf(patch.Description), patch.Phone, patch.Email, patch.Website, patch.CoverPhotoS3Key, patch.Price, patch.AgeRange)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("postgres: update institution: %w", err)
	}

	updated, err := r.GetByID(ctx, id)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("postgres: reload institution after update: %w", err)
	}
	return updated, nil
}

// ListByOwner возвращает все институции (любого moderation_status), которыми владеет userID —
// для кабинета учреждения (GET /api/v1/institutions/mine), в отличие от Service.List/Get,
// которые форсируют approved-only для публичного каталога.
func (r *InstitutionRepo) ListByOwner(ctx context.Context, userID uuid.UUID) ([]domain.Institution, error) {
	// IN (subquery), не JOIN — listColumns содержит неквалифицированные created_at/updated_at,
	// которые стали бы неоднозначными при join с institution_owners (тоже имеет created_at).
	rows, err := r.db.Query(ctx, `
		SELECT `+listColumns+`, NULL::float8
		FROM catalog.institutions i
		WHERE i.id IN (SELECT institution_id FROM catalog.institution_owners WHERE user_id = $1)
		ORDER BY i.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list institutions by owner: %w", err)
	}
	defer rows.Close()

	var out []domain.Institution
	for rows.Next() {
		inst, err := scanInstitution(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan institution: %w", err)
		}
		out = append(out, *inst)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list institutions by owner rows: %w", err)
	}
	return out, nil
}

// Exists проверяет, что институция с id существует (для валидации в других модулях,
// например internal/reviews, которые ссылаются на institution_id по значению без FK).
func (r *InstitutionRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM catalog.institutions WHERE id = $1)`, id)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: check institution exists: %w", err)
	}
	return exists, nil
}

// UpdateRatingAvg перезаписывает агрегат рейтинга институции (E4, упрощённая версия —
// синхронный пересчёт по одной общей оценке вместо порта RatingSync с 8 decay-метриками,
// см. миграцию 00008). Вызывается internal/moderation при approve отзыва.
func (r *InstitutionRepo) UpdateRatingAvg(ctx context.Context, id uuid.UUID, avg float64, count int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE catalog.institutions SET rating_avg = $2, review_count = $3, updated_at = now() WHERE id = $1
	`, id, avg, count)
	if err != nil {
		return fmt.Errorf("postgres: update rating avg: %w", err)
	}
	return nil
}

// GetOwnerID возвращает user_id первого владельца институции. apperr.ErrNotFound — нет строки
// владельца (не должно случаться для институций, созданных через Create, но входные данные
// от вызывающего кода транспорта не доверяются).
func (r *InstitutionRepo) GetOwnerID(ctx context.Context, institutionID uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `
		SELECT user_id FROM catalog.institution_owners WHERE institution_id = $1 ORDER BY created_at LIMIT 1
	`, institutionID)

	var ownerID uuid.UUID
	if err := row.Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("institution_owner", institutionID.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get owner id: %w", err)
	}
	return ownerID, nil
}

// SetModerationStatus меняет статус модерации институции (E3.5).
func (r *InstitutionRepo) SetModerationStatus(ctx context.Context, id uuid.UUID, status string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE catalog.institutions SET moderation_status = $2, updated_at = now() WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("postgres: set moderation status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("institution", id.String())
	}
	return nil
}

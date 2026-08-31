package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// CreateAlumnus вставляет запись выпускника.
func (r *InstitutionRepo) CreateAlumnus(ctx context.Context, institutionID uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.institution_alumni (institution_id, name, photo_url, grad_year, now_label)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, institutionID, bilingualJSONOf(in.Name), in.PhotoURL, in.GradYear, bilingualJSONPtrOf(in.NowLabel))

	a := domain.Alumnus{Name: in.Name, PhotoURL: in.PhotoURL, GradYear: in.GradYear, NowLabel: in.NowLabel}
	if err := row.Scan(&a.ID, &a.CreatedAt); err != nil {
		return domain.Alumnus{}, fmt.Errorf("postgres: insert alumnus: %w", err)
	}
	return a, nil
}

// DeleteAlumnus удаляет запись выпускника.
func (r *InstitutionRepo) DeleteAlumnus(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM catalog.institution_alumni WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete alumnus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("alumnus", id.String())
	}
	return nil
}

// GetAlumnusInstitutionID возвращает institution_id владеющей записи.
func (r *InstitutionRepo) GetAlumnusInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT institution_id FROM catalog.institution_alumni WHERE id = $1`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("alumnus", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get alumnus institution id: %w", err)
	}
	return instID, nil
}

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

// CreateGalleryItem вставляет фото/видео в галерею учреждения.
func (r *InstitutionRepo) CreateGalleryItem(ctx context.Context, institutionID uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.institution_gallery (institution_id, s3_key, label, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, institutionID, in.S3Key, bilingualJSONPtrOf(in.Label), in.SortOrder)

	g := domain.GalleryItem{S3Key: in.S3Key, Label: in.Label, SortOrder: in.SortOrder}
	if err := row.Scan(&g.ID, &g.CreatedAt); err != nil {
		return domain.GalleryItem{}, fmt.Errorf("postgres: insert gallery item: %w", err)
	}
	return g, nil
}

// DeleteGalleryItem удаляет фото/видео из галереи.
func (r *InstitutionRepo) DeleteGalleryItem(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM catalog.institution_gallery WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete gallery item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("gallery_item", id.String())
	}
	return nil
}

// GetGalleryItemInstitutionID возвращает institution_id владеющей записи.
func (r *InstitutionRepo) GetGalleryItemInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT institution_id FROM catalog.institution_gallery WHERE id = $1`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("gallery_item", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get gallery item institution id: %w", err)
	}
	return instID, nil
}

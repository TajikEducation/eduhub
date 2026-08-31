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

type achievementLinkJSON struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

func achievementLinksJSONOf(links []domain.AchievementLink) []byte {
	wire := make([]achievementLinkJSON, len(links))
	for i, l := range links {
		wire[i] = achievementLinkJSON{Label: l.Label, URL: l.URL}
	}
	raw, _ := json.Marshal(wire)
	return raw
}

// CreateAchievement вставляет достижение учреждения (owner_type='institution').
func (r *InstitutionRepo) CreateAchievement(ctx context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.achievements (owner_type, owner_id, title, year, category, description, links)
		VALUES ('institution', $1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, institutionID, bilingualJSONOf(in.Title), in.Year, in.Category, bilingualJSONOf(in.Description), achievementLinksJSONOf(in.Links))

	a := domain.Achievement{OwnerType: "institution", OwnerID: institutionID, Title: in.Title, Year: in.Year, Category: in.Category, Description: in.Description, Links: in.Links}
	if err := row.Scan(&a.ID, &a.CreatedAt); err != nil {
		return domain.Achievement{}, fmt.Errorf("postgres: insert achievement: %w", err)
	}
	return a, nil
}

// DeleteAchievement удаляет достижение.
func (r *InstitutionRepo) DeleteAchievement(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM catalog.achievements WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete achievement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("achievement", id.String())
	}
	return nil
}

// GetAchievementInstitutionID возвращает institution_id владеющей записи
// (только owner_type='institution' — достижения сотрудников/учеников вне этой формы).
func (r *InstitutionRepo) GetAchievementInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT owner_id FROM catalog.achievements WHERE id = $1 AND owner_type = 'institution'`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("achievement", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get achievement institution id: %w", err)
	}
	return instID, nil
}

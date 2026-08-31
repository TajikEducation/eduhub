package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// RefreshTokenRepo — репозиторий refresh-токенов поверх PostgreSQL.
type RefreshTokenRepo struct {
	db querier
}

// NewRefreshTokenRepo создаёт RefreshTokenRepo поверх переданного querier.
func NewRefreshTokenRepo(db querier) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Store сохраняет хеш нового refresh-токена.
func (r *RefreshTokenRepo) Store(ctx context.Context, t usecase.RefreshTokenRecord) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.refresh_tokens (id, user_id, token_hash, family_id, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, t.ID, t.UserID, t.TokenHash, t.FamilyID, t.ExpiresAt)
	if err != nil {
		return fmt.Errorf("postgres: store refresh token: %w", err)
	}
	return nil
}

// GetByHash находит запись токена по хешу. Не найден → apperr.Unauthorized (клиент прислал
// токен, которого нет в БД — с точки зрения API это неотличимо от «истёк/отозван»).
func (r *RefreshTokenRepo) GetByHash(ctx context.Context, hash string) (usecase.RefreshTokenRecord, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, replaced_by
		FROM auth.refresh_tokens
		WHERE token_hash = $1
	`, hash)

	var rec usecase.RefreshTokenRecord
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.TokenHash, &rec.FamilyID, &rec.ExpiresAt, &rec.RevokedAt, &rec.ReplacedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usecase.RefreshTokenRecord{}, apperr.Unauthorized("refresh token not found")
		}
		return usecase.RefreshTokenRecord{}, fmt.Errorf("postgres: get refresh token by hash: %w", err)
	}
	return rec, nil
}

// Revoke помечает токен использованным при ротации.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID, replacedBy uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.refresh_tokens SET revoked_at = now(), replaced_by = $2
		WHERE id = $1
	`, id, replacedBy)
	if err != nil {
		return fmt.Errorf("postgres: revoke refresh token: %w", err)
	}
	return nil
}

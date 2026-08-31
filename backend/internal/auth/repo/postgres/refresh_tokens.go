// Package postgres — реализация internal/auth/usecase.RefreshTokenRepo/UserRoleLookup
// поверх PostgreSQL (E2.3, веха 2).
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// querier — минимальный интерфейс доступа к БД (пул или транзакция) — тот же паттерн, что
// internal/catalog/repo/postgres, плюс Exec (этот репозиторий пишет, не только читает): тесты
// вставляют фикстуры в транзакцию, репозиторий должен видеть их в той же сессии, не через
// отдельное подключение из пула.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// RefreshTokenRepo — репозиторий refresh-токенов поверх PostgreSQL.
type RefreshTokenRepo struct {
	db querier
}

// NewRefreshTokenRepo создаёт RefreshTokenRepo поверх переданного querier (пул или транзакция).
func NewRefreshTokenRepo(db querier) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Create вставляет новую строку refresh-токена.
func (r *RefreshTokenRepo) Create(ctx context.Context, rt domain.RefreshToken) error {
	const q = `
		INSERT INTO auth.refresh_tokens (id, user_id, token_hash, family_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := r.db.Exec(ctx, q, rt.ID, rt.UserID, rt.TokenHash, rt.FamilyID, rt.ExpiresAt, rt.CreatedAt); err != nil {
		return fmt.Errorf("postgres: create refresh token: %w", err)
	}
	return nil
}

// FindByHash ищет токен по хешу. Не найден → apperr.NotFound.
func (r *RefreshTokenRepo) FindByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, replaced_by, created_at
		FROM auth.refresh_tokens
		WHERE token_hash = $1
	`
	var rt domain.RefreshToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.FamilyID, &rt.ExpiresAt, &rt.RevokedAt, &rt.ReplacedBy, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, apperr.NotFound("refresh_token", tokenHash)
		}
		return domain.RefreshToken{}, fmt.Errorf("postgres: find refresh token by hash: %w", err)
	}
	return rt, nil
}

// Revoke отзывает один токен, опционально указывая, чем он заменён при ротации.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time, replacedBy *uuid.UUID) error {
	const q = `UPDATE auth.refresh_tokens SET revoked_at = $2, replaced_by = $3 WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, id, revokedAt, replacedBy); err != nil {
		return fmt.Errorf("postgres: revoke refresh token: %w", err)
	}
	return nil
}

// RevokeFamily отзывает разом все ЕЩЁ НЕ отозванные токены семьи (reuse-detection) — уже
// отозванные ранее сохраняют свою исходную метку revoked_at, не перезаписываются.
func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error {
	const q = `UPDATE auth.refresh_tokens SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL`
	if _, err := r.db.Exec(ctx, q, familyID, revokedAt); err != nil {
		return fmt.Errorf("postgres: revoke refresh token family: %w", err)
	}
	return nil
}

// RevokeAllForUser отзывает разом ВСЕ ещё не отозванные refresh-токены пользователя (все
// семьи/устройства) — password-reset и удаление аккаунта.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	const q = `UPDATE auth.refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err := r.db.Exec(ctx, q, userID, revokedAt); err != nil {
		return fmt.Errorf("postgres: revoke all refresh tokens for user: %w", err)
	}
	return nil
}

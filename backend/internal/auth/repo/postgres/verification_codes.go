package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// VerificationCodeRepo — репозиторий кодов подтверждения (email-верификация, password-reset)
// поверх PostgreSQL.
type VerificationCodeRepo struct{ db querier }

// NewVerificationCodeRepo создаёт VerificationCodeRepo поверх переданного querier (пул или транзакция).
func NewVerificationCodeRepo(db querier) *VerificationCodeRepo { return &VerificationCodeRepo{db: db} }

// Create вставляет новую строку кода подтверждения.
func (r *VerificationCodeRepo) Create(ctx context.Context, vc domain.VerificationCode) error {
	const q = `
		INSERT INTO auth.verification_codes (id, user_id, channel, purpose, code_hash, attempts_count, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, q, vc.ID, vc.UserID, vc.Channel, vc.Purpose, vc.CodeHash, vc.AttemptsCount, vc.ExpiresAt, vc.CreatedAt)
	if err != nil {
		return fmt.Errorf("postgres: create verification code: %w", err)
	}
	return nil
}

// FindLatestActive — самый свежий ещё не истёкший код для (userID, channel, purpose).
// apperr.NotFound, если такого нет (не было запроса кода, либо все истекли).
func (r *VerificationCodeRepo) FindLatestActive(ctx context.Context, userID uuid.UUID, channel, purpose string, now time.Time) (domain.VerificationCode, error) {
	const q = `
		SELECT id, user_id, channel, purpose, code_hash, attempts_count, expires_at, created_at
		FROM auth.verification_codes
		WHERE user_id = $1 AND channel = $2 AND purpose = $3 AND expires_at > $4
		ORDER BY created_at DESC
		LIMIT 1
	`
	var vc domain.VerificationCode
	err := r.db.QueryRow(ctx, q, userID, channel, purpose, now).Scan(
		&vc.ID, &vc.UserID, &vc.Channel, &vc.Purpose, &vc.CodeHash, &vc.AttemptsCount, &vc.ExpiresAt, &vc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VerificationCode{}, apperr.NotFound("verification_code", userID.String())
		}
		return domain.VerificationCode{}, fmt.Errorf("postgres: find verification code: %w", err)
	}
	return vc, nil
}

// IncrementAttempts увеличивает счётчик неверных попыток на 1.
func (r *VerificationCodeRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE auth.verification_codes SET attempts_count = attempts_count + 1 WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("postgres: increment verification code attempts: %w", err)
	}
	return nil
}

// Delete удаляет код целиком (использован/просрочен/исчерпал попытки).
func (r *VerificationCodeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM auth.verification_codes WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("postgres: delete verification code: %w", err)
	}
	return nil
}

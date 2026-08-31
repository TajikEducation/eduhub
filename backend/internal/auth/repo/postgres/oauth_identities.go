package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// OAuthRepo — репозиторий связок пользователь↔внешний провайдер поверх PostgreSQL (auth.oauth_identities).
type OAuthRepo struct {
	db querier
}

// NewOAuthRepo создаёт OAuthRepo поверх переданного querier (пул или транзакция).
func NewOAuthRepo(db querier) *OAuthRepo {
	return &OAuthRepo{db: db}
}

// FindByProvider ищет связку по паре (provider, providerUserID). Не найдена → apperr.NotFound.
func (r *OAuthRepo) FindByProvider(ctx context.Context, provider, providerUserID string) (domain.OAuthIdentity, error) {
	const q = `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM auth.oauth_identities
		WHERE provider = $1 AND provider_user_id = $2
	`
	var oi domain.OAuthIdentity
	err := r.db.QueryRow(ctx, q, provider, providerUserID).Scan(
		&oi.ID, &oi.UserID, &oi.Provider, &oi.ProviderUserID, &oi.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OAuthIdentity{}, apperr.NotFound("oauth_identity", provider+":"+providerUserID)
		}
		return domain.OAuthIdentity{}, fmt.Errorf("postgres: find oauth identity by provider: %w", err)
	}
	return oi, nil
}

// Create вставляет новую связку пользователь↔провайдер. Конфликт UNIQUE(provider,
// provider_user_id) маловероятен в штатном потоке (проверяется до вызова usecase-слоем),
// но на всякий случай тоже apperr.ConflictCode, не голая ошибка БД.
func (r *OAuthRepo) Create(ctx context.Context, oi domain.OAuthIdentity) error {
	const q = `
		INSERT INTO auth.oauth_identities (user_id, provider, provider_user_id)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.Exec(ctx, q, oi.UserID, oi.Provider, oi.ProviderUserID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return apperr.ConflictCode("oauth_identity_taken", "эта связка провайдер+id уже привязана к другому пользователю")
		}
		return fmt.Errorf("postgres: create oauth identity: %w", err)
	}
	return nil
}

// Package postgres — реализация internal/moderation/usecase.Recorder поверх PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/moderation/domain"
)

// querier — минимальный интерфейс доступа к БД (пул или транзакция).
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// AuditLogRepo — репозиторий moderation.audit_log поверх PostgreSQL.
type AuditLogRepo struct {
	db querier
}

// New создаёт AuditLogRepo поверх переданного querier.
func New(db querier) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

// Record вставляет запись в moderation.audit_log.
func (r *AuditLogRepo) Record(ctx context.Context, e domain.Entry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO moderation.audit_log (
			actor_type, actor_id, actor_role, action, target_type, target_id,
			reason_code, reason_text, request_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, e.ActorType, e.ActorID, e.ActorRole, e.Action, e.TargetType, e.TargetID, e.ReasonCode, e.ReasonText, e.RequestID)
	if err != nil {
		return fmt.Errorf("postgres: insert audit log entry: %w", err)
	}
	return nil
}

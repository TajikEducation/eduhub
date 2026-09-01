package moderation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Recorder пишет записи в moderation.audit_log. Единственный способ создать запись журнала —
// физически невозможно вставить строку вне транзакции самого изменения, т.к. Record принимает
// pgx.Tx, а не хранит querier полем (см. docs/EduHub_Database_Schema.md, moderation.audit_log).
type Recorder struct{}

// NewRecorder создаёт Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Record вставляет одну запись журнала модерации внутри переданной транзакции tx.
func (r *Recorder) Record(ctx context.Context, tx pgx.Tx, e Entry) error {
	const q = `
		INSERT INTO moderation.audit_log
			(actor_id, actor_type, actor_role, action, target_type, target_id,
			 reason_code, reason_text, payload_diff, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if _, err := tx.Exec(ctx, q,
		e.ActorID, e.ActorType, e.ActorRole, e.Action, e.TargetType, e.TargetID,
		e.ReasonCode, e.ReasonText, e.PayloadDiff, e.RequestID,
	); err != nil {
		return fmt.Errorf("moderation: record audit log entry: %w", err)
	}
	return nil
}

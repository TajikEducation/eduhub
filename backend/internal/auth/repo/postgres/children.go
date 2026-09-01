package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/moderation"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// txQuerier — querier + Begin, нужен ChildRepo для атомарных Confirm/Reject (SELECT FOR UPDATE →
// UPDATE → запись в moderation.audit_log в одной транзакции). Локален для этого файла — не
// расширяет общий querier в refresh_tokens.go, которому Begin не нужен. И *pgxpool.Pool, и
// pgx.Tx уже структурно удовлетворяют txQuerier: вложенный Begin на pgx.Tx создаёт savepoint,
// что совместимо с паттерном интеграционных тестов (репозиторий поверх внешней тестовой
// транзакции).
type txQuerier interface {
	querier
	Begin(ctx context.Context) (pgx.Tx, error)
}

// ChildRepo — реализация usecase.ChildRepo поверх PostgreSQL (E2.6).
type ChildRepo struct {
	db       txQuerier
	recorder *moderation.Recorder
}

// NewChildRepo создаёт ChildRepo поверх переданного txQuerier (пул или транзакция) и Recorder
// для записи в moderation.audit_log.
func NewChildRepo(db txQuerier, recorder *moderation.Recorder) *ChildRepo {
	return &ChildRepo{db: db, recorder: recorder}
}

// Create вставляет новую привязку родитель↔учреждение. Конфликт UNIQUE(user_id, institution_id) →
// apperr.ConflictCode("child_link_exists", ...).
func (r *ChildRepo) Create(ctx context.Context, c domain.Child) (domain.Child, error) {
	const q = `
		INSERT INTO auth.children (id, user_id, institution_id, age_group, status, confirmation_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING confirmation_status, created_at
	`
	err := r.db.QueryRow(ctx, q,
		c.ID, c.UserID, c.InstitutionID, c.AgeGroup, c.Status, c.ConfirmationStatus, c.CreatedAt,
	).Scan(&c.ConfirmationStatus, &c.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domain.Child{}, apperr.ConflictCode("child_link_exists", "привязка к этому учреждению уже существует")
		}
		return domain.Child{}, fmt.Errorf("postgres: create child: %w", err)
	}
	return c, nil
}

// ListPendingByInstitution возвращает привязки со confirmation_status='pending' для данного
// учреждения — очередь подтверждения.
func (r *ChildRepo) ListPendingByInstitution(ctx context.Context, institutionID uuid.UUID) ([]domain.Child, error) {
	const q = `
		SELECT id, user_id, institution_id, age_group, status, confirmation_status, confirmed_by, confirmed_at, created_at
		FROM auth.children
		WHERE institution_id = $1 AND confirmation_status = 'pending'
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, q, institutionID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list pending children by institution: %w", err)
	}
	defer rows.Close()

	var children []domain.Child
	for rows.Next() {
		var c domain.Child
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.InstitutionID, &c.AgeGroup, &c.Status, &c.ConfirmationStatus, &c.ConfirmedBy, &c.ConfirmedAt, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres: scan pending child: %w", err)
		}
		children = append(children, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list pending children by institution: %w", err)
	}
	return children, nil
}

// Confirm атомарно переводит confirmation_status: pending→confirmed и пишет audit-запись.
func (r *ChildRepo) Confirm(ctx context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error) {
	return r.transitionConfirmation(ctx, childID, actorID, actorRole, "confirmed", "child_confirmed", nil, nil, requestID)
}

// Reject атомарно переводит confirmation_status: pending→rejected (с обязательной причиной) и
// пишет audit-запись.
func (r *ChildRepo) Reject(ctx context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error) {
	return r.transitionConfirmation(ctx, childID, actorID, actorRole, "rejected", "child_rejected", &reasonCode, reasonText, requestID)
}

// transitionConfirmation — общая атомарная логика Confirm/Reject: SELECT FOR UPDATE (различает
// «не найдено» от «есть, но не pending») → UPDATE → запись в moderation.audit_log → commit,
// всё в одной транзакции.
func (r *ChildRepo) transitionConfirmation(
	ctx context.Context, childID, actorID uuid.UUID, actorRole, newStatus, action string, reasonCode, reasonText *string, requestID string,
) (domain.Child, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Child{}, fmt.Errorf("postgres: begin transition confirmation tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после успешного Commit Rollback вернёт pgx.ErrTxClosed — ожидаемо, некритично

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT confirmation_status FROM auth.children WHERE id = $1 FOR UPDATE`, childID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Child{}, apperr.NotFound("child", childID.String())
		}
		return domain.Child{}, fmt.Errorf("postgres: select child for update: %w", err)
	}
	if currentStatus != "pending" {
		return domain.Child{}, apperr.ConflictCode("child_not_pending", "привязка уже обработана")
	}

	var c domain.Child
	const updateQ = `
		UPDATE auth.children
		SET confirmation_status = $1, confirmed_by = $2, confirmed_at = now()
		WHERE id = $3
		RETURNING id, user_id, institution_id, age_group, status, confirmation_status, confirmed_by, confirmed_at, created_at
	`
	err = tx.QueryRow(ctx, updateQ, newStatus, actorID, childID).Scan(
		&c.ID, &c.UserID, &c.InstitutionID, &c.AgeGroup, &c.Status, &c.ConfirmationStatus, &c.ConfirmedBy, &c.ConfirmedAt, &c.CreatedAt,
	)
	if err != nil {
		return domain.Child{}, fmt.Errorf("postgres: update child confirmation status: %w", err)
	}

	entry := moderation.Entry{
		ActorID:    &actorID,
		ActorType:  "user",
		ActorRole:  &actorRole,
		Action:     action,
		TargetType: "child",
		TargetID:   childID,
		ReasonCode: reasonCode,
		ReasonText: reasonText,
		RequestID:  requestID,
	}
	if err := r.recorder.Record(ctx, tx, entry); err != nil {
		return domain.Child{}, fmt.Errorf("postgres: record audit log entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Child{}, fmt.Errorf("postgres: commit transition confirmation tx: %w", err)
	}

	return c, nil
}

// Package postgres — pgx-реализация internal/communications/usecase.{ConversationRepo,MessageRepo}.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/communications/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// querier — минимальный интерфейс доступа к БД, тот же паттерн, что internal/catalog и др.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// ConversationRepo — pgx-реализация usecase.ConversationRepo.
type ConversationRepo struct {
	db querier
}

func NewConversationRepo(db querier) *ConversationRepo {
	return &ConversationRepo{db: db}
}

const conversationColumns = `id, participant_a_type, participant_a_id, participant_b_type, participant_b_id, created_at`

func scanConversation(row interface{ Scan(dest ...any) error }) (domain.Conversation, error) {
	var c domain.Conversation
	if err := row.Scan(&c.ID, &c.ParticipantAType, &c.ParticipantAID, &c.ParticipantBType, &c.ParticipantBID, &c.CreatedAt); err != nil {
		return domain.Conversation{}, err
	}
	return c, nil
}

// GetOrCreate — a/b уже канонизированы вызывающим usecase-слоем (Service.canonicalize), так что
// UNIQUE(participant_a_*, participant_b_*) детерминированно ловит существующую строку.
func (r *ConversationRepo) GetOrCreate(ctx context.Context, a, b domain.Participant) (domain.Conversation, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.conversations (participant_a_type, participant_a_id, participant_b_type, participant_b_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (participant_a_type, participant_a_id, participant_b_type, participant_b_id) DO UPDATE SET participant_a_type = EXCLUDED.participant_a_type
		RETURNING `+conversationColumns, a.Type, a.ID, b.Type, b.ID)
	c, err := scanConversation(row)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: get or create conversation: %w", err)
	}
	return c, nil
}

// ListForParticipant возвращает диалоги, где p — любая из двух сторон, самые новые первыми.
func (r *ConversationRepo) ListForParticipant(ctx context.Context, p domain.Participant) ([]domain.Conversation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+conversationColumns+` FROM communications.conversations
		WHERE (participant_a_type = $1 AND participant_a_id = $2) OR (participant_b_type = $1 AND participant_b_id = $2)
		ORDER BY created_at DESC
	`, p.Type, p.ID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list conversations: %w", err)
	}
	defer rows.Close()

	var out []domain.Conversation
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan conversation: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list conversations rows: %w", err)
	}
	return out, nil
}

// GetByID возвращает диалог по id.
func (r *ConversationRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Conversation, error) {
	row := r.db.QueryRow(ctx, `SELECT `+conversationColumns+` FROM communications.conversations WHERE id = $1`, id)
	c, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Conversation{}, apperr.NotFound("conversation", id.String())
		}
		return domain.Conversation{}, fmt.Errorf("postgres: get conversation: %w", err)
	}
	return c, nil
}

// MessageRepo — pgx-реализация usecase.MessageRepo.
type MessageRepo struct {
	db querier
}

func NewMessageRepo(db querier) *MessageRepo {
	return &MessageRepo{db: db}
}

// Create вставляет сообщение.
func (r *MessageRepo) Create(ctx context.Context, conversationID uuid.UUID, sender domain.Participant, body string) (domain.Message, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.messages (conversation_id, sender_type, sender_id, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, conversationID, sender.Type, sender.ID, body)

	m := domain.Message{ConversationID: conversationID, SenderType: sender.Type, SenderID: sender.ID, Body: body}
	if err := row.Scan(&m.ID, &m.CreatedAt); err != nil {
		return domain.Message{}, fmt.Errorf("postgres: insert message: %w", err)
	}
	return m, nil
}

// ListByConversation возвращает сообщения диалога в хронологическом порядке.
func (r *MessageRepo) ListByConversation(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, conversation_id, sender_type, sender_id, body, created_at
		FROM communications.messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list messages: %w", err)
	}
	defer rows.Close()

	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderID, &m.Body, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan message: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list messages rows: %w", err)
	}
	return out, nil
}

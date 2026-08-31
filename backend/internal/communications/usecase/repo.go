// Package usecase — бизнес-логика чата: получить-или-создать диалог, список диалогов
// участника, отправка/чтение сообщений.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/communications/domain"
)

// ConversationRepo — порт в БД для диалогов. Реализация — internal/communications/repo/postgres.
type ConversationRepo interface {
	// GetOrCreate возвращает существующий диалог между a и b (в любом порядке — уникальность
	// в БД канонизирована usecase-слоем, см. Service.canonicalize) или создаёт новый.
	GetOrCreate(ctx context.Context, a, b domain.Participant) (domain.Conversation, error)
	ListForParticipant(ctx context.Context, p domain.Participant) ([]domain.Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Conversation, error)
}

// MessageRepo — порт в БД для сообщений.
type MessageRepo interface {
	Create(ctx context.Context, conversationID uuid.UUID, sender domain.Participant, body string) (domain.Message, error)
	ListByConversation(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error)
}

// InstitutionChecker — порт в catalog: существование институции и проверка владельца (тот же
// паттерн, что у internal/vacancies и internal/reviews) — нужен, чтобы разрешить владельцу
// институции писать/читать чат «от лица учреждения» (as_institution_id).
type InstitutionChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
}

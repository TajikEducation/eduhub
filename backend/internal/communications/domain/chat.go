// Package domain — доменная модель чата (веха 5, ядро): упрощённая версия SRS-спека
// (docs/EduHub_Database_Schema.md, communications.conversations/messages) — REST-поллинг на
// фронте вместо WebSocket/Redis Pub/Sub (план E5.2), без push-уведомлений, без пометки
// прочтения (participant_*_last_read_at в схеме есть, но usecase их пока не использует).
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Тип участника диалога — полиморфный слот: пользователь или представитель институции.
const (
	ParticipantUser        = "user"
	ParticipantInstitution = "institution"
)

// Participant — одна из сторон диалога/отправитель сообщения.
type Participant struct {
	Type string
	ID   uuid.UUID
}

// Conversation — диалог между двумя участниками (FR-19).
type Conversation struct {
	ID               uuid.UUID
	ParticipantAType string
	ParticipantAID   uuid.UUID
	ParticipantBType string
	ParticipantBID   uuid.UUID
	CreatedAt        time.Time
}

// Message — реплика внутри диалога.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderType     string
	SenderID       uuid.UUID
	Body           string
	CreatedAt      time.Time
}

// SendMessageInput — данные для отправки сообщения.
type SendMessageInput struct {
	Body string
}

func (in SendMessageInput) Validate() error {
	if strings.TrimSpace(in.Body) == "" {
		return apperr.Invalid(map[string]string{"body": "обязательное поле"}, "пустое сообщение")
	}
	return nil
}

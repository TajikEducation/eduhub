package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/communications/domain"
)

// conversationDTO — публичный контракт диалога. Имя/фото собеседника фронт добирает сам через
// уже существующие эндпоинты (GET /api/v1/institutions/{id} для counterpart_type=institution) —
// коммуникационный модуль намеренно не знает о catalog/auth домене (владение схемами), тот же
// принцип, что у reviews (не хранит имя автора отзыва).
type conversationDTO struct {
	ID              uuid.UUID `json:"id"`
	CounterpartType string    `json:"counterpart_type"`
	CounterpartID   uuid.UUID `json:"counterpart_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// toConversationDTO вычисляет «собеседника» относительно me — участника, который не является me.
func toConversationDTO(c domain.Conversation, me domain.Participant) conversationDTO {
	if c.ParticipantAType == me.Type && c.ParticipantAID == me.ID {
		return conversationDTO{ID: c.ID, CounterpartType: c.ParticipantBType, CounterpartID: c.ParticipantBID, CreatedAt: c.CreatedAt}
	}
	return conversationDTO{ID: c.ID, CounterpartType: c.ParticipantAType, CounterpartID: c.ParticipantAID, CreatedAt: c.CreatedAt}
}

type messageDTO struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderType     string    `json:"sender_type"`
	SenderID       uuid.UUID `json:"sender_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

func toMessageDTO(m domain.Message) messageDTO {
	return messageDTO{ID: m.ID, ConversationID: m.ConversationID, SenderType: m.SenderType, SenderID: m.SenderID, Body: m.Body, CreatedAt: m.CreatedAt}
}

type listConversationsResponse struct {
	Items []conversationDTO `json:"items"`
}

type listMessagesResponse struct {
	Items []messageDTO `json:"items"`
}

// createConversationRequest — тело POST /api/v1/conversations.
type createConversationRequest struct {
	CounterpartType string     `json:"counterpart_type"`
	CounterpartID   uuid.UUID  `json:"counterpart_id"`
	AsInstitutionID *uuid.UUID `json:"as_institution_id,omitempty"`
}

// sendMessageRequest — тело POST /api/v1/conversations/{id}/messages.
type sendMessageRequest struct {
	Body            string     `json:"body"`
	AsInstitutionID *uuid.UUID `json:"as_institution_id,omitempty"`
}

package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/communications/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Service — usecase-слой чата.
type Service struct {
	conversations ConversationRepo
	messages      MessageRepo
	institutions  InstitutionChecker
}

// New создаёт Service.
func New(conversations ConversationRepo, messages MessageRepo, institutions InstitutionChecker) *Service {
	return &Service{conversations: conversations, messages: messages, institutions: institutions}
}

// canonicalize сортирует пару участников (по type, затем по id), чтобы диалог A↔B и B↔A всегда
// попадали в одну и ту же строку (см. UNIQUE в 00010_communications_chat.sql) — та же
// канонизация, что описана в docs/EduHub_Database_Schema.md для conversations.
func canonicalize(a, b domain.Participant) (domain.Participant, domain.Participant) {
	if a.Type > b.Type || (a.Type == b.Type && a.ID.String() > b.ID.String()) {
		return b, a
	}
	return a, b
}

// ResolveActor определяет, от чьего лица действует principal: как он сам (user) — по умолчанию,
// или как представитель институции institutionID (asInstitutionID != nil) — только если он её
// владелец. Общая точка входа для всех хендлеров транспорта (см. пакетный комментарий handler.go).
func (s *Service) ResolveActor(ctx context.Context, userID uuid.UUID, asInstitutionID *uuid.UUID) (domain.Participant, error) {
	if asInstitutionID == nil {
		return domain.Participant{Type: domain.ParticipantUser, ID: userID}, nil
	}
	owner, err := s.institutions.IsOwner(ctx, *asInstitutionID, userID)
	if err != nil {
		return domain.Participant{}, fmt.Errorf("usecase: check institution owner: %w", err)
	}
	if !owner {
		return domain.Participant{}, apperr.Forbidden("вы не владелец этой институции")
	}
	return domain.Participant{Type: domain.ParticipantInstitution, ID: *asInstitutionID}, nil
}

// GetOrCreateConversation возвращает (создавая при необходимости) диалог между me и counterpart.
func (s *Service) GetOrCreateConversation(ctx context.Context, me, counterpart domain.Participant) (domain.Conversation, error) {
	if counterpart.Type == domain.ParticipantInstitution {
		exists, err := s.institutions.Exists(ctx, counterpart.ID)
		if err != nil {
			return domain.Conversation{}, fmt.Errorf("usecase: check counterpart institution exists: %w", err)
		}
		if !exists {
			return domain.Conversation{}, apperr.NotFound("institution", counterpart.ID.String())
		}
	}
	pa, pb := canonicalize(me, counterpart)
	conv, err := s.conversations.GetOrCreate(ctx, pa, pb)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("usecase: get or create conversation: %w", err)
	}
	return conv, nil
}

// ListMyConversations возвращает диалоги участника me, самые новые первыми (сортировку по
// последнему сообщению репозиторий не делает — небольшой MVP-каталог, см. пакетный
// комментарий repo/postgres).
func (s *Service) ListMyConversations(ctx context.Context, me domain.Participant) ([]domain.Conversation, error) {
	items, err := s.conversations.ListForParticipant(ctx, me)
	if err != nil {
		return nil, fmt.Errorf("usecase: list conversations: %w", err)
	}
	return items, nil
}

// isParticipant проверяет, является ли p одной из двух сторон диалога.
func isParticipant(conv domain.Conversation, p domain.Participant) bool {
	if conv.ParticipantAType == p.Type && conv.ParticipantAID == p.ID {
		return true
	}
	return conv.ParticipantBType == p.Type && conv.ParticipantBID == p.ID
}

// ListMessages возвращает сообщения диалога — только для участника (иначе Forbidden).
func (s *Service) ListMessages(ctx context.Context, conversationID uuid.UUID, me domain.Participant) ([]domain.Message, error) {
	conv, err := s.conversations.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("usecase: get conversation: %w", err)
	}
	if !isParticipant(conv, me) {
		return nil, apperr.Forbidden("вы не участник этого диалога")
	}
	items, err := s.messages.ListByConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list messages: %w", err)
	}
	return items, nil
}

// SendMessage отправляет сообщение в диалог от лица me — только если me его участник.
func (s *Service) SendMessage(ctx context.Context, conversationID uuid.UUID, me domain.Participant, in domain.SendMessageInput) (domain.Message, error) {
	if err := in.Validate(); err != nil {
		return domain.Message{}, err
	}
	conv, err := s.conversations.GetByID(ctx, conversationID)
	if err != nil {
		return domain.Message{}, fmt.Errorf("usecase: get conversation: %w", err)
	}
	if !isParticipant(conv, me) {
		return domain.Message{}, apperr.Forbidden("вы не участник этого диалога")
	}
	msg, err := s.messages.Create(ctx, conversationID, me, in.Body)
	if err != nil {
		return domain.Message{}, fmt.Errorf("usecase: send message: %w", err)
	}
	return msg, nil
}

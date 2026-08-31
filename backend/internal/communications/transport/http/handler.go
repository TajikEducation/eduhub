// Package http — транспортный слой чата (веха 5, ядро) поверх net/http. Все маршруты — за
// rbac.RequireAuth: анонимный доступ к чату не предусмотрен ни для одной роли.
package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/communications/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// chatService — то, что нужно транспорту от usecase-слоя.
type chatService interface {
	ResolveActor(ctx context.Context, userID uuid.UUID, asInstitutionID *uuid.UUID) (domain.Participant, error)
	GetOrCreateConversation(ctx context.Context, me, counterpart domain.Participant) (domain.Conversation, error)
	ListMyConversations(ctx context.Context, me domain.Participant) ([]domain.Conversation, error)
	ListMessages(ctx context.Context, conversationID uuid.UUID, me domain.Participant) ([]domain.Message, error)
	SendMessage(ctx context.Context, conversationID uuid.UUID, me domain.Participant, in domain.SendMessageInput) (domain.Message, error)
}

func requirePrincipal(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (rbac.Principal, bool) {
	principal, ok := rbac.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
		return rbac.Principal{}, false
	}
	return principal, true
}

// parseAsInstitutionID читает необязательный query-параметр as_institution_id (для GET).
func parseAsInstitutionID(r *http.Request) (*uuid.UUID, error) {
	raw := r.URL.Query().Get("as_institution_id")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperr.Invalid(map[string]string{"as_institution_id": "невалидный UUID"}, "некорректный as_institution_id")
	}
	return &id, nil
}

// isValidParticipantType — допустимые значения полиморфного слота участника.
func isValidParticipantType(t string) bool {
	return t == domain.ParticipantUser || t == domain.ParticipantInstitution
}

// CreateConversationHandler — POST /api/v1/conversations: получить-или-создать диалог с
// counterpart. За rbac.RequireAuth.
func CreateConversationHandler(svc chatService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		var req createConversationRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		if !isValidParticipantType(req.CounterpartType) {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"counterpart_type": "должно быть user или institution"}, "некорректный тип собеседника"))
			return
		}
		me, err := svc.ResolveActor(r.Context(), principal.UserID, req.AsInstitutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		counterpart := domain.Participant{Type: req.CounterpartType, ID: req.CounterpartID}
		conv, err := svc.GetOrCreateConversation(r.Context(), me, counterpart)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toConversationDTO(conv, me))
	}
}

// ListConversationsHandler — GET /api/v1/conversations[?as_institution_id=]. За rbac.RequireAuth.
func ListConversationsHandler(svc chatService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		asInstitutionID, err := parseAsInstitutionID(r)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		me, err := svc.ResolveActor(r.Context(), principal.UserID, asInstitutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		items, err := svc.ListMyConversations(r.Context(), me)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]conversationDTO, len(items))
		for i, c := range items {
			dtos[i] = toConversationDTO(c, me)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listConversationsResponse{Items: dtos})
	}
}

// ListMessagesHandler — GET /api/v1/conversations/{id}/messages[?as_institution_id=]. За
// rbac.RequireAuth.
func ListMessagesHandler(svc chatService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		conversationID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		asInstitutionID, err := parseAsInstitutionID(r)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		me, err := svc.ResolveActor(r.Context(), principal.UserID, asInstitutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		items, err := svc.ListMessages(r.Context(), conversationID, me)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]messageDTO, len(items))
		for i, m := range items {
			dtos[i] = toMessageDTO(m)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listMessagesResponse{Items: dtos})
	}
}

// SendMessageHandler — POST /api/v1/conversations/{id}/messages. За rbac.RequireAuth.
func SendMessageHandler(svc chatService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		conversationID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		var req sendMessageRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		me, err := svc.ResolveActor(r.Context(), principal.UserID, req.AsInstitutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		msg, err := svc.SendMessage(r.Context(), conversationID, me, domain.SendMessageInput{Body: req.Body})
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toMessageDTO(msg))
	}
}

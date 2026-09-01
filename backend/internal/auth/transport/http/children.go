package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// validAgeGroups — допустимые значения age_group (см. CHECK-constraint auth.children в
// 00006_auth_users_and_sessions.sql).
var validAgeGroups = map[string]bool{
	"kindergarten": true,
	"preschool":    true,
	"primary":      true,
	"basic":        true,
	"secondary":    true,
	"university":   true,
}

// validChildStatuses — допустимые значения status (см. CHECK-constraint auth.children).
var validChildStatuses = map[string]bool{
	"current":     true,
	"alumnus":     true,
	"transferred": true,
}

// childService — то, что нужно транспорту от usecase-слоя привязок родитель↔учреждение (E2.6).
type childService interface {
	CreateChild(ctx context.Context, userID, institutionID uuid.UUID, ageGroup, status string) (domain.Child, error)
	ListPendingByInstitution(ctx context.Context, institutionID uuid.UUID) ([]domain.Child, error)
	ConfirmChild(ctx context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error)
	RejectChild(ctx context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error)
}

// toChildResponse маппит domain.Child в публичный DTO.
func toChildResponse(c domain.Child) childResponse {
	return childResponse{
		ID:                 c.ID,
		UserID:             c.UserID,
		InstitutionID:      c.InstitutionID,
		AgeGroup:           c.AgeGroup,
		Status:             c.Status,
		ConfirmationStatus: c.ConfirmationStatus,
		ConfirmedBy:        c.ConfirmedBy,
		ConfirmedAt:        c.ConfirmedAt,
		CreatedAt:          c.CreatedAt,
	}
}

// CreateChildHandler — POST /auth/children, защищён RequireAuth. Владелец привязки — сам
// залогиненный пользователь (Principal.UserID), не поле в теле запроса.
func CreateChildHandler(svc childService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}

		var req createChildRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		fields := map[string]string{}
		if req.InstitutionID == uuid.Nil {
			fields["institution_id"] = "обязателен"
		}
		if !validAgeGroups[req.AgeGroup] {
			fields["age_group"] = "недопустимое значение"
		}
		if !validChildStatuses[req.Status] {
			fields["status"] = "недопустимое значение"
		}
		if len(fields) > 0 {
			httpx.WriteError(w, r, logger, apperr.Invalid(fields, "некорректные данные привязки"))
			return
		}

		c, err := svc.CreateChild(r.Context(), principal.UserID, req.InstitutionID, req.AgeGroup, req.Status)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toChildResponse(c))
	}
}

// ListPendingChildrenHandler — GET /institutions/{id}/children/pending, защищён
// RequireAuth+RequireRole(moderator, admin) на уровне роутинга.
func ListPendingChildrenHandler(svc childService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		children, err := svc.ListPendingByInstitution(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		resp := make([]childResponse, 0, len(children))
		for _, c := range children {
			resp = append(resp, toChildResponse(c))
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

// ConfirmChildHandler — POST /children/{id}/confirm, защищён RequireAuth+RequireRole(moderator,
// admin) на уровне роутинга. Тело не требуется.
func ConfirmChildHandler(svc childService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}

		childID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		c, err := svc.ConfirmChild(r.Context(), childID, principal.UserID, principal.Role, httpx.RequestID(r.Context()))
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toChildResponse(c))
	}
}

// RejectChildHandler — POST /children/{id}/reject, защищён RequireAuth+RequireRole(moderator,
// admin) на уровне роутинга. ReasonCode обязателен.
func RejectChildHandler(svc childService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}

		childID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		var req rejectChildRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if req.ReasonCode == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"reason_code": "обязателен"}, "некорректные данные"))
			return
		}

		c, err := svc.RejectChild(r.Context(), childID, principal.UserID, principal.Role, req.ReasonCode, req.ReasonText, httpx.RequestID(r.Context()))
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toChildResponse(c))
	}
}

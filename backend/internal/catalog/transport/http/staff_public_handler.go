package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// publicStaffService — то, что нужно GetStaffHandler от usecase-слоя.
type publicStaffService interface {
	GetPublicStaff(ctx context.Context, id uuid.UUID) (domain.StaffMember, error)
}

// GetStaffHandler — GET /api/v1/staff/{staffId}: публичный профиль сотрудника (страница
// /people/{id} на фронте). Виден, только если институция сотрудника approved.
func GetStaffHandler(svc publicStaffService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("staffId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"staffId": "невалидный UUID"}, "некорректный id"))
			return
		}
		member, err := svc.GetPublicStaff(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toStaffMemberDTO(member))
	}
}

// moderationListService — то, что нужно ListForModerationHandler от usecase-слоя.
type moderationListService interface {
	ListForModeration(ctx context.Context, status string) ([]domain.Institution, error)
}

// listForModerationResponse — тело ответа GET /api/v1/moderation/institutions.
type listForModerationResponse struct {
	Items []moderationListItemDTO `json:"items"`
}

// ListForModerationHandler — GET /api/v1/moderation/institutions[?status=pending]: очередь
// институций для модератора/админа (любой moderation_status, не только approved). За
// rbac.RequireAuth + rbac.RequireRole("moderator","admin").
func ListForModerationHandler(svc moderationListService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		items, err := svc.ListForModeration(r.Context(), status)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]moderationListItemDTO, len(items))
		for i, inst := range items {
			dtos[i] = toModerationListItemDTO(inst)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listForModerationResponse{Items: dtos})
	}
}

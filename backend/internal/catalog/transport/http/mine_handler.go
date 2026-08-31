package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// mineService — то, что нужно MineHandler/GetMineHandler от usecase-слоя.
type mineService interface {
	ListMine(ctx context.Context, userID uuid.UUID) ([]domain.Institution, error)
	GetMine(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Institution, error)
}

// mineResponse — тело ответа GET /api/v1/institutions/mine.
type mineResponse struct {
	Items []institutionDTO `json:"items"`
}

// MineHandler — GET /api/v1/institutions/mine: институции текущего пользователя (любого
// moderation_status — в отличие от публичного каталога). Должен идти за rbac.RequireAuth.
func MineHandler(svc mineService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}

		insts, err := svc.ListMine(r.Context(), principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		items := make([]institutionDTO, len(insts))
		for i, inst := range insts {
			items[i] = toInstitutionDTO(inst)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, mineResponse{Items: items})
	}
}

// GetMineHandler — GET /api/v1/institutions/{id}/mine: полная карточка (со всеми сателлитами)
// институции для её владельца, независимо от статуса модерации. Должен идти за rbac.RequireAuth.
func GetMineHandler(svc mineService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		inst, err := svc.GetMine(r.Context(), id, principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toInstitutionDTO(inst))
	}
}

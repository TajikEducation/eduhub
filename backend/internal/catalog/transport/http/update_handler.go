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

// updateRequest — тело PATCH /api/v1/institutions/{id} (E3.4, урезанная версия — см.
// domain.UpdateInstitutionInput: nil = не трогать, явный JSON null пока тоже не трогает
// поле, полная семантика отложена).
type updateRequest struct {
	Description     *bilingualDTO `json:"description,omitempty"`
	Phone           *string       `json:"phone,omitempty"`
	Email           *string       `json:"email,omitempty"`
	Website         *string       `json:"website,omitempty"`
	CoverPhotoS3Key *string       `json:"cover_photo_s3_key,omitempty"`
	Price           *int          `json:"price,omitempty"`
	AgeRange        *string       `json:"age_range,omitempty"`
}

func (req updateRequest) toDomain() domain.UpdateInstitutionInput {
	return domain.UpdateInstitutionInput{
		Description:     fromBilingualDTOPtr(req.Description),
		Phone:           req.Phone,
		Email:           req.Email,
		Website:         req.Website,
		CoverPhotoS3Key: req.CoverPhotoS3Key,
		Price:           req.Price,
		AgeRange:        req.AgeRange,
	}
}

// updateService — то, что нужно UpdateHandler от usecase-слоя.
type updateService interface {
	IsOwner(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error)
}

// isPrivilegedRole — модератор/админ может редактировать любую институцию, не только свою.
func isPrivilegedRole(role string) bool {
	return role == "moderator" || role == "admin"
}

// UpdateHandler — PATCH /api/v1/institutions/{id} (E3.4, без ETag/If-Match — см. пакет-комментарий).
// Должен идти за rbac.RequireAuth. Доступ: владелец институции или moderator/admin.
func UpdateHandler(svc updateService, logger *slog.Logger) http.HandlerFunc {
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

		if !isPrivilegedRole(principal.Role) {
			owner, err := svc.IsOwner(r.Context(), id, principal.UserID)
			if err != nil {
				httpx.WriteError(w, r, logger, err)
				return
			}
			if !owner {
				httpx.WriteError(w, r, logger, apperr.Forbidden("вы не владелец этой институции"))
				return
			}
		}

		var req updateRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		updated, err := svc.Update(r.Context(), id, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toInstitutionDTO(updated))
	}
}

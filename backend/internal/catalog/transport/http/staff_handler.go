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

type staffRequest struct {
	Name      bilingualDTO  `json:"name"`
	RoleType  string        `json:"role_type"`
	RoleLabel bilingualDTO  `json:"role_label"`
	Subject   *bilingualDTO `json:"subject,omitempty"`
	PhotoURL  *string       `json:"photo_url,omitempty"`
	Exp       *string       `json:"exp,omitempty"`
	Bio       *bilingualDTO `json:"bio,omitempty"`
	Email     *string       `json:"email,omitempty"`
	Phone     *string       `json:"phone,omitempty"`
}

func (req staffRequest) toDomain() domain.CreateStaffInput {
	return domain.CreateStaffInput{
		Name: domain.Bilingual{RU: req.Name.RU, TG: req.Name.TG}, RoleType: req.RoleType,
		RoleLabel: domain.Bilingual{RU: req.RoleLabel.RU, TG: req.RoleLabel.TG},
		Subject:   fromBilingualDTOPtr(req.Subject), PhotoURL: req.PhotoURL, Exp: req.Exp,
		Bio: fromBilingualDTOPtr(req.Bio), Email: req.Email, Phone: req.Phone,
	}
}

type staffService interface {
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	IsOwnerOfStaff(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	CreateStaff(ctx context.Context, institutionID uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error)
	UpdateStaff(ctx context.Context, id uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error)
	DeleteStaff(ctx context.Context, id uuid.UUID) error
}

// CreateStaffHandler — POST /api/v1/institutions/{id}/staff. За rbac.RequireAuth.
func CreateStaffHandler(svc staffService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwner(ctx, institutionID, principal.UserID)
		}) {
			return
		}

		var req staffRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateStaff(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toStaffMemberDTO(created))
	}
}

// UpdateStaffHandler — PATCH /api/v1/institutions/{id}/staff/{staffId}. За rbac.RequireAuth.
func UpdateStaffHandler(svc staffService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("staffId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"staffId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfStaff(ctx, id, principal.UserID)
		}) {
			return
		}

		var req staffRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		updated, err := svc.UpdateStaff(r.Context(), id, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toStaffMemberDTO(updated))
	}
}

// DeleteStaffHandler — DELETE /api/v1/institutions/{id}/staff/{staffId}. За rbac.RequireAuth.
func DeleteStaffHandler(svc staffService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("staffId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"staffId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfStaff(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.DeleteStaff(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

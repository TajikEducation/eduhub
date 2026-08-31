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

type alumnusRequest struct {
	Name     bilingualDTO  `json:"name"`
	PhotoURL *string       `json:"photo_url,omitempty"`
	GradYear int           `json:"grad_year"`
	NowLabel *bilingualDTO `json:"now_label,omitempty"`
}

func (req alumnusRequest) toDomain() domain.CreateAlumnusInput {
	return domain.CreateAlumnusInput{
		Name: domain.Bilingual{RU: req.Name.RU, TG: req.Name.TG}, PhotoURL: req.PhotoURL,
		GradYear: req.GradYear, NowLabel: fromBilingualDTOPtr(req.NowLabel),
	}
}

type alumnusService interface {
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	IsOwnerOfAlumnus(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	CreateAlumnus(ctx context.Context, institutionID uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error)
	DeleteAlumnus(ctx context.Context, id uuid.UUID) error
}

// CreateAlumnusHandler — POST /api/v1/institutions/{id}/alumni. За rbac.RequireAuth.
func CreateAlumnusHandler(svc alumnusService, logger *slog.Logger) http.HandlerFunc {
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

		var req alumnusRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateAlumnus(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toAlumnusDTO(created))
	}
}

// DeleteAlumnusHandler — DELETE /api/v1/institutions/{id}/alumni/{alumnusId}. За rbac.RequireAuth.
func DeleteAlumnusHandler(svc alumnusService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("alumnusId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"alumnusId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfAlumnus(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.DeleteAlumnus(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

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

type galleryItemRequest struct {
	S3Key     string        `json:"s3_key"`
	Label     *bilingualDTO `json:"label,omitempty"`
	SortOrder int           `json:"sort_order"`
}

func (req galleryItemRequest) toDomain() domain.CreateGalleryItemInput {
	return domain.CreateGalleryItemInput{S3Key: req.S3Key, Label: fromBilingualDTOPtr(req.Label), SortOrder: req.SortOrder}
}

type galleryService interface {
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	IsOwnerOfGalleryItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	CreateGalleryItem(ctx context.Context, institutionID uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error)
	DeleteGalleryItem(ctx context.Context, id uuid.UUID) error
}

// CreateGalleryItemHandler — POST /api/v1/institutions/{id}/gallery. За rbac.RequireAuth.
func CreateGalleryItemHandler(svc galleryService, logger *slog.Logger) http.HandlerFunc {
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

		var req galleryItemRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateGalleryItem(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toGalleryItemDTO(created))
	}
}

// DeleteGalleryItemHandler — DELETE /api/v1/institutions/{id}/gallery/{itemId}. За rbac.RequireAuth.
func DeleteGalleryItemHandler(svc galleryService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("itemId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"itemId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfGalleryItem(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.DeleteGalleryItem(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

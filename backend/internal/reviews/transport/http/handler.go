// Package http — транспортный слой reviews поверх net/http.
package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/reviews/domain"
)

// reviewDTO — публичный контракт отзыва.
type reviewDTO struct {
	ID            uuid.UUID  `json:"id"`
	InstitutionID uuid.UUID  `json:"institution_id"`
	UserID        uuid.UUID  `json:"user_id"`
	Rating        int        `json:"rating"`
	Text          string     `json:"text"`
	Reply         *string    `json:"reply,omitempty"`
	RepliedAt     *time.Time `json:"replied_at,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

func toReviewDTO(r domain.Review) reviewDTO {
	return reviewDTO{
		ID: r.ID, InstitutionID: r.InstitutionID, UserID: r.UserID, Rating: r.Rating, Text: r.Text,
		Reply: r.Reply, RepliedAt: r.RepliedAt, Status: string(r.Status), CreatedAt: r.CreatedAt,
	}
}

type listReviewsResponse struct {
	Items []reviewDTO `json:"items"`
}

type createReviewRequest struct {
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}

type replyRequest struct {
	Reply string `json:"reply"`
}

// reviewService — то, что нужно транспорту от usecase-слоя.
type reviewService interface {
	Create(ctx context.Context, userID uuid.UUID, in domain.CreateReviewInput) (domain.Review, error)
	ListApproved(ctx context.Context, institutionID uuid.UUID) ([]domain.Review, error)
	ListMine(ctx context.Context, institutionID uuid.UUID) ([]domain.Review, error)
	Reply(ctx context.Context, reviewID uuid.UUID, actorUserID uuid.UUID, isPrivileged bool, reply string) (domain.Review, error)
	IsOwnerOfReview(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) (bool, error)
}

func isPrivilegedRole(role string) bool {
	return role == "moderator" || role == "admin"
}

// ListHandler — GET /api/v1/institutions/{id}/reviews: только approved, публичный доступ.
func ListHandler(svc reviewService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		items, err := svc.ListApproved(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]reviewDTO, len(items))
		for i, rv := range items {
			dtos[i] = toReviewDTO(rv)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listReviewsResponse{Items: dtos})
	}
}

// ListMineHandler — GET /api/v1/institutions/{id}/reviews/mine: все статусы — владелец/moderator.
// За rbac.RequireAuth.
func ListMineHandler(svc reviewService, ownerCheck func(context.Context, uuid.UUID, uuid.UUID) (bool, error), logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !isPrivilegedRole(principal.Role) {
			owner, err := ownerCheck(r.Context(), institutionID, principal.UserID)
			if err != nil {
				httpx.WriteError(w, r, logger, err)
				return
			}
			if !owner {
				httpx.WriteError(w, r, logger, apperr.Forbidden("вы не владелец этой институции"))
				return
			}
		}
		items, err := svc.ListMine(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]reviewDTO, len(items))
		for i, rv := range items {
			dtos[i] = toReviewDTO(rv)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listReviewsResponse{Items: dtos})
	}
}

// CreateHandler — POST /api/v1/institutions/{id}/reviews. За rbac.RequireAuth.
func CreateHandler(svc reviewService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		var req createReviewRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		created, err := svc.Create(r.Context(), principal.UserID, domain.CreateReviewInput{InstitutionID: institutionID, Rating: req.Rating, Text: req.Text})
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toReviewDTO(created))
	}
}

// ReplyHandler — POST /api/v1/reviews/{reviewId}/reply. За rbac.RequireAuth.
func ReplyHandler(svc reviewService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}
		reviewID, err := uuid.Parse(r.PathValue("reviewId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"reviewId": "невалидный UUID"}, "некорректный id"))
			return
		}

		var req replyRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		updated, err := svc.Reply(r.Context(), reviewID, principal.UserID, isPrivilegedRole(principal.Role), req.Reply)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toReviewDTO(updated))
	}
}

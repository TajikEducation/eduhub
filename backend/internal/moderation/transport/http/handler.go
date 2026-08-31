// Package http — транспортный слой moderation поверх net/http: approve/reject институций.
package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// moderationService — то, что нужно транспорту от usecase-слоя.
type moderationService interface {
	ApproveInstitution(ctx context.Context, actor rbac.Principal, requestID string, institutionID uuid.UUID) error
	RejectInstitution(ctx context.Context, actor rbac.Principal, requestID string, institutionID uuid.UUID, reasonCode, reasonText string) error
	ApproveReview(ctx context.Context, actor rbac.Principal, requestID string, reviewID uuid.UUID) error
	RejectReview(ctx context.Context, actor rbac.Principal, requestID string, reviewID uuid.UUID, reasonCode, reasonText string) error
}

// rejectRequest — тело POST /api/v1/moderation/institutions/{id}/reject.
type rejectRequest struct {
	ReasonCode string `json:"reason_code"`
	ReasonText string `json:"reason_text"`
}

// principalOrUnauthorized — общая часть Approve/Reject: без RequireAuth Principal в
// контексте нет — конфигурационная ошибка роутинга (401, не 403, см. rbac.RequireRole).
func principalOrUnauthorized(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (rbac.Principal, bool) {
	principal, ok := rbac.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
		return rbac.Principal{}, false
	}
	return principal, true
}

// ApproveHandler — POST /api/v1/moderation/institutions/{id}/approve. Должен идти за
// rbac.RequireAuth + rbac.RequireRole("moderator","admin").
func ApproveHandler(svc moderationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalOrUnauthorized(w, r, logger)
		if !ok {
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		if err := svc.ApproveInstitution(r.Context(), principal, httpx.RequestID(r.Context()), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// RejectHandler — POST /api/v1/moderation/institutions/{id}/reject. Должен идти за
// rbac.RequireAuth + rbac.RequireRole("moderator","admin").
func RejectHandler(svc moderationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalOrUnauthorized(w, r, logger)
		if !ok {
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		var req rejectRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if err := svc.RejectInstitution(r.Context(), principal, httpx.RequestID(r.Context()), id, req.ReasonCode, req.ReasonText); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ApproveReviewHandler — POST /api/v1/moderation/reviews/{reviewId}/approve. Должен идти за
// rbac.RequireAuth + rbac.RequireRole("moderator","admin").
func ApproveReviewHandler(svc moderationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalOrUnauthorized(w, r, logger)
		if !ok {
			return
		}

		id, err := uuid.Parse(r.PathValue("reviewId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"reviewId": "невалидный UUID"}, "некорректный id"))
			return
		}

		if err := svc.ApproveReview(r.Context(), principal, httpx.RequestID(r.Context()), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// RejectReviewHandler — POST /api/v1/moderation/reviews/{reviewId}/reject. Должен идти за
// rbac.RequireAuth + rbac.RequireRole("moderator","admin").
func RejectReviewHandler(svc moderationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalOrUnauthorized(w, r, logger)
		if !ok {
			return
		}

		id, err := uuid.Parse(r.PathValue("reviewId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"reviewId": "невалидный UUID"}, "некорректный id"))
			return
		}

		var req rejectRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if err := svc.RejectReview(r.Context(), principal, httpx.RequestID(r.Context()), id, req.ReasonCode, req.ReasonText); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

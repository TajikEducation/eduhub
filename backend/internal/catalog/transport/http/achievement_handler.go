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

type achievementRequest struct {
	Title       bilingualDTO         `json:"title"`
	Year        int                  `json:"year"`
	Category    string               `json:"category"`
	Description bilingualDTO         `json:"description"`
	Links       []achievementLinkDTO `json:"links,omitempty"`
}

func (req achievementRequest) toDomain() domain.CreateAchievementInput {
	links := make([]domain.AchievementLink, len(req.Links))
	for i, l := range req.Links {
		links[i] = domain.AchievementLink{Label: l.Label, URL: l.URL}
	}
	return domain.CreateAchievementInput{
		Title: domain.Bilingual{RU: req.Title.RU, TG: req.Title.TG}, Year: req.Year, Category: req.Category,
		Description: domain.Bilingual{RU: req.Description.RU, TG: req.Description.TG}, Links: links,
	}
}

type achievementService interface {
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	IsOwnerOfAchievement(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	CreateAchievement(ctx context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error)
	DeleteAchievement(ctx context.Context, id uuid.UUID) error
}

// CreateAchievementHandler — POST /api/v1/institutions/{id}/achievements. За rbac.RequireAuth.
func CreateAchievementHandler(svc achievementService, logger *slog.Logger) http.HandlerFunc {
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

		var req achievementRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateAchievement(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toAchievementDTO(created))
	}
}

// DeleteAchievementHandler — DELETE /api/v1/institutions/{id}/achievements/{achId}. За rbac.RequireAuth.
func DeleteAchievementHandler(svc achievementService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("achId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"achId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfAchievement(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.DeleteAchievement(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

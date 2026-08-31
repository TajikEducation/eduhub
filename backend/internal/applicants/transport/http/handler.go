// Package http — транспортный слой applicants поверх net/http.
package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/applicants/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

func requirePrincipal(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (rbac.Principal, bool) {
	principal, ok := rbac.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
		return rbac.Principal{}, false
	}
	return principal, true
}

// applicantService — то, что нужно транспорту от usecase-слоя.
type applicantService interface {
	UpsertMine(ctx context.Context, userID uuid.UUID, in domain.ApplicantInput) (domain.Applicant, error)
	GetMine(ctx context.Context, userID uuid.UUID) (domain.Applicant, error)
	GetVisible(ctx context.Context, id uuid.UUID) (domain.Applicant, error)
	ListPublic(ctx context.Context) ([]domain.Applicant, error)
	ListAchievements(ctx context.Context, applicantID uuid.UUID) ([]domain.Achievement, error)
	CreateAchievement(ctx context.Context, userID uuid.UUID, in domain.AchievementInput) (domain.Achievement, error)
	DeleteAchievement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Apply(ctx context.Context, userID uuid.UUID, vacancyID uuid.UUID) (domain.Application, error)
	ListMyApplications(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// ListPublicHandler — GET /api/v1/applicants: полностью публичные профили.
func ListPublicHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.ListPublic(r.Context())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]applicantDTO, len(items))
		for i, a := range items {
			dtos[i] = toApplicantDTO(a, false)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listApplicantsResponse{Items: dtos})
	}
}

// GetHandler — GET /api/v1/applicants/{id}: публичный профиль (кроме draft).
func GetHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		a, err := svc.GetVisible(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toApplicantDTO(a, false))
	}
}

// ListAchievementsHandler — GET /api/v1/applicants/{id}/achievements: публично, вместе с профилем.
func ListAchievementsHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		items, err := svc.ListAchievements(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]achievementDTO, len(items))
		for i, a := range items {
			dtos[i] = toAchievementDTO(a)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listAchievementsResponse{Items: dtos})
	}
}

// GetMineHandler — GET /api/v1/applicants/me. За rbac.RequireAuth.
func GetMineHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		a, err := svc.GetMine(r.Context(), principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toApplicantDTO(a, true))
	}
}

// UpsertMineHandler — PUT /api/v1/applicants/me. За rbac.RequireAuth.
func UpsertMineHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		var req applicantRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		a, err := svc.UpsertMine(r.Context(), principal.UserID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toApplicantDTO(a, true))
	}
}

// CreateAchievementHandler — POST /api/v1/applicants/me/achievements. За rbac.RequireAuth.
func CreateAchievementHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		var req achievementRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateAchievement(r.Context(), principal.UserID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toAchievementDTO(created))
	}
}

// DeleteAchievementHandler — DELETE /api/v1/applicants/achievements/{achId}. За rbac.RequireAuth.
func DeleteAchievementHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
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
		if err := svc.DeleteAchievement(r.Context(), id, principal.UserID); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ApplyHandler — POST /api/v1/vacancies/{id}/apply: отклик на вакансию. За rbac.RequireAuth.
func ApplyHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		vacancyID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		app, err := svc.Apply(r.Context(), principal.UserID, vacancyID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toApplicationDTO(app))
	}
}

// ListMyApplicationsHandler — GET /api/v1/applicants/me/applications. За rbac.RequireAuth.
func ListMyApplicationsHandler(svc applicantService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		ids, err := svc.ListMyApplications(r.Context(), principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		if ids == nil {
			ids = []uuid.UUID{}
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listMyApplicationsResponse{VacancyIDs: ids})
	}
}

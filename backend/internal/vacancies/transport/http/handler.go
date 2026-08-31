// Package http — транспортный слой vacancies поверх net/http.
package http

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
)

// defaultGlobalListLimit — сколько вакансий отдаём по умолчанию в /api/v1/vacancies. Небольшой
// MVP-каталог не требует keyset-пагинации (см. тот же выбор в catalog для малых списков).
const defaultGlobalListLimit = 100

// requirePrincipal читает Principal из контекста или пишет 401 — тот же паттерн, что
// internal/catalog/transport/http/satellite_helpers.go (не переиспользуем напрямую: функция
// unexported в чужом пакете).
func requirePrincipal(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (rbac.Principal, bool) {
	principal, ok := rbac.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
		return rbac.Principal{}, false
	}
	return principal, true
}

func isPrivilegedRole(role string) bool {
	return role == "moderator" || role == "admin"
}

// requireOwnerOrPrivileged проверяет доступ: moderator/admin — всегда, иначе checkOwner()
// должен вернуть true.
func requireOwnerOrPrivileged(w http.ResponseWriter, r *http.Request, logger *slog.Logger, principal rbac.Principal, checkOwner func(context.Context) (bool, error)) bool {
	if isPrivilegedRole(principal.Role) {
		return true
	}
	owner, err := checkOwner(r.Context())
	if err != nil {
		httpx.WriteError(w, r, logger, err)
		return false
	}
	if !owner {
		httpx.WriteError(w, r, logger, apperr.Forbidden("доступ запрещён"))
		return false
	}
	return true
}

// vacancyService — то, что нужно транспорту от usecase-слоя.
type vacancyService interface {
	Create(ctx context.Context, institutionID uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error)
	Update(ctx context.Context, id uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListMine(ctx context.Context, institutionID uuid.UUID) ([]domain.Vacancy, error)
	ListPublic(ctx context.Context, institutionID uuid.UUID) ([]domain.Vacancy, error)
	ListGlobalPublished(ctx context.Context, limit int) ([]domain.Vacancy, error)
	GetPublished(ctx context.Context, id uuid.UUID) (domain.Vacancy, error)
	IsOwnerOfVacancy(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
}

// institutionOwnerChecker — то, что нужно транспорту для проверки владения при создании
// вакансии (institution_id известен из URL, не из самой вакансии).
type institutionOwnerChecker func(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)

// institutionSummaryFetcher — порт в catalog для обогащения глобального списка/детали сводкой
// об учреждении (имя/тип/регион/фото) — инжектируется из composition root (cmd/api/router.go),
// который единственный видит оба модуля (vacancies и catalog) одновременно.
type institutionSummaryFetcher func(ctx context.Context, institutionID uuid.UUID) (domain.InstitutionSummary, error)

// ListByInstitutionHandler — GET /api/v1/institutions/{id}/vacancies: только опубликованные,
// публичный доступ (вкладка «Вакансии» профиля учреждения).
func ListByInstitutionHandler(svc vacancyService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		items, err := svc.ListPublic(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]vacancyDTO, len(items))
		for i, v := range items {
			dtos[i] = toVacancyDTO(v)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listVacanciesResponse{Items: dtos})
	}
}

// ListMineHandler — GET /api/v1/institutions/{id}/vacancies/mine: все статусы —
// владелец/moderator. За rbac.RequireAuth.
func ListMineHandler(svc vacancyService, ownerCheck institutionOwnerChecker, logger *slog.Logger) http.HandlerFunc {
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
			return ownerCheck(ctx, institutionID, principal.UserID)
		}) {
			return
		}
		items, err := svc.ListMine(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]vacancyDTO, len(items))
		for i, v := range items {
			dtos[i] = toVacancyDTO(v)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listVacanciesResponse{Items: dtos})
	}
}

// CreateHandler — POST /api/v1/institutions/{id}/vacancies. За rbac.RequireAuth.
func CreateHandler(svc vacancyService, ownerCheck institutionOwnerChecker, logger *slog.Logger) http.HandlerFunc {
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
			return ownerCheck(ctx, institutionID, principal.UserID)
		}) {
			return
		}
		var req vacancyRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.Create(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toVacancyDTO(created))
	}
}

// UpdateHandler — PATCH /api/v1/vacancies/{vacancyId}. За rbac.RequireAuth.
func UpdateHandler(svc vacancyService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("vacancyId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"vacancyId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfVacancy(ctx, id, principal.UserID)
		}) {
			return
		}
		var req vacancyRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		updated, err := svc.Update(r.Context(), id, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toVacancyDTO(updated))
	}
}

// DeleteHandler — DELETE /api/v1/vacancies/{vacancyId}. За rbac.RequireAuth.
func DeleteHandler(svc vacancyService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("vacancyId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"vacancyId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfVacancy(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.Delete(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ListGlobalHandler — GET /api/v1/vacancies: опубликованные вакансии всех институций,
// обогащённые сводкой об учреждении. Регион/тип/поиск/занятость фильтруются на фронте по
// встроенной сводке — небольшой MVP-каталог, без серверной пагинации по фильтрам (см. пакетный
// комментарий VacancyRepo.ListPublished).
func ListGlobalHandler(svc vacancyService, institutionSummary institutionSummaryFetcher, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := defaultGlobalListLimit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		items, err := svc.ListGlobalPublished(r.Context(), limit)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]vacancyWithInstitutionDTO, 0, len(items))
		for _, v := range items {
			summary, err := institutionSummary(r.Context(), v.InstitutionID)
			if err != nil {
				// Учреждение недоступно (не approved/удалено) — осиротевшую вакансию молча
				// пропускаем, не валим весь список (best-effort обогащение).
				continue
			}
			dtos = append(dtos, toVacancyWithInstitutionDTO(v, summary))
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listVacanciesWithInstitutionResponse{Items: dtos})
	}
}

// GetGlobalHandler — GET /api/v1/vacancies/{vacancyId}: одна опубликованная вакансия с
// сводкой об учреждении.
func GetGlobalHandler(svc vacancyService, institutionSummary institutionSummaryFetcher, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("vacancyId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"vacancyId": "невалидный UUID"}, "некорректный id"))
			return
		}
		v, err := svc.GetPublished(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		summary, err := institutionSummary(r.Context(), v.InstitutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toVacancyWithInstitutionDTO(v, summary))
	}
}

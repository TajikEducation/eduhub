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

type newsRequest struct {
	Title      bilingualDTO   `json:"title"`
	Category   *bilingualDTO  `json:"category,omitempty"`
	CoverS3Key *string        `json:"cover_s3_key,omitempty"`
	VideoURL   *string        `json:"video_url,omitempty"`
	Content    bilingualDTO   `json:"content"`
	Tags       []bilingualDTO `json:"tags,omitempty"`
	Status     string         `json:"status"`
}

func (req newsRequest) toDomain() domain.CreateNewsInput {
	tags := make([]domain.Bilingual, len(req.Tags))
	for i, tg := range req.Tags {
		tags[i] = domain.Bilingual{RU: tg.RU, TG: tg.TG}
	}
	return domain.CreateNewsInput{
		Title: domain.Bilingual{RU: req.Title.RU, TG: req.Title.TG}, Category: fromBilingualDTOPtr(req.Category),
		CoverS3Key: req.CoverS3Key, VideoURL: req.VideoURL,
		Content: domain.Bilingual{RU: req.Content.RU, TG: req.Content.TG}, Tags: tags, Status: req.Status,
	}
}

type newsService interface {
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	IsOwnerOfNews(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)
	ListNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error)
	ListPublishedNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error)
	GetPublishedNews(ctx context.Context, id uuid.UUID) (domain.NewsArticle, error)
	CreateNews(ctx context.Context, institutionID uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error)
	UpdateNews(ctx context.Context, id uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error)
	DeleteNews(ctx context.Context, id uuid.UUID) error
}

// listNewsResponse — тело ответа GET /api/v1/institutions/{id}/news.
type listNewsResponse struct {
	Items []newsArticleDTO `json:"items"`
}

// ListNewsHandler — GET /api/v1/institutions/{id}/news: все новости (включая draft) — для
// кабинета учреждения, не для публичной страницы. За rbac.RequireAuth.
func ListNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
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

		items, err := svc.ListNews(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]newsArticleDTO, len(items))
		for i, n := range items {
			dtos[i] = toNewsArticleDTO(n)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listNewsResponse{Items: dtos})
	}
}

// ListPublishedNewsHandler — GET /api/v1/institutions/{id}/news/published: только
// опубликованные новости, публичный доступ (вкладка «Новости» профиля учреждения, FR-24).
func ListPublishedNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		institutionID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}
		items, err := svc.ListPublishedNews(r.Context(), institutionID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		dtos := make([]newsArticleDTO, len(items))
		for i, n := range items {
			dtos[i] = toNewsArticleDTO(n)
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, listNewsResponse{Items: dtos})
	}
}

// GetNewsHandler — GET /api/v1/news/{newsId}: одна опубликованная новость, публичный доступ
// (страница /news/{id}). Черновики чужой институции анонимному посетителю недоступны.
func GetNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("newsId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"newsId": "невалидный UUID"}, "некорректный id"))
			return
		}
		article, err := svc.GetPublishedNews(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toNewsArticleDTO(article))
	}
}

// CreateNewsHandler — POST /api/v1/institutions/{id}/news. За rbac.RequireAuth.
func CreateNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
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

		var req newsRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		created, err := svc.CreateNews(r.Context(), institutionID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toNewsArticleDTO(created))
	}
}

// UpdateNewsHandler — PATCH /api/v1/institutions/{id}/news/{newsId}. За rbac.RequireAuth.
func UpdateNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("newsId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"newsId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfNews(ctx, id, principal.UserID)
		}) {
			return
		}

		var req newsRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		updated, err := svc.UpdateNews(r.Context(), id, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toNewsArticleDTO(updated))
	}
}

// DeleteNewsHandler — DELETE /api/v1/institutions/{id}/news/{newsId}. За rbac.RequireAuth.
func DeleteNewsHandler(svc newsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, logger)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("newsId"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"newsId": "невалидный UUID"}, "некорректный id"))
			return
		}
		if !requireOwnerOrPrivileged(w, r, logger, principal, func(ctx context.Context) (bool, error) {
			return svc.IsOwnerOfNews(ctx, id, principal.UserID)
		}) {
			return
		}
		if err := svc.DeleteNews(r.Context(), id); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

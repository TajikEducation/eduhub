package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// listService — то, что нужно ListHandler от usecase-слоя (порт, определённый потребителем).
type listService interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
}

// ListHandler возвращает http.HandlerFunc для GET /api/v1/institutions: парсит и валидирует
// query-параметры, делегирует usecase-слою, маппит результат в публичный DTO-контракт.
func ListHandler(svc listService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := parseListQuery(r.URL.Query())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if err := filter.Validate(); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		result, err := svc.List(r.Context(), filter)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=60")
		_ = httpx.WriteJSON(w, logger, http.StatusOK, toListResponse(result))
	}
}

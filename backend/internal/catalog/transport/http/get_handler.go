package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// getService — то, что нужно GetHandler от usecase-слоя (порт, определённый потребителем).
type getService interface {
	Get(ctx context.Context, id uuid.UUID) (domain.Institution, error)
}

// buildETag строит детерминированный ETag из id+updated_at — тот же вход всегда даёт тот же ETag,
// любое изменение updated_at (после мутации, веха 3) меняет его.
func buildETag(id uuid.UUID, updatedAt time.Time) string {
	return fmt.Sprintf("%q", id.String()+"."+strconv.FormatInt(updatedAt.UnixNano(), 10))
}

// GetHandler возвращает http.HandlerFunc для GET /api/v1/institutions/{id}: валидирует id из пути,
// делегирует usecase-слою, маппит результат в публичный DTO-контракт полной карточки и поддерживает
// условные запросы через ETag/If-None-Match.
func GetHandler(svc getService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id": "невалидный UUID"}, "некорректный id"))
			return
		}

		inst, err := svc.Get(r.Context(), id)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		etag := buildETag(inst.ID, inst.UpdatedAt)
		w.Header().Set("ETag", etag)

		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toInstitutionDTO(inst))
	}
}

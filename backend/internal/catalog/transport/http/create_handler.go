package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// createRequest — тело POST /api/v1/institutions (E3.3). Зеркалит RegisterInstitutionInput
// из web/lib/app-state.tsx — минимум, остальное в domain.CreateInstitutionInput — дефолты.
type createRequest struct {
	Name        bilingualDTO  `json:"name"`
	Types       []string      `json:"types"`
	Region      string        `json:"region"`
	City        *bilingualDTO `json:"city,omitempty"`
	District    *string       `json:"district,omitempty"`
	Description *bilingualDTO `json:"description,omitempty"`
	Phone       *string       `json:"phone,omitempty"`
	Email       *string       `json:"email,omitempty"`
	Website     *string       `json:"website,omitempty"`
	Price       *int          `json:"price,omitempty"`
	Lat         float64       `json:"lat"`
	Lng         float64       `json:"lng"`
}

func (req createRequest) toDomain() domain.CreateInstitutionInput {
	return domain.CreateInstitutionInput{
		Name:        domain.Bilingual{RU: req.Name.RU, TG: req.Name.TG},
		Types:       req.Types,
		Region:      req.Region,
		City:        fromBilingualDTOPtr(req.City),
		District:    req.District,
		Description: fromBilingualDTOPtr(req.Description),
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		Price:       req.Price,
		Lat:         req.Lat,
		Lng:         req.Lng,
	}
}

func fromBilingualDTOPtr(b *bilingualDTO) *domain.Bilingual {
	if b == nil {
		return nil
	}
	return &domain.Bilingual{RU: b.RU, TG: b.TG}
}

// createService — то, что нужно CreateHandler от usecase-слоя.
type createService interface {
	Register(ctx context.Context, ownerID uuid.UUID, in domain.CreateInstitutionInput) (domain.Institution, error)
}

// CreateHandler — POST /api/v1/institutions (E3.3). Должен идти за rbac.RequireAuth —
// читает Principal из контекста как owner_id.
func CreateHandler(svc createService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}

		var req createRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		created, err := svc.Register(r.Context(), principal.UserID, req.toDomain())
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusCreated, toInstitutionDTO(created))
	}
}

package http

import (
	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// listResponse — тело ответа GET /api/v1/institutions.
type listResponse struct {
	Items      []institutionListItemDTO `json:"items"`
	NextCursor *string                  `json:"next_cursor"`
	// TotalHint всегда nil сейчас — keyset-пагинация не делает COUNT-запрос (задача 23).
	TotalHint *int `json:"total_hint"`
}

// bilingualDTO — двуязычное поле (ru/tg) в публичном JSON-контракте.
type bilingualDTO struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

// institutionListItemDTO — карточка учреждения в списке. Не отдаёт наружу профильные/владельческие
// поля (ModerationStatus/PlanExpiresAt/LicenseNo/Socials/LocationLandmarks/Phone/Email/Website)
// и сателлитные коллекции — List() их не заполняет.
type institutionListItemDTO struct {
	ID                uuid.UUID     `json:"id"`
	Name              bilingualDTO  `json:"name"`
	Types             []string      `json:"types"`
	Region            string        `json:"region"`
	City              *bilingualDTO `json:"city,omitempty"`
	District          *string       `json:"district,omitempty"`
	Price             *int          `json:"price,omitempty"`
	RatingAvg         *float64      `json:"rating_avg,omitempty"`
	ReviewCount       int           `json:"review_count"`
	Verified          bool          `json:"verified"`
	DiscountAvailable bool          `json:"discount_available"`
	CoverPhotoS3Key   *string       `json:"cover_photo_s3_key,omitempty"`
	Tag               *bilingualDTO `json:"tag,omitempty"`
	// DistanceM — omitempty критично: nil означает «гео в фильтре не запрашивался», не «0».
	DistanceM *float64 `json:"distance_m,omitempty"`
}

// toListResponse маппит доменный ListResult в публичный DTO-контракт.
func toListResponse(result domain.ListResult) listResponse {
	items := make([]institutionListItemDTO, len(result.Items))
	for i, inst := range result.Items {
		items[i] = toListItemDTO(inst)
	}
	return listResponse{Items: items, NextCursor: result.NextCursor, TotalHint: nil}
}

// toListItemDTO маппит одну Institution в карточку списка, поле в поле.
func toListItemDTO(inst domain.Institution) institutionListItemDTO {
	return institutionListItemDTO{
		ID:                inst.ID,
		Name:              toBilingualDTO(inst.Name),
		Types:             inst.Types,
		Region:            inst.Region,
		City:              toBilingualDTOPtr(inst.City),
		District:          inst.District,
		Price:             inst.Price,
		RatingAvg:         inst.RatingAvg,
		ReviewCount:       inst.ReviewCount,
		Verified:          inst.Verified,
		DiscountAvailable: inst.DiscountAvailable,
		CoverPhotoS3Key:   inst.CoverPhotoS3Key,
		Tag:               toBilingualDTOPtr(inst.Tag),
		DistanceM:         inst.DistanceM,
	}
}

func toBilingualDTO(b domain.Bilingual) bilingualDTO {
	return bilingualDTO{RU: b.RU, TG: b.TG}
}

func toBilingualDTOPtr(b *domain.Bilingual) *bilingualDTO {
	if b == nil {
		return nil
	}
	dto := toBilingualDTO(*b)
	return &dto
}

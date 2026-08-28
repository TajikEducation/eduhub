package domain

import "github.com/abdulhalim/eduhub/backend/internal/platform/apperr"

// Filter — параметры листинга каталога учреждений (GET /api/v1/institutions).
// Sort-enum и формат Cursor валидируются в других слоях (не здесь).
type Filter struct {
	Q            *string
	Types        []string
	Region       *string
	Area         *string
	MinPrice     *int
	MaxPrice     *int
	MinRating    *float64
	Transport    *bool
	Food         *bool
	Verified     *bool
	Curriculum   []string // пересечение (SQL &&) — институция матчит, если есть хотя бы одно совпадение
	ProgramLevel []string // та же семантика пересечения
	Discount     *bool
	Sort         string
	Lat          *float64
	Lng          *float64
	RadiusKm     *float64
	Limit        int
	Cursor       *string
}

const (
	defaultLimit = 20
	maxLimit     = 50
)

// Normalize приводит Limit к дефолту/потолку. Вызывается до Validate.
func (f *Filter) Normalize() {
	if f.Limit == 0 {
		f.Limit = defaultLimit
	}
	if f.Limit > maxLimit {
		f.Limit = maxLimit
	}
}

// Validate проверяет только межполевые инварианты фильтра — не enum/формат отдельных полей.
func (f *Filter) Validate() error {
	if f.MinPrice != nil && f.MaxPrice != nil && *f.MinPrice > *f.MaxPrice {
		return apperr.Invalid(
			map[string]string{"min_price": "не может быть больше max_price"},
			"min_price больше max_price",
		)
	}

	latSet := f.Lat != nil
	lngSet := f.Lng != nil
	if latSet != lngSet {
		if !latSet {
			return apperr.Invalid(map[string]string{"lat": "обязателен вместе с lng"}, "lat отсутствует")
		}
		return apperr.Invalid(map[string]string{"lng": "обязателен вместе с lat"}, "lng отсутствует")
	}

	if f.RadiusKm != nil && !latSet && !lngSet {
		return apperr.Invalid(
			map[string]string{"radius_km": "требует lat и lng"},
			"radius_km указан без координат",
		)
	}

	return nil
}

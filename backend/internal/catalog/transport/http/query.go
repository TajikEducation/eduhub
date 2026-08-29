// Package http — транспортный слой каталога поверх net/http: парсинг/валидация
// синтаксиса query-параметров GET /api/v1/institutions.
package http

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// parseListQuery разбирает url.Values в domain.Filter. Здесь проверяется только
// синтаксис/формат/enum отдельных параметров — межполевые инварианты (min<max,
// lat требует lng и т.п.) уже реализованы в domain.Filter.Validate() и не дублируются здесь.
func parseListQuery(values url.Values) (domain.Filter, error) {
	var f domain.Filter

	f.Q = stringPtr(values.Get("q"))
	f.Types = splitCSV(values.Get("type"))
	f.Region = stringPtr(values.Get("region"))
	f.Area = stringPtr(values.Get("area"))
	f.Curriculum = splitCSV(values.Get("curriculum"))
	f.ProgramLevel = splitCSV(values.Get("program_level"))
	f.Cursor = stringPtr(values.Get("cursor"))

	minPrice, err := parseIntPtr(values.Get("min_price"), "min_price")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MinPrice = minPrice

	maxPrice, err := parseIntPtr(values.Get("max_price"), "max_price")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MaxPrice = maxPrice

	minRating, err := parseFloatPtr(values.Get("min_rating"), "min_rating")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MinRating = minRating

	transport, err := parseBoolPtr(values.Get("transport"), "transport")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Transport = transport

	food, err := parseBoolPtr(values.Get("food"), "food")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Food = food

	verified, err := parseBoolPtr(values.Get("verified"), "verified")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Verified = verified

	discount, err := parseBoolPtr(values.Get("discount"), "discount")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Discount = discount

	sort, err := parseSort(values.Get("sort"))
	if err != nil {
		return domain.Filter{}, err
	}
	f.Sort = sort

	lat, err := parseLat(values.Get("lat"))
	if err != nil {
		return domain.Filter{}, err
	}
	f.Lat = lat

	lng, err := parseLng(values.Get("lng"))
	if err != nil {
		return domain.Filter{}, err
	}
	f.Lng = lng

	radiusKm, err := parseFloatPtr(values.Get("radius_km"), "radius_km")
	if err != nil {
		return domain.Filter{}, err
	}
	f.RadiusKm = radiusKm

	limit, err := parseLimit(values.Get("limit"))
	if err != nil {
		return domain.Filter{}, err
	}
	f.Limit = limit

	return f, nil
}

// stringPtr — пустая строка (параметр отсутствует) → nil, иначе указатель на значение.
func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// splitCSV — comma-separated значение в срез строк, без enum-валидации элементов.
func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// parseIntPtr парсит непустую строку как int, ошибка формата → apperr.Invalid по field.
func parseIntPtr(v, field string) (*int, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, apperr.Invalid(map[string]string{field: "должно быть целым числом"}, field+" невалиден")
	}
	return &n, nil
}

// parseFloatPtr парсит непустую строку как float64, ошибка формата → apperr.Invalid по field.
func parseFloatPtr(v, field string) (*float64, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, apperr.Invalid(map[string]string{field: "должно быть числом"}, field+" невалиден")
	}
	return &n, nil
}

// parseBoolPtr — пустая строка (параметр отсутствует ИЛИ присутствует без значения) → nil,
// «не указано» отличается от «явно выключено» (false).
func parseBoolPtr(v, field string) (*bool, error) {
	if v == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil, apperr.Invalid(map[string]string{field: "должно быть булевым значением"}, field+" невалиден")
	}
	return &b, nil
}

// parseSort допускает только значения, которые сейчас умеет сортировать репозиторий каталога.
func parseSort(v string) (string, error) {
	switch v {
	case "", "price_asc", "score":
		return v, nil
	default:
		return "", apperr.Invalid(map[string]string{"sort": "недопустимое значение сортировки"}, "sort невалиден")
	}
}

// parseLat парсит lat и проверяет диапазон [-90, 90] включительно.
func parseLat(v string) (*float64, error) {
	lat, err := parseFloatPtr(v, "lat")
	if err != nil {
		return nil, err
	}
	if lat != nil && (*lat < -90 || *lat > 90) {
		return nil, apperr.Invalid(map[string]string{"lat": "должно быть в диапазоне [-90, 90]"}, "lat вне диапазона")
	}
	return lat, nil
}

// parseLng парсит lng и проверяет диапазон [-180, 180] включительно.
func parseLng(v string) (*float64, error) {
	lng, err := parseFloatPtr(v, "lng")
	if err != nil {
		return nil, err
	}
	if lng != nil && (*lng < -180 || *lng > 180) {
		return nil, apperr.Invalid(map[string]string{"lng": "должно быть в диапазоне [-180, 180]"}, "lng вне диапазона")
	}
	return lng, nil
}

// parseLimit — пустая строка оставляет Limit=0 (домен нормализует в дефолт позже),
// отрицательное значение — ошибка.
func parseLimit(v string) (int, error) {
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, apperr.Invalid(map[string]string{"limit": "должно быть неотрицательным целым числом"}, "limit невалиден")
	}
	return n, nil
}

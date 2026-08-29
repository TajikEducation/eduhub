// Package http — HTTP-транспорт каталога институций.
package http

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// allowedSort — сортировки, которые сейчас умеет repo (см. задачу 22). Любое другое
// значение — apperr.Invalid, а не молчаливое игнорирование ниже по стеку.
var allowedSort = map[string]bool{
	"":          true,
	"price_asc": true,
	"score":     true,
}

// parseListQuery разбирает query-параметры GET /api/v1/institutions в domain.Filter.
// Здесь только синтаксис/формат/enum — межполевые инварианты проверяет domain.Filter.Validate().
func parseListQuery(values url.Values) (domain.Filter, error) {
	var f domain.Filter

	f.Q = optionalString(values, "q")
	f.Types = optionalCSV(values, "type")
	f.Region = optionalString(values, "region")
	f.Area = optionalString(values, "area")
	f.Curriculum = optionalCSV(values, "curriculum")
	f.ProgramLevel = optionalCSV(values, "program_level")
	f.Cursor = optionalString(values, "cursor")

	minPrice, err := optionalInt(values, "min_price")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MinPrice = minPrice

	maxPrice, err := optionalInt(values, "max_price")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MaxPrice = maxPrice

	minRating, err := optionalFloat(values, "min_rating")
	if err != nil {
		return domain.Filter{}, err
	}
	f.MinRating = minRating

	transport, err := optionalBool(values, "transport")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Transport = transport

	food, err := optionalBool(values, "food")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Food = food

	verified, err := optionalBool(values, "verified")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Verified = verified

	discount, err := optionalBool(values, "discount")
	if err != nil {
		return domain.Filter{}, err
	}
	f.Discount = discount

	sort := values.Get("sort")
	if !allowedSort[sort] {
		return domain.Filter{}, apperr.Invalid(
			map[string]string{"sort": "неизвестное значение сортировки"},
			"sort должен быть одним из: price_asc, score",
		)
	}
	f.Sort = sort

	lat, err := optionalLatLng(values, "lat", -90, 90)
	if err != nil {
		return domain.Filter{}, err
	}
	f.Lat = lat

	lng, err := optionalLatLng(values, "lng", -180, 180)
	if err != nil {
		return domain.Filter{}, err
	}
	f.Lng = lng

	radiusKm, err := optionalFloat(values, "radius_km")
	if err != nil {
		return domain.Filter{}, err
	}
	f.RadiusKm = radiusKm

	if v := values.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 0 {
			return domain.Filter{}, apperr.Invalid(
				map[string]string{"limit": "должен быть неотрицательным целым числом"},
				"некорректный limit",
			)
		}
		f.Limit = limit
	}

	return f, nil
}

// optionalString возвращает *string по ключу или nil, если параметр пуст/отсутствует.
func optionalString(values url.Values, key string) *string {
	v := values.Get(key)
	if v == "" {
		return nil
	}
	return &v
}

// optionalCSV разбивает значение по запятой в []string, или nil, если параметр пуст/отсутствует.
func optionalCSV(values url.Values, key string) []string {
	v := values.Get(key)
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// optionalInt парсит целое, или nil, если параметр пуст/отсутствует.
func optionalInt(values url.Values, key string) (*int, error) {
	v := values.Get(key)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, apperr.Invalid(
			map[string]string{key: "должен быть целым числом"},
			"некорректный "+key,
		)
	}
	return &n, nil
}

// optionalFloat парсит float64, или nil, если параметр пуст/отсутствует.
func optionalFloat(values url.Values, key string) (*float64, error) {
	v := values.Get(key)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, apperr.Invalid(
			map[string]string{key: "должен быть числом"},
			"некорректный "+key,
		)
	}
	return &n, nil
}

// optionalLatLng парсит координату и проверяет диапазон [min, max].
func optionalLatLng(values url.Values, key string, min, max float64) (*float64, error) {
	n, err := optionalFloat(values, key)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	if *n < min || *n > max {
		return nil, apperr.Invalid(
			map[string]string{key: "вне допустимого диапазона"},
			"некорректный "+key,
		)
	}
	return n, nil
}

// optionalBool парсит булево значение. Пустая строка (параметр отсутствует ИЛИ
// присутствует без значения) — nil, а НЕ false: «не указано» и «явно выключено» различаются.
func optionalBool(values url.Values, key string) (*bool, error) {
	v := values.Get(key)
	if v == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil, apperr.Invalid(
			map[string]string{key: "должен быть булевым значением"},
			"некорректный "+key,
		)
	}
	return &b, nil
}

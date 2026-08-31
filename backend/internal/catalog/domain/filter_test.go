package domain

import (
	"errors"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// p — тестовый хелпер: адрес переданного значения.
func p[T any](v T) *T {
	return &v
}

func TestFilter_Validate_MinPriceGreaterThanMaxPrice(t *testing.T) {
	f := Filter{MinPrice: p(500), MaxPrice: p(100)}

	err := f.Validate()

	if err == nil {
		t.Fatal("ожидалась ошибка, получили nil")
	}
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("ожидалась *apperr.Error, получили %T: %v", err, err)
	}
	if _, ok := appErr.Fields["min_price"]; !ok {
		t.Fatalf("ожидалось поле min_price в Fields, получили: %v", appErr.Fields)
	}
}

func TestFilter_Normalize_DefaultLimit(t *testing.T) {
	f := Filter{Limit: 0}

	f.Normalize()

	if f.Limit != 20 {
		t.Fatalf("ожидали Limit == 20, получили %d", f.Limit)
	}
}

func TestFilter_Normalize_CapsLimitAt50(t *testing.T) {
	f := Filter{Limit: 500}

	f.Normalize()

	if f.Limit != 50 {
		t.Fatalf("ожидали Limit == 50, получили %d", f.Limit)
	}
}

func TestFilter_Validate_LatWithoutLng(t *testing.T) {
	f := Filter{Lat: p(38.5)}

	if err := f.Validate(); err == nil {
		t.Fatal("ожидалась ошибка при Lat без Lng, получили nil")
	}
}

func TestFilter_Validate_RadiusWithoutCoords(t *testing.T) {
	f := Filter{RadiusKm: p(10.0)}

	if err := f.Validate(); err == nil {
		t.Fatal("ожидалась ошибка при RadiusKm без координат, получили nil")
	}
}

func TestFilter_Validate_EmptyFilterIsValid(t *testing.T) {
	f := Filter{}

	if err := f.Validate(); err != nil {
		t.Fatalf("пустой фильтр должен быть валиден, получили: %v", err)
	}
}

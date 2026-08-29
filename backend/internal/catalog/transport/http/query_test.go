package http

import (
	"errors"
	"net/url"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// TestParseListQuery покрывает только синтаксис/формат/enum отдельных query-параметров —
// межполевые инварианты (min<max, lat+lng) проверяет domain.Filter.Validate(), не этот парсер.
func TestParseListQuery(t *testing.T) {
	t.Run("sort=hack невалиден", func(t *testing.T) {
		values := url.Values{"sort": {"hack"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "sort")
	})

	t.Run("min_price=abc не парсится", func(t *testing.T) {
		values := url.Values{"min_price": {"abc"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "min_price")
	})

	t.Run("max_price=xyz не парсится", func(t *testing.T) {
		values := url.Values{"max_price": {"xyz"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "max_price")
	})

	t.Run("lat=200 вне диапазона", func(t *testing.T) {
		values := url.Values{"lat": {"200"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "lat")
	})

	t.Run("lng=-200 вне диапазона", func(t *testing.T) {
		values := url.Values{"lng": {"-200"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "lng")
	})

	t.Run("limit=-1 невалиден", func(t *testing.T) {
		values := url.Values{"limit": {"-1"}}

		_, err := parseListQuery(values)

		assertInvalidField(t, err, "limit")
	})

	t.Run("type=cat_school,cat_uni парсится в срез", func(t *testing.T) {
		values := url.Values{"type": {"cat_school,cat_uni"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		want := []string{"cat_school", "cat_uni"}
		if len(filter.Types) != len(want) || filter.Types[0] != want[0] || filter.Types[1] != want[1] {
			t.Errorf("Types = %v, want %v", filter.Types, want)
		}
	})

	t.Run("transport=1 парсится в true", func(t *testing.T) {
		values := url.Values{"transport": {"1"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Transport == nil || !*filter.Transport {
			t.Errorf("Transport = %v, want *true", filter.Transport)
		}
	})

	t.Run("transport= (пустое значение) означает не указано, не false", func(t *testing.T) {
		values := url.Values{"transport": {""}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Transport != nil {
			t.Errorf("Transport = %v, want nil (не указано)", *filter.Transport)
		}
	})

	t.Run("curriculum=bilingual,stem парсится в срез", func(t *testing.T) {
		values := url.Values{"curriculum": {"bilingual,stem"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		want := []string{"bilingual", "stem"}
		if len(filter.Curriculum) != len(want) || filter.Curriculum[0] != want[0] || filter.Curriculum[1] != want[1] {
			t.Errorf("Curriculum = %v, want %v", filter.Curriculum, want)
		}
	})

	t.Run("discount=true парсится в true", func(t *testing.T) {
		values := url.Values{"discount": {"true"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Discount == nil || !*filter.Discount {
			t.Errorf("Discount = %v, want *true", filter.Discount)
		}
	})

	t.Run("q=гулистон парсится в указатель", func(t *testing.T) {
		values := url.Values{"q": {"гулистон"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Q == nil || *filter.Q != "гулистон" {
			t.Errorf("Q = %v, want *\"гулистон\"", filter.Q)
		}
	})

	t.Run("region=dushanbe единственный параметр — остальные нулевые", func(t *testing.T) {
		values := url.Values{"region": {"dushanbe"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Region == nil || *filter.Region != "dushanbe" {
			t.Errorf("Region = %v, want *\"dushanbe\"", filter.Region)
		}
		if filter.Q != nil {
			t.Errorf("Q = %v, want nil", *filter.Q)
		}
		if filter.Types != nil {
			t.Errorf("Types = %v, want nil", filter.Types)
		}
		if filter.Area != nil {
			t.Errorf("Area = %v, want nil", *filter.Area)
		}
		if filter.MinPrice != nil {
			t.Errorf("MinPrice = %v, want nil", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			t.Errorf("MaxPrice = %v, want nil", *filter.MaxPrice)
		}
		if filter.MinRating != nil {
			t.Errorf("MinRating = %v, want nil", *filter.MinRating)
		}
		if filter.Transport != nil {
			t.Errorf("Transport = %v, want nil", *filter.Transport)
		}
		if filter.Food != nil {
			t.Errorf("Food = %v, want nil", *filter.Food)
		}
		if filter.Verified != nil {
			t.Errorf("Verified = %v, want nil", *filter.Verified)
		}
		if filter.Curriculum != nil {
			t.Errorf("Curriculum = %v, want nil", filter.Curriculum)
		}
		if filter.ProgramLevel != nil {
			t.Errorf("ProgramLevel = %v, want nil", filter.ProgramLevel)
		}
		if filter.Discount != nil {
			t.Errorf("Discount = %v, want nil", *filter.Discount)
		}
		if filter.Sort != "" {
			t.Errorf("Sort = %q, want \"\"", filter.Sort)
		}
		if filter.Lat != nil {
			t.Errorf("Lat = %v, want nil", *filter.Lat)
		}
		if filter.Lng != nil {
			t.Errorf("Lng = %v, want nil", *filter.Lng)
		}
		if filter.RadiusKm != nil {
			t.Errorf("RadiusKm = %v, want nil", *filter.RadiusKm)
		}
		if filter.Limit != 0 {
			t.Errorf("Limit = %d, want 0", filter.Limit)
		}
		if filter.Cursor != nil {
			t.Errorf("Cursor = %v, want nil", *filter.Cursor)
		}
	})

	t.Run("неизвестный параметр игнорируется", func(t *testing.T) {
		values := url.Values{"foo": {"bar"}, "region": {"dushanbe"}}

		filter, err := parseListQuery(values)
		if err != nil {
			t.Fatalf("parseListQuery() error = %v, want nil", err)
		}

		if filter.Region == nil || *filter.Region != "dushanbe" {
			t.Errorf("Region = %v, want *\"dushanbe\"", filter.Region)
		}
	})
}

// assertInvalidField — двойная проверка: категория (errors.Is) и конкретное поле (errors.As + Fields).
func assertInvalidField(t *testing.T, err error, field string) {
	t.Helper()

	if err == nil {
		t.Fatal("err = nil, want error")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, want true; err=%v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err=%v", err)
	}
	if target.Fields[field] == "" {
		t.Errorf("target.Fields[%q] пусто, want непустое сообщение; Fields=%v", field, target.Fields)
	}
}

package http

import (
	"errors"
	"net/url"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func TestParseListQuery(t *testing.T) {
	t.Run("sort=hack невалидный enum", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"sort": {"hack"}})
		assertInvalidField(t, err, "sort")
	})

	t.Run("min_price=abc не парсится", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"min_price": {"abc"}})
		assertInvalidField(t, err, "min_price")
	})

	t.Run("max_price=xyz не парсится", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"max_price": {"xyz"}})
		assertInvalidField(t, err, "max_price")
	})

	t.Run("lat=200 вне диапазона", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"lat": {"200"}})
		assertInvalidField(t, err, "lat")
	})

	t.Run("lng=-200 вне диапазона", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"lng": {"-200"}})
		assertInvalidField(t, err, "lng")
	})

	t.Run("limit=-1 отрицательный", func(t *testing.T) {
		_, err := parseListQuery(url.Values{"limit": {"-1"}})
		assertInvalidField(t, err, "limit")
	})

	t.Run("type=cat_school,cat_uni разбивается по запятой", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"type": {"cat_school,cat_uni"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"cat_school", "cat_uni"}
		if len(f.Types) != len(want) || f.Types[0] != want[0] || f.Types[1] != want[1] {
			t.Fatalf("Types = %v, want %v", f.Types, want)
		}
	})

	t.Run("transport=1 парсится как true", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"transport": {"1"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Transport == nil || *f.Transport != true {
			t.Fatalf("Transport = %v, want *true", f.Transport)
		}
	})

	t.Run("transport= пустое значение остаётся nil, не false", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"transport": {""}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Transport != nil {
			t.Fatalf("Transport = %v, want nil", *f.Transport)
		}
	})

	t.Run("curriculum=bilingual,stem разбивается по запятой", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"curriculum": {"bilingual,stem"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"bilingual", "stem"}
		if len(f.Curriculum) != len(want) || f.Curriculum[0] != want[0] || f.Curriculum[1] != want[1] {
			t.Fatalf("Curriculum = %v, want %v", f.Curriculum, want)
		}
	})

	t.Run("discount=true парсится как true", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"discount": {"true"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Discount == nil || *f.Discount != true {
			t.Fatalf("Discount = %v, want *true", f.Discount)
		}
	})

	t.Run("q=гулистон сохраняется как есть", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"q": {"гулистон"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Q == nil || *f.Q != "гулистон" {
			t.Fatalf("Q = %v, want *\"гулистон\"", f.Q)
		}
	})

	t.Run("region=dushanbe единственный параметр, остальное нулевое", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"region": {"dushanbe"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Region == nil || *f.Region != "dushanbe" {
			t.Fatalf("Region = %v, want *\"dushanbe\"", f.Region)
		}
		if f.Q != nil || f.Area != nil || f.Types != nil || f.MinPrice != nil ||
			f.MaxPrice != nil || f.MinRating != nil || f.Transport != nil ||
			f.Food != nil || f.Verified != nil || f.Curriculum != nil ||
			f.ProgramLevel != nil || f.Discount != nil || f.Sort != "" ||
			f.Lat != nil || f.Lng != nil || f.RadiusKm != nil || f.Limit != 0 ||
			f.Cursor != nil {
			t.Fatalf("ожидались нулевые значения остальных полей, получили: %+v", f)
		}
	})

	t.Run("неизвестный параметр игнорируется", func(t *testing.T) {
		f, err := parseListQuery(url.Values{"foo": {"bar"}, "region": {"dushanbe"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Region == nil || *f.Region != "dushanbe" {
			t.Fatalf("Region = %v, want *\"dushanbe\"", f.Region)
		}
	})
}

// assertInvalidField проверяет, что err — apperr.Invalid с непустой причиной по конкретному полю.
func assertInvalidField(t *testing.T, err error, field string) {
	t.Helper()
	if err == nil {
		t.Fatal("ожидалась ошибка, получен nil")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("ошибка не относится к apperr.ErrInvalid: %v", err)
	}
	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As не смог извлечь *apperr.Error из %v", err)
	}
	if target.Fields[field] == "" {
		t.Fatalf("ожидалась причина по полю %q, Fields=%v", field, target.Fields)
	}
}

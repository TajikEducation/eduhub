package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// fakeListService — тестовая реализация listService: возвращает заранее заданный
// результат/ошибку, фиксирует факт вызова List (нужно для RED-кейса (е)).
type fakeListService struct {
	result  domain.ListResult
	err     error
	calledL bool
}

func (f *fakeListService) List(ctx context.Context, filter domain.Filter) (domain.ListResult, error) {
	f.calledL = true
	return f.result, f.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func floatPtr(v float64) *float64 { return &v }
func strPtr(v string) *string     { return &v }

// TestListHandler — unit-тесты ListHandler через httptest, без БД.
func TestListHandler(t *testing.T) {
	t.Run("успешный список сериализуется в items/next_cursor/total_hint:null", func(t *testing.T) {
		fake := &fakeListService{
			result: domain.ListResult{
				Items: []domain.Institution{
					{ID: uuid.New(), Name: domain.Bilingual{RU: "Сад №1", TG: "Боғи №1"}, Region: "dushanbe"},
					{ID: uuid.New(), Name: domain.Bilingual{RU: "Сад №2", TG: "Боғи №2"}, Region: "sughd"},
				},
				NextCursor: strPtr("abc123"),
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		body := rec.Body.String()
		if !strings.Contains(body, `"total_hint":null`) {
			t.Errorf("тело не содержит %q: body=%s", `"total_hint":null`, body)
		}

		var parsed listResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal error = %v; body=%s", err, body)
		}
		if len(parsed.Items) != 2 {
			t.Errorf("len(items) = %d, want 2", len(parsed.Items))
		}
		if parsed.NextCursor == nil || *parsed.NextCursor != "abc123" {
			t.Errorf("next_cursor = %v, want %q", parsed.NextCursor, "abc123")
		}
	})

	t.Run("пустой список сериализуется как items:[] не items:null", func(t *testing.T) {
		fake := &fakeListService{
			result: domain.ListResult{Items: []domain.Institution{}, NextCursor: nil},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		body := rec.Body.String()
		if !strings.Contains(body, `"items":[]`) {
			t.Errorf("тело не содержит %q: body=%s", `"items":[]`, body)
		}
		if strings.Contains(body, `"items":null`) {
			t.Errorf("тело содержит %q, не должно: body=%s", `"items":null`, body)
		}
	})

	t.Run("distance_m присутствует только когда задан в домене", func(t *testing.T) {
		fake := &fakeListService{
			result: domain.ListResult{
				Items: []domain.Institution{
					{ID: uuid.New(), Name: domain.Bilingual{RU: "С дистанцией", TG: "Бо масофа"}, Region: "dushanbe", DistanceM: floatPtr(123.45)},
					{ID: uuid.New(), Name: domain.Bilingual{RU: "Без дистанции", TG: "Бе масофа"}, Region: "sughd"},
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var raw struct {
			Items []map[string]any `json:"items"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatalf("json.Unmarshal error = %v; body=%s", err, rec.Body.String())
		}
		if len(raw.Items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(raw.Items))
		}

		distVal, ok := raw.Items[0]["distance_m"]
		if !ok {
			t.Error("items[0] не содержит ключ distance_m, а должен")
		}
		if distVal != 123.45 {
			t.Errorf("items[0].distance_m = %v, want 123.45", distVal)
		}

		if _, ok := raw.Items[1]["distance_m"]; ok {
			t.Error("items[1] содержит ключ distance_m, а не должен (nil в домене)")
		}
	})

	t.Run("ошибка сервиса маппится через httpx.WriteError", func(t *testing.T) {
		fake := &fakeListService{err: apperr.NotFound("institution", "irrelevant")}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
		}

		var parsed struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal error = %v; body=%s", err, rec.Body.String())
		}
		if parsed.Error.Code == "" {
			t.Error("error.code пустой, ожидался непустой код")
		}
	})

	t.Run("успешный ответ содержит Cache-Control", func(t *testing.T) {
		fake := &fakeListService{result: domain.ListResult{Items: []domain.Institution{}}}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=60" {
			t.Errorf("Cache-Control = %q, want %q", got, "public, max-age=60")
		}
	})

	t.Run("невалидный диапазон цены отклоняется до вызова сервиса", func(t *testing.T) {
		fake := &fakeListService{}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions?min_price=1000&max_price=100", nil)
		rec := httptest.NewRecorder()

		ListHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		if fake.calledL {
			t.Error("List() не должен был вызываться при невалидном фильтре")
		}
	})
}

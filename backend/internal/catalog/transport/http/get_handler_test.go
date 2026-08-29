package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// fakeGetService — тестовая реализация getService: возвращает заранее заданный результат/ошибку.
type fakeGetService struct {
	inst domain.Institution
	err  error
}

func (f *fakeGetService) Get(ctx context.Context, id uuid.UUID) (domain.Institution, error) {
	return f.inst, f.err
}

// TestGetHandler — unit-тесты GetHandler через httptest, без БД.
func TestGetHandler(t *testing.T) {
	t.Run("валидный id возвращает 200 с телом и ETag", func(t *testing.T) {
		id := uuid.New()
		fake := &fakeGetService{
			inst: domain.Institution{
				ID:        id,
				Name:      domain.Bilingual{RU: "Сад №1", TG: "Боғи №1"},
				Region:    "dushanbe",
				UpdatedAt: time.Now(),
				Staff: []domain.StaffMember{
					{ID: uuid.New(), Name: domain.Bilingual{RU: "Иванова", TG: "Иванова"}, RoleType: "teacher"},
					{ID: uuid.New(), Name: domain.Bilingual{RU: "Петров", TG: "Петров"}, RoleType: "teacher"},
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
		req.SetPathValue("id", id.String())
		rec := httptest.NewRecorder()

		GetHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var parsed institutionDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal error = %v; body=%s", err, rec.Body.String())
		}
		if parsed.ID != id {
			t.Errorf("id = %v, want %v", parsed.ID, id)
		}
		if len(parsed.Staff) != 2 {
			t.Errorf("len(staff) = %d, want 2", len(parsed.Staff))
		}

		if got := rec.Header().Get("ETag"); got == "" {
			t.Error("ETag заголовок отсутствует или пуст")
		}
	})

	t.Run("невалидный id возвращает 400", func(t *testing.T) {
		fake := &fakeGetService{}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/not-a-uuid", nil)
		req.SetPathValue("id", "not-a-uuid")
		rec := httptest.NewRecorder()

		GetHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("повторный запрос с If-None-Match возвращает 304 с пустым телом", func(t *testing.T) {
		id := uuid.New()
		fake := &fakeGetService{
			inst: domain.Institution{
				ID:        id,
				Name:      domain.Bilingual{RU: "Сад №1", TG: "Боғи №1"},
				Region:    "dushanbe",
				UpdatedAt: time.Now(),
			},
		}

		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
		req1.SetPathValue("id", id.String())
		rec1 := httptest.NewRecorder()
		GetHandler(fake, testLogger()).ServeHTTP(rec1, req1)

		if rec1.Code != http.StatusOK {
			t.Fatalf("первый запрос: status = %d, want %d; body=%s", rec1.Code, http.StatusOK, rec1.Body.String())
		}
		etag := rec1.Header().Get("ETag")
		if etag == "" {
			t.Fatal("первый запрос: ETag заголовок отсутствует или пуст")
		}

		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
		req2.SetPathValue("id", id.String())
		req2.Header.Set("If-None-Match", etag)
		rec2 := httptest.NewRecorder()
		GetHandler(fake, testLogger()).ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusNotModified {
			t.Fatalf("второй запрос: status = %d, want %d; body=%s", rec2.Code, http.StatusNotModified, rec2.Body.String())
		}
		if rec2.Body.Len() != 0 {
			t.Errorf("второй запрос: тело не пустое, len=%d, body=%s", rec2.Body.Len(), rec2.Body.String())
		}
	})

	t.Run("ошибка сервиса NotFound маппится через httpx.WriteError в 404", func(t *testing.T) {
		fake := &fakeGetService{err: apperr.NotFound("institution", "irrelevant")}

		id := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
		req.SetPathValue("id", id.String())
		rec := httptest.NewRecorder()

		GetHandler(fake, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}

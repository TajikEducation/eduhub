package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// fakeCatalogRepo — дублёр InstitutionRepo для сквозного smoke-теста: не трогает реальную БД,
// просто отдаёт заранее подготовленные значения.
type fakeCatalogRepo struct {
	listResult domain.ListResult
	getInst    domain.Institution
	getErr     error
}

func (f *fakeCatalogRepo) List(ctx context.Context, filter domain.Filter) (domain.ListResult, error) {
	return f.listResult, nil
}

func (f *fakeCatalogRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Institution, error) {
	return f.getInst, f.getErr
}

func (f *fakeCatalogRepo) Create(_ context.Context, inst domain.Institution, _ uuid.UUID) (domain.Institution, error) {
	return inst, nil
}

func (f *fakeCatalogRepo) Update(_ context.Context, _ uuid.UUID, _ domain.UpdateInstitutionInput) (domain.Institution, error) {
	return domain.Institution{}, nil
}

func (f *fakeCatalogRepo) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (f *fakeCatalogRepo) UpdateRatingAvg(_ context.Context, _ uuid.UUID, _ float64, _ int) error {
	return nil
}

func (f *fakeCatalogRepo) ListByOwner(_ context.Context, _ uuid.UUID) ([]domain.Institution, error) {
	return nil, nil
}

func (f *fakeCatalogRepo) GetOwnerID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}

func (f *fakeCatalogRepo) SetModerationStatus(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (f *fakeCatalogRepo) CreateStaff(_ context.Context, _ uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	return domain.StaffMember{Name: in.Name, RoleType: in.RoleType, RoleLabel: in.RoleLabel}, nil
}
func (f *fakeCatalogRepo) UpdateStaff(_ context.Context, id uuid.UUID, _ domain.CreateStaffInput) (domain.StaffMember, error) {
	return domain.StaffMember{ID: id}, nil
}
func (f *fakeCatalogRepo) DeleteStaff(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeCatalogRepo) GetStaffInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
func (f *fakeCatalogRepo) GetPublicStaffByID(_ context.Context, _ uuid.UUID) (domain.StaffMember, error) {
	return domain.StaffMember{}, nil
}
func (f *fakeCatalogRepo) CreateAchievement(_ context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error) {
	return domain.Achievement{OwnerID: institutionID, Title: in.Title}, nil
}
func (f *fakeCatalogRepo) DeleteAchievement(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeCatalogRepo) GetAchievementInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
func (f *fakeCatalogRepo) CreateGalleryItem(_ context.Context, _ uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error) {
	return domain.GalleryItem{S3Key: in.S3Key}, nil
}
func (f *fakeCatalogRepo) DeleteGalleryItem(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeCatalogRepo) GetGalleryItemInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
func (f *fakeCatalogRepo) CreateAlumnus(_ context.Context, _ uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error) {
	return domain.Alumnus{Name: in.Name}, nil
}
func (f *fakeCatalogRepo) DeleteAlumnus(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeCatalogRepo) GetAlumnusInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
func (f *fakeCatalogRepo) CreateNews(_ context.Context, _ uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	return domain.NewsArticle{Title: in.Title}, nil
}
func (f *fakeCatalogRepo) UpdateNews(_ context.Context, id uuid.UUID, _ domain.CreateNewsInput) (domain.NewsArticle, error) {
	return domain.NewsArticle{ID: id}, nil
}
func (f *fakeCatalogRepo) DeleteNews(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeCatalogRepo) GetNewsInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}

func (f *fakeCatalogRepo) ListNews(_ context.Context, _ uuid.UUID) ([]domain.NewsArticle, error) {
	return nil, nil
}

func (f *fakeCatalogRepo) ListPublishedNews(_ context.Context, _ uuid.UUID) ([]domain.NewsArticle, error) {
	return nil, nil
}

func (f *fakeCatalogRepo) GetPublishedNewsByID(_ context.Context, _ uuid.UUID) (domain.NewsArticle, error) {
	return domain.NewsArticle{}, nil
}

// doGet — обёртка над http.NewRequestWithContext+Do: линтер (noctx) требует контекст
// на каждом внешнем вызове, простой http.Get его не пробрасывает.
func doGet(ctx context.Context, t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	return http.DefaultClient.Do(req)
}

// TestSmoke_CatalogRoutesThroughRealServer поднимает реальный HTTP-сервер (через run(), как
// в run_test.go) и делает настоящие сетевые запросы к маршрутам каталога — проверяет, что
// newHandler правильно собирает роутер + middleware-цепочку + usecase поверх фейкового репозитория.
func TestSmoke_CatalogRoutesThroughRealServer(t *testing.T) {
	log := logger.New("info", "test", io.Discard)

	someID := uuid.New()
	fakeRepo := &fakeCatalogRepo{
		listResult: domain.ListResult{
			Items: []domain.Institution{
				{ID: someID, Name: domain.Bilingual{RU: "Сад №1", TG: "Боғи №1"}, Region: "dushanbe", ModerationStatus: "approved"},
			},
		},
		getInst: domain.Institution{ID: someID, Name: domain.Bilingual{RU: "Сад №1", TG: "Боғи №1"}, Region: "dushanbe", ModerationStatus: "approved"},
	}

	handler := newHandler(log, nil, nil, fakeRepo, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil)

	cfg := config.Config{HTTPAddr: ":0", ShutdownTimeout: 2 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrCh := make(chan string, 1)
	deps := Deps{Logger: log, Pool: &fakePinger{}, Handler: handler, Ready: func(addr string) { addrCh <- addr }}

	errCh := make(chan error, 1)
	go func() { errCh <- run(ctx, cfg, deps) }()

	var addr string
	select {
	case addr = <-addrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server address")
	}

	t.Run("GET /api/v1/institutions возвращает список и X-Request-ID", func(t *testing.T) {
		resp, err := doGet(ctx, t, "http://"+addr+"/api/v1/institutions")
		if err != nil {
			t.Fatalf("GET /api/v1/institutions: %v", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("resp.Body.Close: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatal("expected non-empty X-Request-ID header")
		}

		var body struct {
			Items []struct {
				ID uuid.UUID `json:"id"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(body.Items) != 1 {
			t.Fatalf("items count = %d, want 1", len(body.Items))
		}
		if body.Items[0].ID != someID {
			t.Fatalf("items[0].id = %s, want %s", body.Items[0].ID, someID)
		}
	})

	t.Run("GET /api/v1/institutions/{id} возвращает карточку", func(t *testing.T) {
		resp, err := doGet(ctx, t, "http://"+addr+"/api/v1/institutions/"+someID.String())
		if err != nil {
			t.Fatalf("GET /api/v1/institutions/{id}: %v", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("resp.Body.Close: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var body struct {
			ID uuid.UUID `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.ID != someID {
			t.Fatalf("id = %s, want %s", body.ID, someID)
		}
	})

	t.Run("неизвестный путь возвращает 404 в JSON", func(t *testing.T) {
		resp, err := doGet(ctx, t, "http://"+addr+"/unknown/nowhere")
		if err != nil {
			t.Fatalf("GET /unknown/nowhere: %v", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Errorf("resp.Body.Close: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
		if ct := resp.Header.Get("Content-Type"); ct == "" {
			t.Fatal("expected non-empty Content-Type header")
		}

		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		errField, ok := body["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected body[\"error\"] to be an object, got %#v", body["error"])
		}
		if code, _ := errField["code"].(string); code == "" {
			t.Fatal("expected non-empty error.code")
		}
	})

	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for graceful shutdown")
	}
}

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

type fakeCreateService struct {
	gotOwnerID uuid.UUID
	gotInput   domain.CreateInstitutionInput
	inst       domain.Institution
	err        error
}

func (f *fakeCreateService) Register(_ context.Context, ownerID uuid.UUID, in domain.CreateInstitutionInput) (domain.Institution, error) {
	f.gotOwnerID, f.gotInput = ownerID, in
	return f.inst, f.err
}

func TestCreateHandler_WithoutPrincipalReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeCreateService{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	CreateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCreateHandler_ValidBodyReturns201WithOwnerFromPrincipal(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	ownerID := uuid.New()
	instID := uuid.New()
	fake := &fakeCreateService{inst: domain.Institution{ID: instID, Name: domain.Bilingual{RU: "Сад", TG: "Боғча"}, Region: "dushanbe"}}

	body := bytes.NewBufferString(`{"name":{"ru":"Сад","tg":"Боғча"},"types":["kindergarten"],"region":"dushanbe","lat":38.5,"lng":68.7}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", body)
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: ownerID, Role: "user"}))
	rec := httptest.NewRecorder()

	CreateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if fake.gotOwnerID != ownerID {
		t.Errorf("Register() called with ownerID = %v, want %v", fake.gotOwnerID, ownerID)
	}
	if fake.gotInput.Name.RU != "Сад" {
		t.Errorf("Register() called with Name.RU = %q, want %q", fake.gotInput.Name.RU, "Сад")
	}

	var resp institutionDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != instID {
		t.Errorf("resp.ID = %v, want %v", resp.ID, instID)
	}
}

func TestCreateHandler_ServiceErrorPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeCreateService{err: apperr.Invalid(map[string]string{"region": "bad"}, "invalid")}

	body := bytes.NewBufferString(`{"name":{"ru":"Сад","tg":"Боғча"},"types":["kindergarten"],"region":"dushanbe"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", body)
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "user"}))
	rec := httptest.NewRecorder()

	CreateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

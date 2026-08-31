package http

import (
	"bytes"
	"context"
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

type fakeUpdateService struct {
	ownerOf  uuid.UUID
	isOwner  bool
	ownerErr error

	gotID    uuid.UUID
	gotPatch domain.UpdateInstitutionInput
	inst     domain.Institution
	err      error
}

func (f *fakeUpdateService) IsOwner(_ context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	if f.ownerErr != nil {
		return false, f.ownerErr
	}
	return f.isOwner, nil
}

func (f *fakeUpdateService) Update(_ context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error) {
	f.gotID, f.gotPatch = id, patch
	return f.inst, f.err
}

func newUpdateRequest(id uuid.UUID, principal rbac.Principal, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/institutions/"+id.String(), bytes.NewBufferString(body))
	req.SetPathValue("id", id.String())
	return req.WithContext(rbac.NewContext(req.Context(), principal))
}

func TestUpdateHandler_WithoutPrincipalReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeUpdateService{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/institutions/"+uuid.New().String(), bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdateHandler_OwnerCanUpdate(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	userID := uuid.New()
	fake := &fakeUpdateService{isOwner: true, inst: domain.Institution{ID: id}}

	req := newUpdateRequest(id, rbac.Principal{UserID: userID, Role: "user"}, `{"price":700}`)
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if fake.gotPatch.Price == nil || *fake.gotPatch.Price != 700 {
		t.Errorf("patch.Price = %v, want 700", fake.gotPatch.Price)
	}
}

func TestUpdateHandler_NonOwnerReturns403(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	fake := &fakeUpdateService{isOwner: false}

	req := newUpdateRequest(id, rbac.Principal{UserID: uuid.New(), Role: "user"}, `{"price":700}`)
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUpdateHandler_ModeratorCanUpdateWithoutOwnership(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	fake := &fakeUpdateService{isOwner: false, inst: domain.Institution{ID: id}}

	req := newUpdateRequest(id, rbac.Principal{UserID: uuid.New(), Role: "moderator"}, `{"price":700}`)
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateHandler_InvalidIDReturns400(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeUpdateService{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/institutions/not-a-uuid", bytes.NewBufferString(`{}`))
	req.SetPathValue("id", "not-a-uuid")
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "moderator"}))
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateHandler_ServiceErrorPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	fake := &fakeUpdateService{isOwner: true, err: apperr.NotFound("institution", id.String())}

	req := newUpdateRequest(id, rbac.Principal{UserID: uuid.New(), Role: "user"}, `{"price":700}`)
	rec := httptest.NewRecorder()

	UpdateHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

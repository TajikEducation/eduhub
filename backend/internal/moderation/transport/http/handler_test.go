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
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

type fakeModerationService struct {
	approveActor rbac.Principal
	approveID    uuid.UUID
	approveErr   error

	rejectActor      rbac.Principal
	rejectID         uuid.UUID
	rejectReasonCode string
	rejectReasonText string
	rejectErr        error
}

func (f *fakeModerationService) ApproveInstitution(_ context.Context, actor rbac.Principal, _ string, institutionID uuid.UUID) error {
	f.approveActor, f.approveID = actor, institutionID
	return f.approveErr
}

func (f *fakeModerationService) RejectInstitution(_ context.Context, actor rbac.Principal, _ string, institutionID uuid.UUID, reasonCode, reasonText string) error {
	f.rejectActor, f.rejectID, f.rejectReasonCode, f.rejectReasonText = actor, institutionID, reasonCode, reasonText
	return f.rejectErr
}

func (f *fakeModerationService) ApproveReview(_ context.Context, actor rbac.Principal, _ string, reviewID uuid.UUID) error {
	f.approveActor, f.approveID = actor, reviewID
	return f.approveErr
}

func (f *fakeModerationService) RejectReview(_ context.Context, actor rbac.Principal, _ string, reviewID uuid.UUID, reasonCode, reasonText string) error {
	f.rejectActor, f.rejectID, f.rejectReasonCode, f.rejectReasonText = actor, reviewID, reasonCode, reasonText
	return f.rejectErr
}

func TestApproveHandler_WithoutPrincipalReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeModerationService{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/institutions/"+uuid.New().String()+"/approve", nil)
	rec := httptest.NewRecorder()

	ApproveHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestApproveHandler_ValidRequestReturns204(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeModerationService{}
	id := uuid.New()
	actor := rbac.Principal{UserID: uuid.New(), Role: "moderator"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/institutions/"+id.String()+"/approve", nil)
	req.SetPathValue("id", id.String())
	req = req.WithContext(rbac.NewContext(req.Context(), actor))
	rec := httptest.NewRecorder()

	ApproveHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if fake.approveID != id {
		t.Errorf("ApproveInstitution() called with id = %v, want %v", fake.approveID, id)
	}
	if fake.approveActor.UserID != actor.UserID {
		t.Errorf("ApproveInstitution() called with actor = %v, want %v", fake.approveActor, actor)
	}
}

func TestApproveHandler_ServiceErrorPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	fake := &fakeModerationService{approveErr: apperr.NotFound("institution", id.String())}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/institutions/"+id.String()+"/approve", nil)
	req.SetPathValue("id", id.String())
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "admin"}))
	rec := httptest.NewRecorder()

	ApproveHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRejectHandler_ValidRequestReturns204(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeModerationService{}
	id := uuid.New()

	body := bytes.NewBufferString(`{"reason_code":"incomplete_docs","reason_text":"Не хватает лицензии"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/institutions/"+id.String()+"/reject", body)
	req.SetPathValue("id", id.String())
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "admin"}))
	rec := httptest.NewRecorder()

	RejectHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if fake.rejectReasonText != "Не хватает лицензии" {
		t.Errorf("rejectReasonText = %q, want %q", fake.rejectReasonText, "Не хватает лицензии")
	}
}

func TestRejectHandler_ServiceErrorPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	id := uuid.New()
	fake := &fakeModerationService{rejectErr: apperr.Invalid(map[string]string{"reason_text": "обязательное поле"}, "reject требует причину")}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/institutions/"+id.String()+"/reject", bytes.NewBufferString(`{}`))
	req.SetPathValue("id", id.String())
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "admin"}))
	rec := httptest.NewRecorder()

	RejectHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

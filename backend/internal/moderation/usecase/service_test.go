package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	moddomain "github.com/abdulhalim/eduhub/backend/internal/moderation/domain"
	"github.com/abdulhalim/eduhub/backend/internal/moderation/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

type fakeInstitutions struct {
	id     uuid.UUID
	status string
	err    error
}

func (f *fakeInstitutions) SetModerationStatus(_ context.Context, id uuid.UUID, status string) error {
	f.id, f.status = id, status
	return f.err
}

type fakeRecorder struct {
	entry moddomain.Entry
	err   error
}

func (f *fakeRecorder) Record(_ context.Context, e moddomain.Entry) error {
	f.entry = e
	return f.err
}

func TestService_ApproveInstitution_SetsStatusAndRecordsAudit(t *testing.T) {
	institutions := &fakeInstitutions{}
	audit := &fakeRecorder{}
	svc := usecase.New(institutions, audit, nil)

	actor := rbac.Principal{UserID: uuid.New(), Role: "moderator"}
	institutionID := uuid.New()

	if err := svc.ApproveInstitution(context.Background(), actor, "req-1", institutionID); err != nil {
		t.Fatalf("ApproveInstitution() unexpected error: %v", err)
	}

	if institutions.id != institutionID || institutions.status != "approved" {
		t.Errorf("institutions received (%v, %q), want (%v, %q)", institutions.id, institutions.status, institutionID, "approved")
	}
	if audit.entry.Action != "approve_institution" {
		t.Errorf("audit.Action = %q, want %q", audit.entry.Action, "approve_institution")
	}
	if audit.entry.ActorID == nil || *audit.entry.ActorID != actor.UserID {
		t.Errorf("audit.ActorID = %v, want %v", audit.entry.ActorID, actor.UserID)
	}
	if audit.entry.TargetID != institutionID {
		t.Errorf("audit.TargetID = %v, want %v", audit.entry.TargetID, institutionID)
	}
}

func TestService_RejectInstitution_RequiresReasonText(t *testing.T) {
	institutions := &fakeInstitutions{}
	audit := &fakeRecorder{}
	svc := usecase.New(institutions, audit, nil)

	err := svc.RejectInstitution(context.Background(), rbac.Principal{UserID: uuid.New(), Role: "admin"}, "req-1", uuid.New(), "code", "")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("RejectInstitution() error = %v, want apperr.ErrInvalid", err)
	}
	if institutions.status != "" {
		t.Error("RejectInstitution() without reason must not touch institutions repo")
	}
}

func TestService_RejectInstitution_SetsStatusAndRecordsAudit(t *testing.T) {
	institutions := &fakeInstitutions{}
	audit := &fakeRecorder{}
	svc := usecase.New(institutions, audit, nil)

	institutionID := uuid.New()
	if err := svc.RejectInstitution(context.Background(), rbac.Principal{UserID: uuid.New(), Role: "admin"}, "req-1", institutionID, "incomplete_docs", "Не хватает лицензии"); err != nil {
		t.Fatalf("RejectInstitution() unexpected error: %v", err)
	}

	if institutions.status != "rejected" {
		t.Errorf("institutions.status = %q, want %q", institutions.status, "rejected")
	}
	if audit.entry.ReasonText == nil || *audit.entry.ReasonText != "Не хватает лицензии" {
		t.Errorf("audit.ReasonText = %v, want %q", audit.entry.ReasonText, "Не хватает лицензии")
	}
	if audit.entry.ReasonCode == nil || *audit.entry.ReasonCode != "incomplete_docs" {
		t.Errorf("audit.ReasonCode = %v, want %q", audit.entry.ReasonCode, "incomplete_docs")
	}
}

func TestService_ApproveInstitution_StatusErrorSkipsAudit(t *testing.T) {
	institutions := &fakeInstitutions{err: apperr.NotFound("institution", "x")}
	audit := &fakeRecorder{}
	svc := usecase.New(institutions, audit, nil)

	err := svc.ApproveInstitution(context.Background(), rbac.Principal{UserID: uuid.New(), Role: "admin"}, "req-1", uuid.New())
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("ApproveInstitution() error = %v, want apperr.ErrNotFound", err)
	}
	if audit.entry.Action != "" {
		t.Error("ApproveInstitution() must not record audit when status update fails")
	}
}

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// fakeChildRepo — тестовый двойник usecase.ChildRepo.
type fakeChildRepo struct {
	createErr error
	created   []domain.Child

	listPendingResult []domain.Child
	listPendingErr    error

	confirmResult domain.Child
	confirmErr    error
	confirmCalls  []confirmCall

	rejectResult domain.Child
	rejectErr    error
	rejectCalls  []rejectCall
}

// confirmCall/rejectCall — захваченные аргументы вызова, чтобы тесты могли проверить, что
// usecase-слой прокидывает actorID/actorRole/requestID без подмены (не просто компилируется —
// оба ведущих параметра одного типа uuid.UUID, порядок легко перепутать местами).
type confirmCall struct {
	childID, actorID uuid.UUID
	actorRole, reqID string
}

type rejectCall struct {
	childID, actorID             uuid.UUID
	actorRole, reasonCode, reqID string
	reasonText                   *string
}

func (f *fakeChildRepo) Confirm(_ context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error) {
	f.confirmCalls = append(f.confirmCalls, confirmCall{childID: childID, actorID: actorID, actorRole: actorRole, reqID: requestID})
	if f.confirmErr != nil {
		return domain.Child{}, f.confirmErr
	}
	return f.confirmResult, nil
}

func (f *fakeChildRepo) Reject(_ context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error) {
	f.rejectCalls = append(f.rejectCalls, rejectCall{
		childID: childID, actorID: actorID, actorRole: actorRole,
		reasonCode: reasonCode, reasonText: reasonText, reqID: requestID,
	})
	if f.rejectErr != nil {
		return domain.Child{}, f.rejectErr
	}
	return f.rejectResult, nil
}

func (f *fakeChildRepo) Create(_ context.Context, c domain.Child) (domain.Child, error) {
	if f.createErr != nil {
		return domain.Child{}, f.createErr
	}
	f.created = append(f.created, c)
	return c, nil
}

func (f *fakeChildRepo) ListPendingByInstitution(_ context.Context, _ uuid.UUID) ([]domain.Child, error) {
	if f.listPendingErr != nil {
		return nil, f.listPendingErr
	}
	return f.listPendingResult, nil
}

// fakeInstitutionStatusChecker — тестовый двойник usecase.InstitutionStatusChecker.
type fakeInstitutionStatusChecker struct {
	approved bool
	err      error
}

func (f *fakeInstitutionStatusChecker) IsApproved(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.approved, f.err
}

func TestCreateChild_ApprovedInstitution_Success(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	children := &fakeChildRepo{}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	userID := uuid.New()
	institutionID := uuid.New()

	got, err := svc.CreateChild(context.Background(), userID, institutionID, "primary", "current")
	if err != nil {
		t.Fatalf("CreateChild() вернул ошибку: %v", err)
	}

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.InstitutionID != institutionID {
		t.Errorf("InstitutionID = %v, want %v", got.InstitutionID, institutionID)
	}
	if got.AgeGroup != "primary" {
		t.Errorf("AgeGroup = %q, want %q", got.AgeGroup, "primary")
	}
	if got.Status != "current" {
		t.Errorf("Status = %q, want %q", got.Status, "current")
	}
	if got.ConfirmationStatus != "pending" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "pending")
	}
	if !got.CreatedAt.Equal(start) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, start)
	}
	if got.ID == uuid.Nil {
		t.Error("ID пуст")
	}
}

func TestCreateChild_InstitutionNotApproved_ReturnsConflict(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{}
	institutions := &fakeInstitutionStatusChecker{approved: false}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.CreateChild(context.Background(), uuid.New(), uuid.New(), "primary", "current")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "institution_not_approved" {
		t.Errorf("Code() = %q, want %q", target.Code(), "institution_not_approved")
	}
	if len(children.created) != 0 {
		t.Errorf("children.Create() вызван, хотя учреждение не approved")
	}
}

func TestCreateChild_InstitutionNotFound_PropagatesError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{}
	institutions := &fakeInstitutionStatusChecker{err: apperr.NotFound("institution", "irrelevant")}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.CreateChild(context.Background(), uuid.New(), uuid.New(), "primary", "current")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestCreateChild_RepoConflict_PropagatesError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{createErr: apperr.ConflictCode("child_link_exists", "привязка уже существует")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.CreateChild(context.Background(), uuid.New(), uuid.New(), "primary", "current")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_link_exists" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_link_exists")
	}
}

func TestListPendingByInstitution_Success(t *testing.T) {
	clk := clock.NewFake(time.Now())
	institutionID := uuid.New()
	want := []domain.Child{{ID: uuid.New(), InstitutionID: institutionID, ConfirmationStatus: "pending"}}
	children := &fakeChildRepo{listPendingResult: want}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	got, err := svc.ListPendingByInstitution(context.Background(), institutionID)
	if err != nil {
		t.Fatalf("ListPendingByInstitution() вернул ошибку: %v", err)
	}
	wantID := want[0].ID
	if len(got) != 1 || got[0].ID != wantID {
		t.Errorf("ListPendingByInstitution() = %v, want %v", got, want)
	}
}

func TestListPendingByInstitution_RepoError_PropagatesError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{listPendingErr: errors.New("db error")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.ListPendingByInstitution(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("ListPendingByInstitution() не вернул ошибку")
	}
}

func TestConfirmChild_Success(t *testing.T) {
	clk := clock.NewFake(time.Now())
	want := domain.Child{ID: uuid.New(), ConfirmationStatus: "confirmed"}
	children := &fakeChildRepo{confirmResult: want}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	got, err := svc.ConfirmChild(context.Background(), want.ID, uuid.New(), "moderator", "req-1")
	if err != nil {
		t.Fatalf("ConfirmChild() вернул ошибку: %v", err)
	}
	if got.ConfirmationStatus != "confirmed" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "confirmed")
	}
}

// TestConfirmChild_PassesArgsThrough защищает от перепутанных местами childID/actorID —
// оба параметра одного типа uuid.UUID, компилятор такую подмену не поймает.
func TestConfirmChild_PassesArgsThrough(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	childID, actorID := uuid.New(), uuid.New()
	if _, err := svc.ConfirmChild(context.Background(), childID, actorID, "moderator", "req-1"); err != nil {
		t.Fatalf("ConfirmChild() вернул ошибку: %v", err)
	}

	if len(children.confirmCalls) != 1 {
		t.Fatalf("Confirm() вызван %d раз, want 1", len(children.confirmCalls))
	}
	call := children.confirmCalls[0]
	if call.childID != childID {
		t.Errorf("childID = %v, want %v", call.childID, childID)
	}
	if call.actorID != actorID {
		t.Errorf("actorID = %v, want %v", call.actorID, actorID)
	}
	if call.actorRole != "moderator" {
		t.Errorf("actorRole = %q, want %q", call.actorRole, "moderator")
	}
	if call.reqID != "req-1" {
		t.Errorf("requestID = %q, want %q", call.reqID, "req-1")
	}
}

func TestConfirmChild_NotPending_ReturnsConflict(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{confirmErr: apperr.ConflictCode("child_not_pending", "привязка уже обработана")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.ConfirmChild(context.Background(), uuid.New(), uuid.New(), "moderator", "req-1")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_not_pending" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_not_pending")
	}
}

func TestConfirmChild_NotFound_PropagatesError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{confirmErr: apperr.NotFound("child", "irrelevant")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.ConfirmChild(context.Background(), uuid.New(), uuid.New(), "moderator", "req-1")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

// TestRejectChild_PassesArgsThrough защищает от перепутанных местами childID/actorID — оба
// параметра одного типа uuid.UUID, компилятор такую подмену не поймает.
func TestRejectChild_PassesArgsThrough(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	childID, actorID := uuid.New(), uuid.New()
	reasonText := "документ не подтверждает обучение"
	if _, err := svc.RejectChild(context.Background(), childID, actorID, "admin", "invalid_document", &reasonText, "req-2"); err != nil {
		t.Fatalf("RejectChild() вернул ошибку: %v", err)
	}

	if len(children.rejectCalls) != 1 {
		t.Fatalf("Reject() вызван %d раз, want 1", len(children.rejectCalls))
	}
	call := children.rejectCalls[0]
	if call.childID != childID {
		t.Errorf("childID = %v, want %v", call.childID, childID)
	}
	if call.actorID != actorID {
		t.Errorf("actorID = %v, want %v", call.actorID, actorID)
	}
	if call.actorRole != "admin" {
		t.Errorf("actorRole = %q, want %q", call.actorRole, "admin")
	}
	if call.reasonCode != "invalid_document" {
		t.Errorf("reasonCode = %q, want %q", call.reasonCode, "invalid_document")
	}
	if call.reasonText == nil || *call.reasonText != reasonText {
		t.Errorf("reasonText = %v, want %q", call.reasonText, reasonText)
	}
	if call.reqID != "req-2" {
		t.Errorf("requestID = %q, want %q", call.reqID, "req-2")
	}
}

func TestRejectChild_Success(t *testing.T) {
	clk := clock.NewFake(time.Now())
	want := domain.Child{ID: uuid.New(), ConfirmationStatus: "rejected"}
	children := &fakeChildRepo{rejectResult: want}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	reasonText := "документ не подтверждает обучение"
	got, err := svc.RejectChild(context.Background(), want.ID, uuid.New(), "moderator", "invalid_document", &reasonText, "req-1")
	if err != nil {
		t.Fatalf("RejectChild() вернул ошибку: %v", err)
	}
	if got.ConfirmationStatus != "rejected" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "rejected")
	}
}

func TestRejectChild_NotPending_ReturnsConflict(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{rejectErr: apperr.ConflictCode("child_not_pending", "привязка уже обработана")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.RejectChild(context.Background(), uuid.New(), uuid.New(), "moderator", "invalid_document", nil, "req-1")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_not_pending" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_not_pending")
	}
}

func TestRejectChild_NotFound_PropagatesError(t *testing.T) {
	clk := clock.NewFake(time.Now())
	children := &fakeChildRepo{rejectErr: apperr.NotFound("child", "irrelevant")}
	institutions := &fakeInstitutionStatusChecker{approved: true}
	svc := usecase.NewChildService(children, institutions, clk)

	_, err := svc.RejectChild(context.Background(), uuid.New(), uuid.New(), "moderator", "invalid_document", nil, "req-1")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

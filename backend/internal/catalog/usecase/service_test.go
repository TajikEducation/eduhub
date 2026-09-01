package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// fakeRepo — тестовый двойник usecase.InstitutionRepo, фиксирующий последние
// полученные аргументы и отдающий настраиваемые результаты/ошибки.
type fakeRepo struct {
	listFilter domain.Filter
	listCtx    context.Context
	listResult domain.ListResult
	listErr    error

	getByIDInst domain.Institution
	getByIDErr  error
}

func (f *fakeRepo) List(ctx context.Context, filter domain.Filter) (domain.ListResult, error) {
	f.listFilter = filter
	f.listCtx = ctx
	return f.listResult, f.listErr
}

func (f *fakeRepo) GetByID(_ context.Context, _ uuid.UUID) (domain.Institution, error) {
	return f.getByIDInst, f.getByIDErr
}

// IsApproved не используется тестами в этом файле — реализован только чтобы fakeRepo продолжал
// удовлетворять usecase.InstitutionRepo после добавления метода в интерфейс (E2.6).
func (f *fakeRepo) IsApproved(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func TestService_List_ForcesApprovedStatus(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)

	_, err := svc.List(context.Background(), domain.Filter{Statuses: []string{"pending", "rejected"}})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if got := fake.listFilter.Statuses; len(got) != 1 || got[0] != "approved" {
		t.Fatalf("listFilter.Statuses = %v, want [approved]", got)
	}
}

func TestService_List_NormalizesBeforeCallingRepo(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)

	_, err := svc.List(context.Background(), domain.Filter{Limit: 0})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if fake.listFilter.Limit != 20 {
		t.Fatalf("listFilter.Limit = %d, want 20", fake.listFilter.Limit)
	}
}

func TestService_List_WrapsRepoErrorPreservingIs(t *testing.T) {
	errBoom := errors.New("boom")
	fake := &fakeRepo{listErr: errBoom}
	svc := usecase.New(fake)

	_, err := svc.List(context.Background(), domain.Filter{})
	if err == nil {
		t.Fatal("List() error = nil, want non-nil")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("errors.Is(err, errBoom) = false, want true; err=%v", err)
	}
}

func TestService_Get_PendingReturnsNotFound(t *testing.T) {
	fake := &fakeRepo{getByIDInst: domain.Institution{ModerationStatus: "pending"}}
	svc := usecase.New(fake)

	_, err := svc.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("Get() error = nil, want NotFound")
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, want true; err=%v", err)
	}
}

func TestService_List_PropagatesCancelledContext(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.List(ctx, domain.Filter{})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if fake.listCtx == nil || fake.listCtx.Err() != context.Canceled {
		t.Fatalf("listCtx.Err() = %v, want context.Canceled", fake.listCtx)
	}
}

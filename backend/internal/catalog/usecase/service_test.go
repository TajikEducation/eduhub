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

	createInst    domain.Institution
	createOwnerID uuid.UUID
	createErr     error

	updateID    uuid.UUID
	updatePatch domain.UpdateInstitutionInput
	updateErr   error

	ownerID  uuid.UUID
	ownerErr error

	statusID    uuid.UUID
	statusValue string
	statusErr   error
}

func (f *fakeRepo) List(ctx context.Context, filter domain.Filter) (domain.ListResult, error) {
	f.listFilter = filter
	f.listCtx = ctx
	return f.listResult, f.listErr
}

func (f *fakeRepo) GetByID(_ context.Context, _ uuid.UUID) (domain.Institution, error) {
	return f.getByIDInst, f.getByIDErr
}

func (f *fakeRepo) Create(_ context.Context, inst domain.Institution, ownerID uuid.UUID) (domain.Institution, error) {
	f.createInst, f.createOwnerID = inst, ownerID
	if f.createErr != nil {
		return domain.Institution{}, f.createErr
	}
	return inst, nil
}

func (f *fakeRepo) Update(_ context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error) {
	f.updateID, f.updatePatch = id, patch
	if f.updateErr != nil {
		return domain.Institution{}, f.updateErr
	}
	return domain.Institution{ID: id}, nil
}

func (f *fakeRepo) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (f *fakeRepo) UpdateRatingAvg(_ context.Context, _ uuid.UUID, _ float64, _ int) error {
	return nil
}

func (f *fakeRepo) ListByOwner(_ context.Context, _ uuid.UUID) ([]domain.Institution, error) {
	return nil, nil
}

func (f *fakeRepo) GetOwnerID(_ context.Context, id uuid.UUID) (uuid.UUID, error) {
	if f.ownerErr != nil {
		return uuid.UUID{}, f.ownerErr
	}
	return f.ownerID, nil
}

func (f *fakeRepo) SetModerationStatus(_ context.Context, id uuid.UUID, status string) error {
	f.statusID, f.statusValue = id, status
	return f.statusErr
}

func (f *fakeRepo) CreateStaff(_ context.Context, _ uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	return domain.StaffMember{Name: in.Name, RoleType: in.RoleType, RoleLabel: in.RoleLabel}, nil
}

func (f *fakeRepo) UpdateStaff(_ context.Context, id uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	return domain.StaffMember{ID: id, Name: in.Name, RoleType: in.RoleType, RoleLabel: in.RoleLabel}, nil
}

func (f *fakeRepo) DeleteStaff(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeRepo) GetStaffInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return f.ownerID, nil
}

func (f *fakeRepo) GetPublicStaffByID(_ context.Context, _ uuid.UUID) (domain.StaffMember, error) {
	return domain.StaffMember{}, nil
}

func (f *fakeRepo) CreateAchievement(_ context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error) {
	return domain.Achievement{OwnerType: "institution", OwnerID: institutionID, Title: in.Title, Year: in.Year, Category: in.Category, Description: in.Description}, nil
}

func (f *fakeRepo) DeleteAchievement(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeRepo) GetAchievementInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return f.ownerID, nil
}

func (f *fakeRepo) CreateGalleryItem(_ context.Context, _ uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error) {
	return domain.GalleryItem{S3Key: in.S3Key, Label: in.Label, SortOrder: in.SortOrder}, nil
}

func (f *fakeRepo) DeleteGalleryItem(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeRepo) GetGalleryItemInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return f.ownerID, nil
}

func (f *fakeRepo) CreateAlumnus(_ context.Context, _ uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error) {
	return domain.Alumnus{Name: in.Name, PhotoURL: in.PhotoURL, GradYear: in.GradYear, NowLabel: in.NowLabel}, nil
}

func (f *fakeRepo) DeleteAlumnus(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeRepo) GetAlumnusInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return f.ownerID, nil
}

func (f *fakeRepo) CreateNews(_ context.Context, _ uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	return domain.NewsArticle{Title: in.Title, Content: in.Content, Status: in.Status}, nil
}

func (f *fakeRepo) UpdateNews(_ context.Context, id uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	return domain.NewsArticle{ID: id, Title: in.Title, Content: in.Content, Status: in.Status}, nil
}

func (f *fakeRepo) DeleteNews(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeRepo) GetNewsInstitutionID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return f.ownerID, nil
}

func (f *fakeRepo) ListNews(_ context.Context, _ uuid.UUID) ([]domain.NewsArticle, error) {
	return nil, nil
}

func (f *fakeRepo) ListPublishedNews(_ context.Context, _ uuid.UUID) ([]domain.NewsArticle, error) {
	return nil, nil
}

func (f *fakeRepo) GetPublishedNewsByID(_ context.Context, _ uuid.UUID) (domain.NewsArticle, error) {
	return domain.NewsArticle{}, nil
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

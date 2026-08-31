package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// Авторизация (владелец институции/moderator/admin) — забота вызывающего транспорта через
// IsOwnerOfStaff/... ниже, тем же паттерном, что IsOwner для самой институции (write.go).

func (s *Service) CreateStaff(ctx context.Context, institutionID uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	if err := in.Validate(); err != nil {
		return domain.StaffMember{}, err
	}
	created, err := s.repo.CreateStaff(ctx, institutionID, in)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("usecase: create staff: %w", err)
	}
	return created, nil
}

func (s *Service) UpdateStaff(ctx context.Context, id uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	if err := in.Validate(); err != nil {
		return domain.StaffMember{}, err
	}
	updated, err := s.repo.UpdateStaff(ctx, id, in)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("usecase: update staff: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteStaff(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteStaff(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete staff: %w", err)
	}
	return nil
}

func (s *Service) IsOwnerOfStaff(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	institutionID, err := s.repo.GetStaffInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get staff institution id: %w", err)
	}
	return s.IsOwner(ctx, institutionID, userID)
}

// GetPublicStaff возвращает сотрудника по id — публичная страница /people/{id}, только если
// его институция approved (см. GetPublicStaffByID).
func (s *Service) GetPublicStaff(ctx context.Context, id uuid.UUID) (domain.StaffMember, error) {
	m, err := s.repo.GetPublicStaffByID(ctx, id)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("usecase: get public staff: %w", err)
	}
	return m, nil
}

func (s *Service) CreateAchievement(ctx context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error) {
	if err := in.Validate(); err != nil {
		return domain.Achievement{}, err
	}
	created, err := s.repo.CreateAchievement(ctx, institutionID, in)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("usecase: create achievement: %w", err)
	}
	return created, nil
}

func (s *Service) DeleteAchievement(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteAchievement(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete achievement: %w", err)
	}
	return nil
}

func (s *Service) IsOwnerOfAchievement(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	institutionID, err := s.repo.GetAchievementInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get achievement institution id: %w", err)
	}
	return s.IsOwner(ctx, institutionID, userID)
}

func (s *Service) CreateGalleryItem(ctx context.Context, institutionID uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error) {
	if err := in.Validate(); err != nil {
		return domain.GalleryItem{}, err
	}
	created, err := s.repo.CreateGalleryItem(ctx, institutionID, in)
	if err != nil {
		return domain.GalleryItem{}, fmt.Errorf("usecase: create gallery item: %w", err)
	}
	return created, nil
}

func (s *Service) DeleteGalleryItem(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteGalleryItem(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete gallery item: %w", err)
	}
	return nil
}

func (s *Service) IsOwnerOfGalleryItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	institutionID, err := s.repo.GetGalleryItemInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get gallery item institution id: %w", err)
	}
	return s.IsOwner(ctx, institutionID, userID)
}

func (s *Service) CreateAlumnus(ctx context.Context, institutionID uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error) {
	if err := in.Validate(); err != nil {
		return domain.Alumnus{}, err
	}
	created, err := s.repo.CreateAlumnus(ctx, institutionID, in)
	if err != nil {
		return domain.Alumnus{}, fmt.Errorf("usecase: create alumnus: %w", err)
	}
	return created, nil
}

func (s *Service) DeleteAlumnus(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteAlumnus(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete alumnus: %w", err)
	}
	return nil
}

func (s *Service) IsOwnerOfAlumnus(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	institutionID, err := s.repo.GetAlumnusInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get alumnus institution id: %w", err)
	}
	return s.IsOwner(ctx, institutionID, userID)
}

func (s *Service) ListNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error) {
	items, err := s.repo.ListNews(ctx, institutionID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list news: %w", err)
	}
	return items, nil
}

// ListPublishedNews возвращает только опубликованные новости учреждения — для публичной
// вкладки профиля (в отличие от ListNews, который отдаёт все статусы для кабинета).
func (s *Service) ListPublishedNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error) {
	items, err := s.repo.ListPublishedNews(ctx, institutionID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list published news: %w", err)
	}
	return items, nil
}

// GetPublishedNews возвращает одну опубликованную новость по id — для публичной страницы
// /news/{id}; draft чужой новости анонимному посетителю недоступен.
func (s *Service) GetPublishedNews(ctx context.Context, id uuid.UUID) (domain.NewsArticle, error) {
	article, err := s.repo.GetPublishedNewsByID(ctx, id)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("usecase: get published news: %w", err)
	}
	return article, nil
}

func (s *Service) CreateNews(ctx context.Context, institutionID uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	if err := in.Validate(); err != nil {
		return domain.NewsArticle{}, err
	}
	created, err := s.repo.CreateNews(ctx, institutionID, in)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("usecase: create news: %w", err)
	}
	return created, nil
}

func (s *Service) UpdateNews(ctx context.Context, id uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	if err := in.Validate(); err != nil {
		return domain.NewsArticle{}, err
	}
	updated, err := s.repo.UpdateNews(ctx, id, in)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("usecase: update news: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteNews(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteNews(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete news: %w", err)
	}
	return nil
}

func (s *Service) IsOwnerOfNews(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	institutionID, err := s.repo.GetNewsInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get news institution id: %w", err)
	}
	return s.IsOwner(ctx, institutionID, userID)
}

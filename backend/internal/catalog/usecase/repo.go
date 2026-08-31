// Package usecase содержит порты и оркестрацию каталога учреждений
// (сервисный слой между transport/http и repo/postgres).
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// InstitutionRepo — порт в БД для каталога институций. Реализация — internal/catalog/repo/postgres.
type InstitutionRepo interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Institution, error)
	Create(ctx context.Context, inst domain.Institution, ownerID uuid.UUID) (domain.Institution, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error)
	ListByOwner(ctx context.Context, userID uuid.UUID) ([]domain.Institution, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateRatingAvg(ctx context.Context, id uuid.UUID, avg float64, count int) error
	GetOwnerID(ctx context.Context, institutionID uuid.UUID) (uuid.UUID, error)
	SetModerationStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateStaff(ctx context.Context, institutionID uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error)
	UpdateStaff(ctx context.Context, id uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error)
	DeleteStaff(ctx context.Context, id uuid.UUID) error
	GetStaffInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetPublicStaffByID(ctx context.Context, id uuid.UUID) (domain.StaffMember, error)

	CreateAchievement(ctx context.Context, institutionID uuid.UUID, in domain.CreateAchievementInput) (domain.Achievement, error)
	DeleteAchievement(ctx context.Context, id uuid.UUID) error
	GetAchievementInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)

	CreateGalleryItem(ctx context.Context, institutionID uuid.UUID, in domain.CreateGalleryItemInput) (domain.GalleryItem, error)
	DeleteGalleryItem(ctx context.Context, id uuid.UUID) error
	GetGalleryItemInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)

	CreateAlumnus(ctx context.Context, institutionID uuid.UUID, in domain.CreateAlumnusInput) (domain.Alumnus, error)
	DeleteAlumnus(ctx context.Context, id uuid.UUID) error
	GetAlumnusInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)

	ListNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error)
	ListPublishedNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error)
	GetPublishedNewsByID(ctx context.Context, id uuid.UUID) (domain.NewsArticle, error)
	CreateNews(ctx context.Context, institutionID uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error)
	UpdateNews(ctx context.Context, id uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error)
	DeleteNews(ctx context.Context, id uuid.UUID) error
	GetNewsInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

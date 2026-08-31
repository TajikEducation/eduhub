// Package usecase — бизнес-логика профиля соискателя, его достижений и откликов на вакансии.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/applicants/domain"
)

// ApplicantRepo — порт в БД для профилей соискателей. Реализация — internal/applicants/repo/postgres.
type ApplicantRepo interface {
	Upsert(ctx context.Context, userID uuid.UUID, in domain.ApplicantInput) (domain.Applicant, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Applicant, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Applicant, error)
	ListPublic(ctx context.Context) ([]domain.Applicant, error)
}

// AchievementRepo — порт в БД для достижений соискателя.
type AchievementRepo interface {
	Create(ctx context.Context, applicantID uuid.UUID, in domain.AchievementInput) (domain.Achievement, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByApplicant(ctx context.Context, applicantID uuid.UUID) ([]domain.Achievement, error)
	GetApplicantID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

// ApplicationRepo — порт в БД для откликов на вакансии.
type ApplicationRepo interface {
	// Create идемпотентен — повторный отклик на ту же вакансию возвращает существующую запись
	// (UNIQUE(applicant_id, vacancy_id) в БД, см. миграцию 00011).
	Create(ctx context.Context, applicantID, vacancyID uuid.UUID) (domain.Application, error)
	ListVacancyIDsByApplicant(ctx context.Context, applicantID uuid.UUID) ([]uuid.UUID, error)
}

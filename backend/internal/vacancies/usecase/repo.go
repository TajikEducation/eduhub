// Package usecase — бизнес-логика модуля vacancies: создание, чтение, обновление, удаление.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
)

// VacancyRepo — порт в БД для вакансий. Реализация — internal/vacancies/repo/postgres.
type VacancyRepo interface {
	Create(ctx context.Context, institutionID uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error)
	Update(ctx context.Context, id uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Vacancy, error)
	GetInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ListByInstitution(ctx context.Context, institutionID uuid.UUID, onlyPublished bool) ([]domain.Vacancy, error)
	ListPublished(ctx context.Context, limit int) ([]domain.Vacancy, error)
}

// InstitutionChecker — порт в catalog: существование институции (нет физического FK между
// схемами communications и catalog) и проверка владельца — тот же паттерн, что у reviews.
type InstitutionChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
}

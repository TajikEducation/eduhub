// Package usecase — бизнес-логика модуля reviews: создание, чтение, ответ, модерация.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/reviews/domain"
)

// ReviewRepo — порт в БД для отзывов. Реализация — internal/reviews/repo/postgres.
type ReviewRepo interface {
	Create(ctx context.Context, r domain.Review) (domain.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Review, error)
	ListByInstitution(ctx context.Context, institutionID uuid.UUID, onlyApproved bool) ([]domain.Review, error)
	SetReply(ctx context.Context, id uuid.UUID, reply string) (domain.Review, error)
	SetStatus(ctx context.Context, id uuid.UUID, status domain.Status) error
	AggregateApproved(ctx context.Context, institutionID uuid.UUID) (avg float64, count int, err error)
}

// InstitutionChecker — порт в catalog: существование институции (нет физического FK),
// проверка владельца и синхронизация агрегата рейтинга при approve.
type InstitutionChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	IsOwner(ctx context.Context, institutionID uuid.UUID, userID uuid.UUID) (bool, error)
	UpdateRatingAvg(ctx context.Context, id uuid.UUID, avg float64, count int) error
}

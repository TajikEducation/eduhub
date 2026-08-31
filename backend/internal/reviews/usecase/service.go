package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/reviews/domain"
)

// Service — usecase-слой reviews.
type Service struct {
	repo         ReviewRepo
	institutions InstitutionChecker
}

// New создаёт Service.
func New(repo ReviewRepo, institutions InstitutionChecker) *Service {
	return &Service{repo: repo, institutions: institutions}
}

// Create создаёт отзыв со статусом pending (FR-16: обязательная модерация). Один отзыв
// на пару (институция, пользователь) — повтор → apperr.Conflict через UNIQUE в БД.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, in domain.CreateReviewInput) (domain.Review, error) {
	if err := in.Validate(); err != nil {
		return domain.Review{}, err
	}

	exists, err := s.institutions.Exists(ctx, in.InstitutionID)
	if err != nil {
		return domain.Review{}, fmt.Errorf("usecase: check institution exists: %w", err)
	}
	if !exists {
		return domain.Review{}, apperr.NotFound("institution", in.InstitutionID.String())
	}

	created, err := s.repo.Create(ctx, domain.Review{
		InstitutionID: in.InstitutionID,
		UserID:        userID,
		Rating:        in.Rating,
		Text:          in.Text,
		Status:        domain.StatusPending,
	})
	if err != nil {
		return domain.Review{}, fmt.Errorf("usecase: create review: %w", err)
	}
	return created, nil
}

// ListApproved возвращает опубликованные отзывы институции — для публичной страницы.
func (s *Service) ListApproved(ctx context.Context, institutionID uuid.UUID) ([]domain.Review, error) {
	items, err := s.repo.ListByInstitution(ctx, institutionID, true)
	if err != nil {
		return nil, fmt.Errorf("usecase: list approved reviews: %w", err)
	}
	return items, nil
}

// ListMine возвращает все отзывы институции (любой статус) — для кабинета учреждения.
// Доступ проверяется вызывающим транспортом (владелец/moderator).
func (s *Service) ListMine(ctx context.Context, institutionID uuid.UUID) ([]domain.Review, error) {
	items, err := s.repo.ListByInstitution(ctx, institutionID, false)
	if err != nil {
		return nil, fmt.Errorf("usecase: list all reviews: %w", err)
	}
	return items, nil
}

// Reply — ответ учреждения на отзыв (FR-17). Только владелец институции отзыва.
func (s *Service) Reply(ctx context.Context, reviewID uuid.UUID, actorUserID uuid.UUID, isPrivileged bool, reply string) (domain.Review, error) {
	if reply == "" {
		return domain.Review{}, apperr.Invalid(map[string]string{"reply": "обязательное поле"}, "пустой ответ")
	}

	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return domain.Review{}, fmt.Errorf("usecase: get review: %w", err)
	}

	if !isPrivileged {
		owner, err := s.institutions.IsOwner(ctx, review.InstitutionID, actorUserID)
		if err != nil {
			return domain.Review{}, fmt.Errorf("usecase: check institution owner: %w", err)
		}
		if !owner {
			return domain.Review{}, apperr.Forbidden("вы не владелец этой институции")
		}
	}

	updated, err := s.repo.SetReply(ctx, reviewID, reply)
	if err != nil {
		return domain.Review{}, fmt.Errorf("usecase: set reply: %w", err)
	}
	return updated, nil
}

// IsOwnerOfReview проверяет, является ли userID владельцем институции, к которой относится
// отзыв reviewID — тот же паттерн, что IsOwnerOfStaff и т.п. в internal/catalog/usecase.
func (s *Service) IsOwnerOfReview(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) (bool, error) {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return false, fmt.Errorf("usecase: get review: %w", err)
	}
	return s.institutions.IsOwner(ctx, review.InstitutionID, userID)
}

// Approve одобряет отзыв и синхронно пересчитывает rating_avg/review_count институции
// (упрощённая версия RatingSync — см. пакетный комментарий в catalog/repo/postgres/write.go).
func (s *Service) Approve(ctx context.Context, reviewID uuid.UUID) error {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("usecase: get review: %w", err)
	}

	if err := s.repo.SetStatus(ctx, reviewID, domain.StatusApproved); err != nil {
		return fmt.Errorf("usecase: approve review: %w", err)
	}

	avg, count, err := s.repo.AggregateApproved(ctx, review.InstitutionID)
	if err != nil {
		return fmt.Errorf("usecase: aggregate approved reviews: %w", err)
	}
	if err := s.institutions.UpdateRatingAvg(ctx, review.InstitutionID, avg, count); err != nil {
		return fmt.Errorf("usecase: sync rating avg: %w", err)
	}
	return nil
}

// Reject отклоняет отзыв (без апелляции — тот же принцип, что reject институции).
func (s *Service) Reject(ctx context.Context, reviewID uuid.UUID) error {
	if err := s.repo.SetStatus(ctx, reviewID, domain.StatusRejected); err != nil {
		return fmt.Errorf("usecase: reject review: %w", err)
	}
	return nil
}

// GetInstitutionID — для audit-записи модерации (см. internal/moderation).
func (s *Service) GetInstitutionID(ctx context.Context, reviewID uuid.UUID) (uuid.UUID, error) {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return uuid.Nil, err
		}
		return uuid.Nil, fmt.Errorf("usecase: get review: %w", err)
	}
	return review.InstitutionID, nil
}

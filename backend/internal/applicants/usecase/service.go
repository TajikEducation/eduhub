package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/applicants/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Service — usecase-слой applicants.
type Service struct {
	applicants   ApplicantRepo
	achievements AchievementRepo
	applications ApplicationRepo
}

// New создаёт Service.
func New(applicants ApplicantRepo, achievements AchievementRepo, applications ApplicationRepo) *Service {
	return &Service{applicants: applicants, achievements: achievements, applications: applications}
}

// UpsertMine создаёт или полностью заменяет профиль соискателя userID.
func (s *Service) UpsertMine(ctx context.Context, userID uuid.UUID, in domain.ApplicantInput) (domain.Applicant, error) {
	if err := in.Validate(); err != nil {
		return domain.Applicant{}, err
	}
	a, err := s.applicants.Upsert(ctx, userID, in)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("usecase: upsert applicant: %w", err)
	}
	return a, nil
}

// GetMine возвращает профиль соискателя userID (NotFound, если ещё не создан).
func (s *Service) GetMine(ctx context.Context, userID uuid.UUID) (domain.Applicant, error) {
	a, err := s.applicants.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("usecase: get my applicant profile: %w", err)
	}
	return a, nil
}

// GetVisible возвращает профиль по id — публичная страница /applicants/{id}; скрыт только
// draft (on_response тоже открывается по прямой ссылке — см. SRS-комментарий в мок-версии).
func (s *Service) GetVisible(ctx context.Context, id uuid.UUID) (domain.Applicant, error) {
	a, err := s.applicants.GetByID(ctx, id)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("usecase: get applicant: %w", err)
	}
	if a.Visibility == domain.VisibilityDraft {
		return domain.Applicant{}, apperr.NotFound("applicant", id.String())
	}
	return a, nil
}

// ListPublic возвращает полностью публичные профили — для каталога /applicants.
func (s *Service) ListPublic(ctx context.Context) ([]domain.Applicant, error) {
	items, err := s.applicants.ListPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase: list public applicants: %w", err)
	}
	return items, nil
}

// ListAchievements возвращает достижения соискателя — публичны вместе с самим профилем.
func (s *Service) ListAchievements(ctx context.Context, applicantID uuid.UUID) ([]domain.Achievement, error) {
	items, err := s.achievements.ListByApplicant(ctx, applicantID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list applicant achievements: %w", err)
	}
	return items, nil
}

// CreateAchievement добавляет достижение в собственный профиль userID.
func (s *Service) CreateAchievement(ctx context.Context, userID uuid.UUID, in domain.AchievementInput) (domain.Achievement, error) {
	if err := in.Validate(); err != nil {
		return domain.Achievement{}, err
	}
	mine, err := s.applicants.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("usecase: get my applicant profile: %w", err)
	}
	created, err := s.achievements.Create(ctx, mine.ID, in)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("usecase: create applicant achievement: %w", err)
	}
	return created, nil
}

// DeleteAchievement удаляет достижение — только если оно принадлежит профилю userID.
func (s *Service) DeleteAchievement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	applicantID, err := s.achievements.GetApplicantID(ctx, id)
	if err != nil {
		return fmt.Errorf("usecase: get achievement applicant id: %w", err)
	}
	mine, err := s.applicants.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("usecase: get my applicant profile: %w", err)
	}
	if applicantID != mine.ID {
		return apperr.Forbidden("это не ваше достижение")
	}
	if err := s.achievements.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete applicant achievement: %w", err)
	}
	return nil
}

// Apply создаёт отклик userID (соискателя) на вакансию vacancyID — идемпотентно. Требует, чтобы
// у пользователя уже был создан профиль соискателя (см. GetMine/UpsertMine).
func (s *Service) Apply(ctx context.Context, userID uuid.UUID, vacancyID uuid.UUID) (domain.Application, error) {
	mine, err := s.applicants.GetByUserID(ctx, userID)
	if err != nil {
		return domain.Application{}, fmt.Errorf("usecase: get my applicant profile: %w", err)
	}
	app, err := s.applications.Create(ctx, mine.ID, vacancyID)
	if err != nil {
		return domain.Application{}, fmt.Errorf("usecase: apply to vacancy: %w", err)
	}
	return app, nil
}

// ListMyApplications возвращает id вакансий, на которые userID уже откликнулся.
func (s *Service) ListMyApplications(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	mine, err := s.applicants.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("usecase: get my applicant profile: %w", err)
	}
	ids, err := s.applications.ListVacancyIDsByApplicant(ctx, mine.ID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list my applications: %w", err)
	}
	return ids, nil
}

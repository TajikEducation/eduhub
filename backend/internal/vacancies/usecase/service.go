package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
)

// Service — usecase-слой vacancies.
type Service struct {
	repo         VacancyRepo
	institutions InstitutionChecker
}

// New создаёт Service.
func New(repo VacancyRepo, institutions InstitutionChecker) *Service {
	return &Service{repo: repo, institutions: institutions}
}

// Create создаёт вакансию — только владелец институции. Авторизация (владелец/модератор) —
// забота вызывающего транспорта (см. internal/catalog аналогичный комментарий), здесь только
// принадлежность institutionID реально существующей институции.
func (s *Service) Create(ctx context.Context, institutionID uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error) {
	if err := in.Validate(); err != nil {
		return domain.Vacancy{}, err
	}
	exists, err := s.institutions.Exists(ctx, institutionID)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("usecase: check institution exists: %w", err)
	}
	if !exists {
		return domain.Vacancy{}, apperr.NotFound("institution", institutionID.String())
	}

	created, err := s.repo.Create(ctx, institutionID, in)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("usecase: create vacancy: %w", err)
	}
	return created, nil
}

// Update полностью заменяет данные вакансии.
func (s *Service) Update(ctx context.Context, id uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error) {
	if err := in.Validate(); err != nil {
		return domain.Vacancy{}, err
	}
	updated, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("usecase: update vacancy: %w", err)
	}
	return updated, nil
}

// Delete удаляет вакансию.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete vacancy: %w", err)
	}
	return nil
}

// ListMine возвращает все вакансии институции (любой статус) — для кабинета учреждения.
// Доступ проверяется вызывающим транспортом (владелец/moderator), тот же паттерн, что
// catalog.Service.ListNews.
func (s *Service) ListMine(ctx context.Context, institutionID uuid.UUID) ([]domain.Vacancy, error) {
	items, err := s.repo.ListByInstitution(ctx, institutionID, false)
	if err != nil {
		return nil, fmt.Errorf("usecase: list all vacancies: %w", err)
	}
	return items, nil
}

// ListPublic возвращает опубликованные вакансии институции — для публичной вкладки профиля.
func (s *Service) ListPublic(ctx context.Context, institutionID uuid.UUID) ([]domain.Vacancy, error) {
	items, err := s.repo.ListByInstitution(ctx, institutionID, true)
	if err != nil {
		return nil, fmt.Errorf("usecase: list published vacancies: %w", err)
	}
	return items, nil
}

// ListGlobalPublished возвращает опубликованные вакансии всех институций (для /vacancies) —
// самые новые первыми, до limit штук.
func (s *Service) ListGlobalPublished(ctx context.Context, limit int) ([]domain.Vacancy, error) {
	items, err := s.repo.ListPublished(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("usecase: list global published vacancies: %w", err)
	}
	return items, nil
}

// GetPublished возвращает вакансию, только если она опубликована — для публичной страницы
// /vacancies/{id}; чужая draft-вакансия недоступна анонимному посетителю.
func (s *Service) GetPublished(ctx context.Context, id uuid.UUID) (domain.Vacancy, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("usecase: get vacancy: %w", err)
	}
	if v.Status != domain.StatusPublished {
		return domain.Vacancy{}, apperr.NotFound("vacancy", id.String())
	}
	return v, nil
}

// GetInstitutionID — институция-владелец вакансии (для audit/owner-проверок).
func (s *Service) GetInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	instID, err := s.repo.GetInstitutionID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("usecase: get vacancy institution id: %w", err)
	}
	return instID, nil
}

// IsOwnerOfVacancy проверяет, является ли userID владельцем институции, которой принадлежит
// вакансия id — тот же паттерн, что reviews.IsOwnerOfReview.
func (s *Service) IsOwnerOfVacancy(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	instID, err := s.repo.GetInstitutionID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get vacancy institution id: %w", err)
	}
	return s.institutions.IsOwner(ctx, instID, userID)
}

package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// approvedOnly — единственный статус, который видит публичный каталог. Сервис форсирует
// его независимо от того, что пришло в фильтре — клиент не может запросить другие статусы.
var approvedOnly = []string{"approved"}

// Service — usecase-слой каталога институций: List/Get поверх InstitutionRepo.
type Service struct {
	repo InstitutionRepo
}

// New создаёт Service поверх переданного InstitutionRepo.
func New(repo InstitutionRepo) *Service {
	return &Service{repo: repo}
}

// List возвращает страницу институций. Всегда форсирует Statuses=["approved"] — публичный
// каталог не показывает pending/rejected ни при каких условиях, что бы ни передал вызывающий.
func (s *Service) List(ctx context.Context, f domain.Filter) (domain.ListResult, error) {
	f.Statuses = approvedOnly
	f.Normalize()

	result, err := s.repo.List(ctx, f)
	if err != nil {
		return domain.ListResult{}, fmt.Errorf("usecase: list institutions: %w", err)
	}
	return result, nil
}

// Get возвращает институцию по id. Не approved (pending/rejected) — NotFound, не Forbidden:
// публичный каталог не должен раскрывать сам факт существования немодерированной институции.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.Institution, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("usecase: get institution: %w", err)
	}
	if inst.ModerationStatus != "approved" {
		return domain.Institution{}, apperr.NotFound("institution", id.String())
	}
	return inst, nil
}

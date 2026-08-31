package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Register создаёт новую институцию со статусом pending и владельцем ownerID (E3.3).
// Публичный каталог её не увидит, пока модератор не одобрит (см. Service.List/Get выше).
func (s *Service) Register(ctx context.Context, ownerID uuid.UUID, in domain.CreateInstitutionInput) (domain.Institution, error) {
	if err := in.Validate(); err != nil {
		return domain.Institution{}, err
	}

	inst := *domain.NewInstitution(uuid.New(), in.Name, in.Region)
	inst.Types = in.Types
	inst.City = in.City
	inst.District = in.District
	inst.Description = in.Description
	inst.Phone = in.Phone
	inst.Email = in.Email
	inst.Website = in.Website
	inst.Price = in.Price
	inst.Lat = in.Lat
	inst.Lng = in.Lng
	inst.ModerationStatus = "pending"
	inst.Plan = "free"

	created, err := s.repo.Create(ctx, inst, ownerID)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("usecase: register institution: %w", err)
	}
	return created, nil
}

// Update частично обновляет профиль институции (E3.4, урезанная версия — см.
// domain.UpdateInstitutionInput). Авторизация (владелец/модератор) — забота вызывающего
// транспорта через IsOwner, не usecase: usecase не знает про роли (см. internal/auth/rbac).
func (s *Service) Update(ctx context.Context, id uuid.UUID, patch domain.UpdateInstitutionInput) (domain.Institution, error) {
	updated, err := s.repo.Update(ctx, id, patch)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("usecase: update institution: %w", err)
	}
	return updated, nil
}

// ListMine возвращает все институции userID (любого moderation_status) — для кабинета учреждения.
func (s *Service) ListMine(ctx context.Context, userID uuid.UUID) ([]domain.Institution, error) {
	insts, err := s.repo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list my institutions: %w", err)
	}
	return insts, nil
}

// GetMine возвращает полную карточку институции (со всеми сателлитами) для её владельца —
// в отличие от Service.Get, не форсирует approved-only: владельцу нужно видеть свою
// институцию в кабинете независимо от статуса модерации (pending/rejected/approved).
func (s *Service) GetMine(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Institution, error) {
	owner, err := s.IsOwner(ctx, id, userID)
	if err != nil {
		return domain.Institution{}, err
	}
	if !owner {
		return domain.Institution{}, apperr.Forbidden("вы не владелец этой институции")
	}

	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Institution{}, fmt.Errorf("usecase: get my institution: %w", err)
	}
	return inst, nil
}

// Exists проверяет существование институции — порт для internal/reviews (валидация
// institution_id, на который там нет физического FK).
func (s *Service) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	ok, err := s.repo.Exists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: check institution exists: %w", err)
	}
	return ok, nil
}

// UpdateRatingAvg — см. пакетный комментарий в repo/postgres/write.go. Вызывается
// internal/moderation при approve отзыва.
func (s *Service) UpdateRatingAvg(ctx context.Context, id uuid.UUID, avg float64, count int) error {
	if err := s.repo.UpdateRatingAvg(ctx, id, avg, count); err != nil {
		return fmt.Errorf("usecase: update rating avg: %w", err)
	}
	return nil
}

// IsOwner проверяет, является ли userID владельцем институции id.
func (s *Service) IsOwner(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	ownerID, err := s.repo.GetOwnerID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("usecase: get institution owner: %w", err)
	}
	return ownerID == userID, nil
}

// SetModerationStatus меняет статус модерации — вызывается internal/moderation/usecase,
// не публичным транспортом каталога напрямую (см. internal/moderation/transport/http).
func (s *Service) SetModerationStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := s.repo.SetModerationStatus(ctx, id, status); err != nil {
		return fmt.Errorf("usecase: set moderation status: %w", err)
	}
	return nil
}

package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// ChildService — оркестрация создания привязки родитель↔учреждение (E2.6, FR-15).
type ChildService struct {
	children     ChildRepo
	institutions InstitutionStatusChecker
	clock        clock.Clock
}

// NewChildService создаёт ChildService.
func NewChildService(children ChildRepo, institutions InstitutionStatusChecker, clk clock.Clock) *ChildService {
	return &ChildService{children: children, institutions: institutions, clock: clk}
}

// CreateChild создаёт привязку родитель↔учреждение со confirmation_status='pending'.
// Enum-валидация ageGroup/status — забота transport-слоя (синтаксис), здесь — только
// инвариант операции: учреждение должно быть approved. FK на catalog.institutions гарантирует
// только существование строки, не значение moderation_status — поэтому явный guard.
func (s *ChildService) CreateChild(ctx context.Context, userID, institutionID uuid.UUID, ageGroup, status string) (domain.Child, error) {
	approved, err := s.institutions.IsApproved(ctx, institutionID)
	if err != nil {
		return domain.Child{}, fmt.Errorf("create child: %w", err)
	}
	if !approved {
		return domain.Child{}, apperr.ConflictCode("institution_not_approved", "учреждение не прошло модерацию")
	}

	c := domain.Child{
		ID:                 uuid.New(),
		UserID:             userID,
		InstitutionID:      institutionID,
		AgeGroup:           ageGroup,
		Status:             status,
		ConfirmationStatus: "pending",
		CreatedAt:          s.clock.Now(),
	}

	created, err := s.children.Create(ctx, c)
	if err != nil {
		return domain.Child{}, fmt.Errorf("create child: %w", err)
	}
	return created, nil
}

// ListPendingByInstitution возвращает привязки со confirmation_status='pending' для очереди
// подтверждения учреждением.
func (s *ChildService) ListPendingByInstitution(ctx context.Context, institutionID uuid.UUID) ([]domain.Child, error) {
	children, err := s.children.ListPendingByInstitution(ctx, institutionID)
	if err != nil {
		return nil, fmt.Errorf("list pending children by institution: %w", err)
	}
	return children, nil
}

// ConfirmChild подтверждает привязку родитель↔учреждение (pending→confirmed) с audit-записью.
func (s *ChildService) ConfirmChild(ctx context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error) {
	c, err := s.children.Confirm(ctx, childID, actorID, actorRole, requestID)
	if err != nil {
		return domain.Child{}, fmt.Errorf("confirm child: %w", err)
	}
	return c, nil
}

// RejectChild отклоняет привязку родитель↔учреждение (pending→rejected) с обязательной
// структурированной причиной и audit-записью.
func (s *ChildService) RejectChild(ctx context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error) {
	c, err := s.children.Reject(ctx, childID, actorID, actorRole, reasonCode, reasonText, requestID)
	if err != nil {
		return domain.Child{}, fmt.Errorf("reject child: %w", err)
	}
	return c, nil
}

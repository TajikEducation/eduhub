// Package usecase — бизнес-логика модуля moderation: approve/reject институций с записью
// в moderation.audit_log (E3.5).
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/moderation/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// InstitutionStatusSetter — порт в каталог: единственное, что нужно moderation от catalog
// (internal/catalog/usecase.Service удовлетворяет этому интерфейсу).
type InstitutionStatusSetter interface {
	SetModerationStatus(ctx context.Context, id uuid.UUID, status string) error
}

// Recorder — порт записи в moderation.audit_log. Реализация — internal/moderation/repo/postgres.
type Recorder interface {
	Record(ctx context.Context, e domain.Entry) error
}

// ReviewModerator — порт в reviews: approve/reject отзыва (internal/reviews/usecase.Service
// удовлетворяет этому интерфейсу).
type ReviewModerator interface {
	Approve(ctx context.Context, reviewID uuid.UUID) error
	Reject(ctx context.Context, reviewID uuid.UUID) error
}

// Service — usecase-слой moderation: approve/reject c обязательной записью в audit_log.
// Запись статуса и audit — два последовательных вызова, не единая транзакция (см. пакетный
// комментарий catalog/repo/postgres/write.go — тот же осознанный компромисс).
type Service struct {
	institutions InstitutionStatusSetter
	audit        Recorder
	reviews      ReviewModerator
}

// New создаёт Service поверх переданных портов. reviews может быть nil, если модуль reviews
// не подключён (тогда ApproveReview/RejectReview недоступны — см. router.go).
func New(institutions InstitutionStatusSetter, audit Recorder, reviews ReviewModerator) *Service {
	return &Service{institutions: institutions, audit: audit, reviews: reviews}
}

// ApproveInstitution переводит институцию в approved и пишет решение в audit_log.
func (s *Service) ApproveInstitution(ctx context.Context, actor rbac.Principal, requestID string, institutionID uuid.UUID) error {
	if err := s.institutions.SetModerationStatus(ctx, institutionID, "approved"); err != nil {
		return fmt.Errorf("usecase: approve institution: %w", err)
	}

	if err := s.audit.Record(ctx, domain.Entry{
		ActorType:  "user",
		ActorID:    &actor.UserID,
		ActorRole:  actor.Role,
		Action:     "approve_institution",
		TargetType: "institution",
		TargetID:   institutionID,
		RequestID:  requestID,
	}); err != nil {
		return fmt.Errorf("usecase: record approve audit: %w", err)
	}
	return nil
}

// ApproveReview одобряет отзыв (запускает пересчёт rating_avg внутри reviews.Service.Approve)
// и пишет решение в audit_log.
func (s *Service) ApproveReview(ctx context.Context, actor rbac.Principal, requestID string, reviewID uuid.UUID) error {
	if err := s.reviews.Approve(ctx, reviewID); err != nil {
		return fmt.Errorf("usecase: approve review: %w", err)
	}
	if err := s.audit.Record(ctx, domain.Entry{
		ActorType: "user", ActorID: &actor.UserID, ActorRole: actor.Role,
		Action: "approve_review", TargetType: "review", TargetID: reviewID, RequestID: requestID,
	}); err != nil {
		return fmt.Errorf("usecase: record approve review audit: %w", err)
	}
	return nil
}

// RejectReview отклоняет отзыв. reasonText обязателен — тот же принцип, что RejectInstitution.
func (s *Service) RejectReview(ctx context.Context, actor rbac.Principal, requestID string, reviewID uuid.UUID, reasonCode, reasonText string) error {
	if reasonText == "" {
		return apperr.Invalid(map[string]string{"reason_text": "обязательное поле"}, "reject требует причину")
	}
	if err := s.reviews.Reject(ctx, reviewID); err != nil {
		return fmt.Errorf("usecase: reject review: %w", err)
	}
	if err := s.audit.Record(ctx, domain.Entry{
		ActorType: "user", ActorID: &actor.UserID, ActorRole: actor.Role,
		Action: "reject_review", TargetType: "review", TargetID: reviewID,
		ReasonCode: &reasonCode, ReasonText: &reasonText, RequestID: requestID,
	}); err != nil {
		return fmt.Errorf("usecase: record reject review audit: %w", err)
	}
	return nil
}

// RejectInstitution переводит институцию в rejected. reasonText обязателен (без апелляции —
// см. SRS: «обязательный структурированный reason для reject»).
func (s *Service) RejectInstitution(ctx context.Context, actor rbac.Principal, requestID string, institutionID uuid.UUID, reasonCode, reasonText string) error {
	if reasonText == "" {
		return apperr.Invalid(map[string]string{"reason_text": "обязательное поле"}, "reject требует причину")
	}

	if err := s.institutions.SetModerationStatus(ctx, institutionID, "rejected"); err != nil {
		return fmt.Errorf("usecase: reject institution: %w", err)
	}

	if err := s.audit.Record(ctx, domain.Entry{
		ActorType:  "user",
		ActorID:    &actor.UserID,
		ActorRole:  actor.Role,
		Action:     "reject_institution",
		TargetType: "institution",
		TargetID:   institutionID,
		ReasonCode: &reasonCode,
		ReasonText: &reasonText,
		RequestID:  requestID,
	}); err != nil {
		return fmt.Errorf("usecase: record reject audit: %w", err)
	}
	return nil
}

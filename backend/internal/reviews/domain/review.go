// Package domain — доменная модель модуля reviews: упрощённая версия SRS-спека
// (docs/EduHub_Database_Schema.md, reviews.reviews) — одна общая оценка вместо 8 метрик,
// без верификации через auth.children (не построено), без dispute-workflow.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Status — статус модерации отзыва (FR-16: обязательная модерация).
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Review — отзыв пользователя об институции.
type Review struct {
	ID            uuid.UUID
	InstitutionID uuid.UUID
	UserID        uuid.UUID
	Rating        int
	Text          string
	Reply         *string
	RepliedAt     *time.Time
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateReviewInput — данные для создания отзыва.
type CreateReviewInput struct {
	InstitutionID uuid.UUID
	Rating        int
	Text          string
}

func (in CreateReviewInput) Validate() error {
	fields := map[string]string{}
	if in.Rating < 1 || in.Rating > 5 {
		fields["rating"] = "должно быть от 1 до 5"
	}
	if strings.TrimSpace(in.Text) == "" {
		fields["text"] = "обязательное поле"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные отзыва")
	}
	return nil
}

// Package usecase содержит порты и оркестрацию каталога учреждений
// (сервисный слой между transport/http и repo/postgres).
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// InstitutionRepo — порт в БД для каталога институций. Реализация — internal/catalog/repo/postgres.
type InstitutionRepo interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Institution, error)
	// IsApproved — apperr.NotFound, если учреждение с id не существует. Порт также реализует
	// auth/usecase.InstitutionStatusChecker (E2.6, кросс-схемная проверка перед созданием
	// привязки родитель↔учреждение) — структурная типизация Go, без отдельного адаптера.
	IsApproved(ctx context.Context, id uuid.UUID) (bool, error)
}

// Package usecase содержит порты и оркестрацию каталога учреждений
// (сервисный слой между transport/http и repo/postgres).
package usecase

import (
	"context"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// InstitutionRepo — порт в БД для каталога институций. Реализация — internal/catalog/repo/postgres.
type InstitutionRepo interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
}

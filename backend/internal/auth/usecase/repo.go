// Package usecase — оркестрация сессий auth-сервиса (выпуск/ротация токенов, E2.3).
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
)

// RefreshTokenRepo — порт в БД для refresh-токенов. Реализация — internal/auth/repo/postgres.
type RefreshTokenRepo interface {
	Create(ctx context.Context, rt domain.RefreshToken) error
	// FindByHash возвращает apperr.NotFound (обёрнутый), если хеш не найден.
	FindByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	// Revoke отзывает один токен; replacedBy — nil при отзыве без ротации (logout/reuse).
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time, replacedBy *uuid.UUID) error
	// RevokeFamily отзывает разом все ещё не отозванные токены семьи — reuse-detection.
	RevokeFamily(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error
}

// UserRoleLookup — порт чтения текущей роли пользователя. Нужен Rotate(): роль в новом
// access-токене должна быть АКТУАЛЬНОЙ на момент ротации, не унаследованной от исходного
// Issue() — иначе смена роли (модерация забанила/повысила) не подействует, пока не истечёт
// 30-дневный refresh и человек не перелогинится заново.
type UserRoleLookup interface {
	RoleByUserID(ctx context.Context, userID uuid.UUID) (string, error)
}

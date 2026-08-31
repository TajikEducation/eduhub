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

// UserRepo — порт в БД для пользователей (E2.4). Реализация — internal/auth/repo/postgres.
type UserRepo interface {
	// Create вставляет нового пользователя, возвращает его с полями, сгенерированными БД
	// (id, created_at, updated_at). Конфликт email → apperr.ConflictCode("email_taken", ...).
	Create(ctx context.Context, u domain.User) (domain.User, error)
	// FindByEmail — apperr.NotFound, если email не найден.
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	// FindByID — apperr.NotFound, если не найден.
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

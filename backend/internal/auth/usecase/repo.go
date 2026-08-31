// Package usecase — бизнес-логика модуля auth: регистрация, вход, refresh, чтение своего профиля.
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
)

// UserRepo — порт доступа к auth.users, определённый потребителем (usecase), не реализацией.
type UserRepo interface {
	// Create вставляет нового пользователя. Повтор email → apperr.Conflict("email_taken").
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// RefreshTokenRepo — порт доступа к auth.refresh_tokens.
type RefreshTokenRepo interface {
	// Store сохраняет хеш нового refresh-токена.
	Store(ctx context.Context, t RefreshTokenRecord) error
	// GetByHash находит токен по хешу. Не найден/отозван/истёк — apperr.Unauthorized.
	GetByHash(ctx context.Context, hash string) (RefreshTokenRecord, error)
	// Revoke помечает токен использованным при ротации (replacedBy — id нового токена).
	Revoke(ctx context.Context, id uuid.UUID, replacedBy uuid.UUID) error
}

// RefreshTokenRecord — запись auth.refresh_tokens.
type RefreshTokenRecord struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	FamilyID   uuid.UUID
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *uuid.UUID
}

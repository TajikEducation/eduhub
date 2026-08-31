// Package domain — сущности auth-сервиса, без зависимостей от transport/DB.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken — строка auth.refresh_tokens. FamilyID объединяет всю цепочку ротации одной
// сессии — reuse-detection (E2.3) отзывает разом все токены с одним FamilyID.
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	FamilyID   uuid.UUID
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *uuid.UUID
	CreatedAt  time.Time
}

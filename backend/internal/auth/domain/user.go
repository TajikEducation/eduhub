package domain

import (
	"time"

	"github.com/google/uuid"
)

// User — строка auth.users. Поля 1:1 с миграцией 00006_auth_users_and_sessions.sql.
type User struct {
	ID               uuid.UUID
	Email            string
	DisplayName      *string
	Locale           string
	Phone            *string
	PasswordHash     *string
	Role             string
	Status           string
	EmailVerifiedAt  *time.Time
	ConsentAt        time.Time
	ConsentVersion   string
	FailedLoginCount int
	LockedUntil      *time.Time
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

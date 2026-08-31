// Package domain — доменная модель модуля auth: пользователь и его инварианты.
// Хеширование пароля и выпуск токенов — не доменная забота, см. internal/auth/security.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Role — роль пользователя для RBAC. Матрица «роль × эндпоинт» растёт в вехах 3-6.
type Role string

const (
	RoleUser        Role = "user"
	RoleInstitution Role = "institution"
	RoleModerator   Role = "moderator"
	RoleAdmin       Role = "admin"
)

// Status — статус учётной записи. "unverified" зарезервирован под будущий email-verification.
type Status string

const (
	StatusUnverified Status = "unverified"
	StatusActive     Status = "active"
	StatusBanned     Status = "banned"
	StatusDeleted    Status = "deleted"
)

// User — учётная запись (auth.users). PasswordHash пуст для будущих Google-only аккаунтов
// (см. SRS Веха 2, ещё не реализовано в этой версии).
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         Role
	Status       Status
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// minPasswordLen — минимальная длина пароля при регистрации. Не NIST-полная политика
// (энтропия/чёрные списки) — базовая защита от тривиальных паролей.
const minPasswordLen = 8

// RegisterInput — то, что нужно usecase.Register от вызывающего.
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

// Validate проверяет синтаксис входных данных регистрации. Проверка занятости email — это
// уникальность в БД, а не синтаксическая валидация, поэтому здесь не проверяется.
func (in RegisterInput) Validate() error {
	fields := map[string]string{}

	email := strings.TrimSpace(in.Email)
	if email == "" {
		fields["email"] = "обязательное поле"
	} else if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		fields["email"] = "некорректный формат email"
	}

	if len(in.Password) < minPasswordLen {
		fields["password"] = "минимум 8 символов"
	}

	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные регистрации")
	}
	return nil
}

// NormalizedEmail — email в каноническом виде для хранения/сравнения (нижний регистр, без пробелов).
func (in RegisterInput) NormalizedEmail() string {
	return strings.ToLower(strings.TrimSpace(in.Email))
}

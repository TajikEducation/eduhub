package domain

import (
	"time"

	"github.com/google/uuid"
)

// VerificationCode — строка auth.verification_codes. Код короткоживущий и низкоэнтропийный
// (6 цифр, вводится вручную из письма) — защита не от брутфорса хеша (sha256 шестизначного
// числа тривиально обратим по радужным таблицам), а от AttemptsCount (лимит попыток) и
// короткого ExpiresAt. Использованный/просроченный/исчерпавший попытки код удаляется целиком,
// отдельного поля "consumed" в схеме нет.
type VerificationCode struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Channel       string // "email" | "phone" (сейчас только "email" используется)
	Purpose       string // "register" | "password_reset"
	CodeHash      string
	AttemptsCount int
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

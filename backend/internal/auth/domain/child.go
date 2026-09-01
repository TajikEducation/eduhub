package domain

import (
	"time"

	"github.com/google/uuid"
)

// Child — привязка родитель↔учреждение (SRS §7, FR-15). Явно не хранит имя/фамилию/фото/дату
// рождения/контакты ребёнка — минимизация PII (см. docs/EduHub_Database_Schema.md, auth.children).
type Child struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	InstitutionID      uuid.UUID
	AgeGroup           string
	Status             string
	ConfirmationStatus string
	ConfirmedBy        *uuid.UUID
	ConfirmedAt        *time.Time
	CreatedAt          time.Time
}

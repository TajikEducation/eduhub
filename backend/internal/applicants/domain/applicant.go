// Package domain — доменная модель профиля соискателя и откликов на вакансии (веха 5, ядро):
// упрощённая версия SRS-спека (docs/EduHub_Database_Schema.md, communications.applicants/
// applicant_achievements/applications) — без employer_responses (приглашения работодателя,
// не запрошены явно).
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Bilingual — двуязычное поле (ru/tg). Собственная копия типа — схема communications не
// разделяет типы с catalog (тот же принцип, что у reviews/vacancies/chat).
type Bilingual struct {
	RU string
	TG string
}

const (
	VisibilityDraft      = "draft"
	VisibilityOnResponse = "on_response"
	VisibilityPublic     = "public"
)

// Applicant — профиль соискателя (резюме), один на пользователя.
type Applicant struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         Bilingual
	PhotoURL     *string
	Position     Bilingual
	Bio          *Bilingual
	Education    []Bilingual
	Experience   []Bilingual
	Skills       []Bilingual
	Email        *string
	Phone        *string
	CVS3Key      *string
	Visibility   string
	HideContacts bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ApplicantInput — данные для создания/полной замены профиля.
type ApplicantInput struct {
	Name         Bilingual
	PhotoURL     *string
	Position     Bilingual
	Bio          *Bilingual
	Education    []Bilingual
	Experience   []Bilingual
	Skills       []Bilingual
	Email        *string
	Phone        *string
	CVS3Key      *string
	Visibility   string
	HideContacts bool
}

func (in ApplicantInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Name.RU) == "" || strings.TrimSpace(in.Name.TG) == "" {
		fields["name"] = "обязательное поле (ru и tg)"
	}
	if strings.TrimSpace(in.Position.RU) == "" || strings.TrimSpace(in.Position.TG) == "" {
		fields["position"] = "обязательное поле (ru и tg)"
	}
	switch in.Visibility {
	case VisibilityDraft, VisibilityOnResponse, VisibilityPublic:
	default:
		fields["visibility"] = "должно быть draft, on_response или public"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные профиля соискателя")
	}
	return nil
}

// Achievement — достижение соискателя (не переиспользует catalog.achievements — правило
// владения схемами, см. docs/EduHub_Database_Schema.md). Title/Description — свободный текст
// (пишет сам соискатель), не Bilingual.
type Achievement struct {
	ID          uuid.UUID
	ApplicantID uuid.UUID
	Title       string
	Year        *int
	Category    *string
	Description *string
	CreatedAt   time.Time
}

// AchievementInput — данные для создания достижения.
type AchievementInput struct {
	Title       string
	Year        *int
	Category    *string
	Description *string
}

func (in AchievementInput) Validate() error {
	if strings.TrimSpace(in.Title) == "" {
		return apperr.Invalid(map[string]string{"title": "обязательное поле"}, "пустой заголовок достижения")
	}
	if in.Category != nil {
		switch *in.Category {
		case "gold", "silver", "bronze", "special":
		default:
			return apperr.Invalid(map[string]string{"category": "должно быть gold, silver, bronze или special"}, "некорректная категория")
		}
	}
	return nil
}

const (
	ApplicationStatusSent   = "sent"
	ApplicationStatusViewed = "viewed"
	ApplicationStatusClosed = "closed"
)

// Application — отклик соискателя на вакансию, идемпотентно на уровне БД
// (UNIQUE(applicant_id, vacancy_id)).
type Application struct {
	ID          uuid.UUID
	ApplicantID uuid.UUID
	VacancyID   uuid.UUID
	Status      string
	CreatedAt   time.Time
}

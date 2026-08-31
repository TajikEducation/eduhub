// Package domain — доменная модель модуля vacancies: упрощённая версия SRS-спека
// (docs/EduHub_Database_Schema.md, communications.vacancies) — без applicants/applications
// (отклики соискателей) — вторая волна, не запрошена явно.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// Bilingual — двуязычное поле (ru/tg). Собственная копия типа, не catalog.domain.Bilingual —
// схема communications не разделяет типы с catalog (тот же принцип, что у reviews).
type Bilingual struct {
	RU string
	TG string
}

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// Vacancy — вакансия учреждения.
type Vacancy struct {
	ID            uuid.UUID
	InstitutionID uuid.UUID
	Title         Bilingual
	Description   Bilingual
	Requirements  []Bilingual
	SalaryFrom    *int
	SalaryTo      *int
	Employment    Bilingual
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// VacancyInput — данные для создания/полной замены вакансии (Create и Update используют
// одну форму — тот же паттерн, что у catalog.CreateNewsInput/UpdateNews).
type VacancyInput struct {
	Title        Bilingual
	Description  Bilingual
	Requirements []Bilingual
	SalaryFrom   *int
	SalaryTo     *int
	Employment   Bilingual
	Status       string
}

func (in VacancyInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Title.RU) == "" || strings.TrimSpace(in.Title.TG) == "" {
		fields["title"] = "обязательное поле (ru и tg)"
	}
	if strings.TrimSpace(in.Description.RU) == "" || strings.TrimSpace(in.Description.TG) == "" {
		fields["description"] = "обязательное поле (ru и tg)"
	}
	if strings.TrimSpace(in.Employment.RU) == "" || strings.TrimSpace(in.Employment.TG) == "" {
		fields["employment"] = "обязательное поле (ru и tg)"
	}
	if in.Status != StatusDraft && in.Status != StatusPublished {
		fields["status"] = "должно быть draft или published"
	}
	if in.SalaryFrom != nil && *in.SalaryFrom < 0 {
		fields["salary_from"] = "не может быть отрицательным"
	}
	if in.SalaryTo != nil && *in.SalaryTo < 0 {
		fields["salary_to"] = "не может быть отрицательным"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные вакансии")
	}
	return nil
}

// InstitutionSummary — минимальный набор полей институции, нужный для публичной карточки
// вакансии (список/детали) — заполняется через порт InstitutionInfo (см. usecase/repo.go),
// не через прямой импорт catalog.domain (владение схемами).
type InstitutionSummary struct {
	ID              uuid.UUID
	Name            Bilingual
	Types           []string
	Region          string
	District        *string
	City            *Bilingual
	CoverPhotoS3Key *string
	Verified        bool
}

// VacancyWithInstitution — вакансия, обогащённая сводкой об учреждении, для публичного
// глобального списка/детали (/api/v1/vacancies).
type VacancyWithInstitution struct {
	Vacancy     Vacancy
	Institution InstitutionSummary
}

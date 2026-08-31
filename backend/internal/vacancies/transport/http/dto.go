package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
)

type bilingualDTO struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func toBilingualDTO(b domain.Bilingual) bilingualDTO { return bilingualDTO{RU: b.RU, TG: b.TG} }
func fromBilingualDTO(b bilingualDTO) domain.Bilingual {
	return domain.Bilingual{RU: b.RU, TG: b.TG}
}

// vacancyDTO — публичный контракт вакансии (без сводки об учреждении — используется там, где
// институция уже известна из URL: /institutions/{id}/vacancies[/mine]).
type vacancyDTO struct {
	ID            uuid.UUID      `json:"id"`
	InstitutionID uuid.UUID      `json:"institution_id"`
	Title         bilingualDTO   `json:"title"`
	Description   bilingualDTO   `json:"description"`
	Requirements  []bilingualDTO `json:"requirements,omitempty"`
	SalaryFrom    *int           `json:"salary_from,omitempty"`
	SalaryTo      *int           `json:"salary_to,omitempty"`
	Employment    bilingualDTO   `json:"employment"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func toVacancyDTO(v domain.Vacancy) vacancyDTO {
	reqs := make([]bilingualDTO, len(v.Requirements))
	for i, r := range v.Requirements {
		reqs[i] = toBilingualDTO(r)
	}
	return vacancyDTO{
		ID: v.ID, InstitutionID: v.InstitutionID, Title: toBilingualDTO(v.Title), Description: toBilingualDTO(v.Description),
		Requirements: reqs, SalaryFrom: v.SalaryFrom, SalaryTo: v.SalaryTo, Employment: toBilingualDTO(v.Employment),
		Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}

// institutionSummaryDTO — сводка об учреждении, встраиваемая в глобальный список/деталь
// вакансий (/api/v1/vacancies), где учреждение заранее не известно из URL.
type institutionSummaryDTO struct {
	ID              uuid.UUID     `json:"id"`
	Name            bilingualDTO  `json:"name"`
	Types           []string      `json:"types"`
	Region          string        `json:"region"`
	District        *string       `json:"district,omitempty"`
	City            *bilingualDTO `json:"city,omitempty"`
	CoverPhotoS3Key *string       `json:"cover_photo_s3_key,omitempty"`
	Verified        bool          `json:"verified"`
}

func toInstitutionSummaryDTO(s domain.InstitutionSummary) institutionSummaryDTO {
	var city *bilingualDTO
	if s.City != nil {
		dto := toBilingualDTO(*s.City)
		city = &dto
	}
	return institutionSummaryDTO{
		ID: s.ID, Name: toBilingualDTO(s.Name), Types: s.Types, Region: s.Region,
		District: s.District, City: city, CoverPhotoS3Key: s.CoverPhotoS3Key, Verified: s.Verified,
	}
}

// vacancyWithInstitutionDTO — вакансия + сводка об учреждении (глобальный список/деталь).
type vacancyWithInstitutionDTO struct {
	vacancyDTO
	Institution institutionSummaryDTO `json:"institution"`
}

func toVacancyWithInstitutionDTO(v domain.Vacancy, inst domain.InstitutionSummary) vacancyWithInstitutionDTO {
	return vacancyWithInstitutionDTO{vacancyDTO: toVacancyDTO(v), Institution: toInstitutionSummaryDTO(inst)}
}

// vacancyRequest — тело POST/PATCH запроса на вакансию.
type vacancyRequest struct {
	Title        bilingualDTO   `json:"title"`
	Description  bilingualDTO   `json:"description"`
	Requirements []bilingualDTO `json:"requirements,omitempty"`
	SalaryFrom   *int           `json:"salary_from,omitempty"`
	SalaryTo     *int           `json:"salary_to,omitempty"`
	Employment   bilingualDTO   `json:"employment"`
	Status       string         `json:"status"`
}

func (req vacancyRequest) toDomain() domain.VacancyInput {
	reqs := make([]domain.Bilingual, len(req.Requirements))
	for i, r := range req.Requirements {
		reqs[i] = fromBilingualDTO(r)
	}
	return domain.VacancyInput{
		Title: fromBilingualDTO(req.Title), Description: fromBilingualDTO(req.Description),
		Requirements: reqs, SalaryFrom: req.SalaryFrom, SalaryTo: req.SalaryTo,
		Employment: fromBilingualDTO(req.Employment), Status: req.Status,
	}
}

type listVacanciesResponse struct {
	Items []vacancyDTO `json:"items"`
}

type listVacanciesWithInstitutionResponse struct {
	Items []vacancyWithInstitutionDTO `json:"items"`
}

package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/applicants/domain"
)

type bilingualDTO struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func toBilingualDTO(b domain.Bilingual) bilingualDTO { return bilingualDTO{RU: b.RU, TG: b.TG} }
func fromBilingualDTO(b bilingualDTO) domain.Bilingual {
	return domain.Bilingual{RU: b.RU, TG: b.TG}
}
func toBilingualDTOPtr(b *domain.Bilingual) *bilingualDTO {
	if b == nil {
		return nil
	}
	dto := toBilingualDTO(*b)
	return &dto
}
func fromBilingualDTOPtr(b *bilingualDTO) *domain.Bilingual {
	if b == nil {
		return nil
	}
	d := fromBilingualDTO(*b)
	return &d
}
func toBilingualSliceDTO(items []domain.Bilingual) []bilingualDTO {
	out := make([]bilingualDTO, len(items))
	for i, b := range items {
		out[i] = toBilingualDTO(b)
	}
	return out
}
func fromBilingualSliceDTO(items []bilingualDTO) []domain.Bilingual {
	out := make([]domain.Bilingual, len(items))
	for i, b := range items {
		out[i] = fromBilingualDTO(b)
	}
	return out
}

// applicantDTO — публичный контракт профиля соискателя. Email/Phone заполняются, только если
// !HideContacts||Visibility=="public" — фильтрация в toApplicantDTO (единая точка маппинга),
// не отдельным SQL-проекцией шагом (упрощение относительно SRS E5.5 — небольшой MVP-каталог,
// без риска рассинхронизации логики в двух местах).
type applicantDTO struct {
	ID           uuid.UUID      `json:"id"`
	UserID       uuid.UUID      `json:"user_id"`
	Name         bilingualDTO   `json:"name"`
	PhotoURL     *string        `json:"photo_url,omitempty"`
	Position     bilingualDTO   `json:"position"`
	Bio          *bilingualDTO  `json:"bio,omitempty"`
	Education    []bilingualDTO `json:"education,omitempty"`
	Experience   []bilingualDTO `json:"experience,omitempty"`
	Skills       []bilingualDTO `json:"skills,omitempty"`
	Email        *string        `json:"email,omitempty"`
	Phone        *string        `json:"phone,omitempty"`
	CVS3Key      *string        `json:"cv_s3_key,omitempty"`
	Visibility   string         `json:"visibility"`
	HideContacts bool           `json:"hide_contacts"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// toApplicantDTO строит публичный DTO. mine==true — это собственный профиль вызывающего
// (контакты видны всегда, независимо от hide_contacts) — используется в /applicants/me.
func toApplicantDTO(a domain.Applicant, mine bool) applicantDTO {
	dto := applicantDTO{
		ID: a.ID, UserID: a.UserID, Name: toBilingualDTO(a.Name), PhotoURL: a.PhotoURL, Position: toBilingualDTO(a.Position),
		Bio: toBilingualDTOPtr(a.Bio), Education: toBilingualSliceDTO(a.Education), Experience: toBilingualSliceDTO(a.Experience),
		Skills: toBilingualSliceDTO(a.Skills), Visibility: a.Visibility, HideContacts: a.HideContacts,
		CVS3Key: a.CVS3Key, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
	showContacts := mine || !a.HideContacts || a.Visibility == domain.VisibilityPublic
	if showContacts {
		dto.Email, dto.Phone = a.Email, a.Phone
	}
	return dto
}

type listApplicantsResponse struct {
	Items []applicantDTO `json:"items"`
}

// applicantRequest — тело PUT /api/v1/applicants/me.
type applicantRequest struct {
	Name         bilingualDTO   `json:"name"`
	PhotoURL     *string        `json:"photo_url,omitempty"`
	Position     bilingualDTO   `json:"position"`
	Bio          *bilingualDTO  `json:"bio,omitempty"`
	Education    []bilingualDTO `json:"education,omitempty"`
	Experience   []bilingualDTO `json:"experience,omitempty"`
	Skills       []bilingualDTO `json:"skills,omitempty"`
	Email        *string        `json:"email,omitempty"`
	Phone        *string        `json:"phone,omitempty"`
	CVS3Key      *string        `json:"cv_s3_key,omitempty"`
	Visibility   string         `json:"visibility"`
	HideContacts bool           `json:"hide_contacts"`
}

func (req applicantRequest) toDomain() domain.ApplicantInput {
	return domain.ApplicantInput{
		Name: fromBilingualDTO(req.Name), PhotoURL: req.PhotoURL, Position: fromBilingualDTO(req.Position),
		Bio: fromBilingualDTOPtr(req.Bio), Education: fromBilingualSliceDTO(req.Education),
		Experience: fromBilingualSliceDTO(req.Experience), Skills: fromBilingualSliceDTO(req.Skills),
		Email: req.Email, Phone: req.Phone, CVS3Key: req.CVS3Key, Visibility: req.Visibility, HideContacts: req.HideContacts,
	}
}

// achievementDTO — достижение соискателя.
type achievementDTO struct {
	ID          uuid.UUID `json:"id"`
	ApplicantID uuid.UUID `json:"applicant_id"`
	Title       string    `json:"title"`
	Year        *int      `json:"year,omitempty"`
	Category    *string   `json:"category,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toAchievementDTO(a domain.Achievement) achievementDTO {
	return achievementDTO{ID: a.ID, ApplicantID: a.ApplicantID, Title: a.Title, Year: a.Year, Category: a.Category, Description: a.Description, CreatedAt: a.CreatedAt}
}

type listAchievementsResponse struct {
	Items []achievementDTO `json:"items"`
}

// achievementRequest — тело POST /api/v1/applicants/me/achievements.
type achievementRequest struct {
	Title       string  `json:"title"`
	Year        *int    `json:"year,omitempty"`
	Category    *string `json:"category,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (req achievementRequest) toDomain() domain.AchievementInput {
	return domain.AchievementInput{Title: req.Title, Year: req.Year, Category: req.Category, Description: req.Description}
}

// applicationDTO — отклик на вакансию.
type applicationDTO struct {
	ID          uuid.UUID `json:"id"`
	ApplicantID uuid.UUID `json:"applicant_id"`
	VacancyID   uuid.UUID `json:"vacancy_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func toApplicationDTO(a domain.Application) applicationDTO {
	return applicationDTO{ID: a.ID, ApplicantID: a.ApplicantID, VacancyID: a.VacancyID, Status: a.Status, CreatedAt: a.CreatedAt}
}

type listMyApplicationsResponse struct {
	VacancyIDs []uuid.UUID `json:"vacancy_ids"`
}

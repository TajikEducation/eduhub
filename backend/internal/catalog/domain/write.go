package domain

import (
	"strings"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// validRegions зеркалит CHECK-ограничение catalog.institutions.region (миграция 00002) —
// синтаксическая валидация здесь возвращает чистый 400 вместо голой ошибки БД.
var validRegions = map[string]bool{
	"dushanbe": true, "sughd": true, "khatlon": true, "gbao": true, "rrp": true,
}

// validInstitutionTypes — типы, которые бэкенд умеет сидировать/фильтровать (см.
// backend/cmd/devseed/data.go: cat_kg→kindergarten и т.д.).
var validInstitutionTypes = map[string]bool{
	"kindergarten": true, "school": true, "center": true, "university": true,
}

// CreateInstitutionInput — минимум для регистрации учреждения (E3.3), остальные поля — дефолты.
// Зеркалит RegisterInstitutionInput из web/lib/app-state.tsx.
type CreateInstitutionInput struct {
	Name        Bilingual
	Types       []string
	Region      string
	City        *Bilingual
	District    *string
	Description *Bilingual
	Phone       *string
	Email       *string
	Website     *string
	Price       *int
	Lat         float64
	Lng         float64
}

// Validate проверяет синтаксис входных данных регистрации учреждения.
func (in CreateInstitutionInput) Validate() error {
	fields := map[string]string{}

	if strings.TrimSpace(in.Name.RU) == "" || strings.TrimSpace(in.Name.TG) == "" {
		fields["name"] = "обязательное поле (ru и tg)"
	}

	if len(in.Types) == 0 {
		fields["types"] = "обязательное поле"
	} else {
		for _, t := range in.Types {
			if !validInstitutionTypes[t] {
				fields["types"] = "недопустимый тип: " + t
				break
			}
		}
	}

	if !validRegions[in.Region] {
		fields["region"] = "недопустимый регион"
	}

	if in.Lat < -90 || in.Lat > 90 {
		fields["lat"] = "должно быть в диапазоне [-90, 90]"
	}
	if in.Lng < -180 || in.Lng > 180 {
		fields["lng"] = "должно быть в диапазоне [-180, 180]"
	}

	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные регистрации учреждения")
	}
	return nil
}

// CreateStaffInput — данные педагога/персонала для institution_staff (create и update — одна
// форма, полная замена полей, не частичный patch; см. пакетный комментарий об урезанной
// семантике PATCH выше).
type CreateStaffInput struct {
	Name      Bilingual
	RoleType  string
	RoleLabel Bilingual
	Subject   *Bilingual
	PhotoURL  *string
	Exp       *string
	Bio       *Bilingual
	Email     *string
	Phone     *string
}

func (in CreateStaffInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Name.RU) == "" || strings.TrimSpace(in.Name.TG) == "" {
		fields["name"] = "обязательное поле (ru и tg)"
	}
	if strings.TrimSpace(in.RoleType) == "" {
		fields["role_type"] = "обязательное поле"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные сотрудника")
	}
	return nil
}

// CreateAchievementInput — данные достижения для catalog.achievements (owner_type/owner_id
// проставляет usecase — этот вход всегда для owner_type='institution').
type CreateAchievementInput struct {
	Title       Bilingual
	Year        int
	Category    string
	Description Bilingual
	Links       []AchievementLink
}

func (in CreateAchievementInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Title.RU) == "" || strings.TrimSpace(in.Title.TG) == "" {
		fields["title"] = "обязательное поле (ru и tg)"
	}
	if in.Year < 1900 || in.Year > 2100 {
		fields["year"] = "некорректный год"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные достижения")
	}
	return nil
}

// CreateGalleryItemInput — фото/видео для institution_gallery.
type CreateGalleryItemInput struct {
	S3Key     string
	Label     *Bilingual
	SortOrder int
}

func (in CreateGalleryItemInput) Validate() error {
	if strings.TrimSpace(in.S3Key) == "" {
		return apperr.Invalid(map[string]string{"s3_key": "обязательное поле"}, "некорректные данные галереи")
	}
	return nil
}

// CreateAlumnusInput — выпускник для institution_alumni.
type CreateAlumnusInput struct {
	Name     Bilingual
	PhotoURL *string
	GradYear int
	NowLabel *Bilingual
}

func (in CreateAlumnusInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Name.RU) == "" || strings.TrimSpace(in.Name.TG) == "" {
		fields["name"] = "обязательное поле (ru и tg)"
	}
	if in.GradYear < 1900 || in.GradYear > 2100 {
		fields["grad_year"] = "некорректный год"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные выпускника")
	}
	return nil
}

// validNewsStatuses зеркалит CHECK-ограничение catalog.news_articles.status.
var validNewsStatuses = map[string]bool{"draft": true, "published": true}

// CreateNewsInput — новость для catalog.news_articles.
type CreateNewsInput struct {
	Title      Bilingual
	Category   *Bilingual
	CoverS3Key *string
	VideoURL   *string
	Content    Bilingual
	Tags       []Bilingual
	Status     string
}

func (in CreateNewsInput) Validate() error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Title.RU) == "" || strings.TrimSpace(in.Title.TG) == "" {
		fields["title"] = "обязательное поле (ru и tg)"
	}
	if strings.TrimSpace(in.Content.RU) == "" || strings.TrimSpace(in.Content.TG) == "" {
		fields["content"] = "обязательное поле (ru и tg)"
	}
	if !validNewsStatuses[in.Status] {
		fields["status"] = "должно быть draft или published"
	}
	if len(fields) > 0 {
		return apperr.Invalid(fields, "некорректные данные новости")
	}
	return nil
}

// UpdateInstitutionInput — частичное обновление профиля учреждения (E3.4, урезанная версия:
// только nil="не трогать", без различения "не передано" от явного JSON null — см. SRS,
// полная семантика требует custom-unmarshal и отложена).
type UpdateInstitutionInput struct {
	Description     *Bilingual
	Phone           *string
	Email           *string
	Website         *string
	CoverPhotoS3Key *string
	Price           *int
	AgeRange        *string
}

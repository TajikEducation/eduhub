package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
)

// listResponse — тело ответа GET /api/v1/institutions.
type listResponse struct {
	Items      []institutionListItemDTO `json:"items"`
	NextCursor *string                  `json:"next_cursor"`
	// TotalHint всегда nil сейчас — keyset-пагинация не делает COUNT-запрос (задача 23).
	TotalHint *int `json:"total_hint"`
}

// bilingualDTO — двуязычное поле (ru/tg) в публичном JSON-контракте.
type bilingualDTO struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

// institutionListItemDTO — карточка учреждения в списке. Не отдаёт наружу профильные/владельческие
// поля (ModerationStatus/PlanExpiresAt/LicenseNo/Socials/LocationLandmarks/Phone/Email/Website)
// и сателлитные коллекции — List() их не заполняет.
type institutionListItemDTO struct {
	ID                uuid.UUID     `json:"id"`
	Name              bilingualDTO  `json:"name"`
	Types             []string      `json:"types"`
	Region            string        `json:"region"`
	City              *bilingualDTO `json:"city,omitempty"`
	District          *string       `json:"district,omitempty"`
	Price             *int          `json:"price,omitempty"`
	RatingAvg         *float64      `json:"rating_avg,omitempty"`
	ReviewCount       int           `json:"review_count"`
	Verified          bool          `json:"verified"`
	DiscountAvailable bool          `json:"discount_available"`
	CoverPhotoS3Key   *string       `json:"cover_photo_s3_key,omitempty"`
	Tag               *bilingualDTO `json:"tag,omitempty"`
	Lat               float64       `json:"lat"`
	Lng               float64       `json:"lng"`
	// DistanceM — omitempty критично: nil означает «гео в фильтре не запрашивался», не «0».
	DistanceM *float64 `json:"distance_m,omitempty"`
}

// toListResponse маппит доменный ListResult в публичный DTO-контракт.
func toListResponse(result domain.ListResult) listResponse {
	items := make([]institutionListItemDTO, len(result.Items))
	for i, inst := range result.Items {
		items[i] = toListItemDTO(inst)
	}
	return listResponse{Items: items, NextCursor: result.NextCursor, TotalHint: nil}
}

// toListItemDTO маппит одну Institution в карточку списка, поле в поле.
func toListItemDTO(inst domain.Institution) institutionListItemDTO {
	return institutionListItemDTO{
		ID:                inst.ID,
		Name:              toBilingualDTO(inst.Name),
		Types:             inst.Types,
		Region:            inst.Region,
		City:              toBilingualDTOPtr(inst.City),
		District:          inst.District,
		Price:             inst.Price,
		RatingAvg:         inst.RatingAvg,
		ReviewCount:       inst.ReviewCount,
		Verified:          inst.Verified,
		DiscountAvailable: inst.DiscountAvailable,
		CoverPhotoS3Key:   inst.CoverPhotoS3Key,
		Tag:               toBilingualDTOPtr(inst.Tag),
		Lat:               inst.Lat,
		Lng:               inst.Lng,
		DistanceM:         inst.DistanceM,
	}
}

func toBilingualDTO(b domain.Bilingual) bilingualDTO {
	return bilingualDTO{RU: b.RU, TG: b.TG}
}

func toBilingualDTOPtr(b *domain.Bilingual) *bilingualDTO {
	if b == nil {
		return nil
	}
	dto := toBilingualDTO(*b)
	return &dto
}

// socialsDTO — ссылки на соцсети учреждения в публичном JSON-контракте.
type socialsDTO struct {
	Instagram *string `json:"instagram,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	Facebook  *string `json:"facebook,omitempty"`
}

func toSocialsDTOPtr(s *domain.Socials) *socialsDTO {
	if s == nil {
		return nil
	}
	return &socialsDTO{Instagram: s.Instagram, Telegram: s.Telegram, Facebook: s.Facebook}
}

// staffMemberDTO — педагог/персонал учреждения в публичной карточке. InstitutionID заполняется
// только в ответе GET /api/v1/staff/{id} (публичный профиль сотрудника, страница /people/{id}
// на фронте) — при вложении в institutionDTO.Staff он избыточен (institution уже известна из URL).
type staffMemberDTO struct {
	ID            uuid.UUID     `json:"id"`
	InstitutionID uuid.UUID     `json:"institution_id,omitzero"`
	Name          bilingualDTO  `json:"name"`
	RoleType      string        `json:"role_type"`
	RoleLabel     bilingualDTO  `json:"role_label"`
	Subject       *bilingualDTO `json:"subject,omitempty"`
	PhotoURL      *string       `json:"photo_url,omitempty"`
	Exp           *string       `json:"exp,omitempty"`
	Bio           *bilingualDTO `json:"bio,omitempty"`
	Email         *string       `json:"email,omitempty"`
	Phone         *string       `json:"phone,omitempty"`
}

func toStaffMemberDTO(m domain.StaffMember) staffMemberDTO {
	return staffMemberDTO{
		ID:            m.ID,
		InstitutionID: m.InstitutionID,
		Name:          toBilingualDTO(m.Name),
		RoleType:      m.RoleType,
		RoleLabel:     toBilingualDTO(m.RoleLabel),
		Subject:       toBilingualDTOPtr(m.Subject),
		PhotoURL:      m.PhotoURL,
		Exp:           m.Exp,
		Bio:           toBilingualDTOPtr(m.Bio),
		Email:         m.Email,
		Phone:         m.Phone,
	}
}

// achievementLinkDTO — ссылка в описании достижения.
type achievementLinkDTO struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// achievementDTO — достижение учреждения/сотрудника/ученика в публичной карточке.
type achievementDTO struct {
	ID          uuid.UUID            `json:"id"`
	Title       bilingualDTO         `json:"title"`
	Year        int                  `json:"year"`
	Category    string               `json:"category"`
	Description bilingualDTO         `json:"description"`
	Links       []achievementLinkDTO `json:"links"`
}

func toAchievementDTO(a domain.Achievement) achievementDTO {
	links := make([]achievementLinkDTO, len(a.Links))
	for i, l := range a.Links {
		links[i] = achievementLinkDTO{Label: l.Label, URL: l.URL}
	}
	return achievementDTO{
		ID:          a.ID,
		Title:       toBilingualDTO(a.Title),
		Year:        a.Year,
		Category:    a.Category,
		Description: toBilingualDTO(a.Description),
		Links:       links,
	}
}

// galleryItemDTO — фото/видео учреждения в публичной карточке.
type galleryItemDTO struct {
	ID        uuid.UUID     `json:"id"`
	S3Key     string        `json:"s3_key"`
	Label     *bilingualDTO `json:"label,omitempty"`
	SortOrder int           `json:"sort_order"`
}

func toGalleryItemDTO(g domain.GalleryItem) galleryItemDTO {
	return galleryItemDTO{ID: g.ID, S3Key: g.S3Key, Label: toBilingualDTOPtr(g.Label), SortOrder: g.SortOrder}
}

// alumnusDTO — выпускник учреждения в публичной карточке.
type alumnusDTO struct {
	ID       uuid.UUID     `json:"id"`
	Name     bilingualDTO  `json:"name"`
	PhotoURL *string       `json:"photo_url,omitempty"`
	GradYear int           `json:"grad_year"`
	NowLabel *bilingualDTO `json:"now_label,omitempty"`
}

func toAlumnusDTO(a domain.Alumnus) alumnusDTO {
	return alumnusDTO{ID: a.ID, Name: toBilingualDTO(a.Name), PhotoURL: a.PhotoURL, GradYear: a.GradYear, NowLabel: toBilingualDTOPtr(a.NowLabel)}
}

// transportRouteDTO — маршрут развозки учреждения в публичной карточке.
type transportRouteDTO struct {
	ID         uuid.UUID      `json:"id"`
	Type       string         `json:"type"`
	Label      *bilingualDTO  `json:"label,omitempty"`
	Areas      []bilingualDTO `json:"areas"`
	Cost       *int           `json:"cost,omitempty"`
	CostPeriod string         `json:"cost_period"`
	SortOrder  int            `json:"sort_order"`
}

func toTransportRouteDTO(r domain.TransportRoute) transportRouteDTO {
	areas := make([]bilingualDTO, len(r.Areas))
	for i, a := range r.Areas {
		areas[i] = toBilingualDTO(a)
	}
	return transportRouteDTO{
		ID:         r.ID,
		Type:       r.Type,
		Label:      toBilingualDTOPtr(r.Label),
		Areas:      areas,
		Cost:       r.Cost,
		CostPeriod: r.CostPeriod,
		SortOrder:  r.SortOrder,
	}
}

// mealPlanDTO — вариант питания учреждения в публичной карточке.
type mealPlanDTO struct {
	ID         uuid.UUID     `json:"id"`
	MealType   string        `json:"meal_type"`
	Label      *bilingualDTO `json:"label,omitempty"`
	Cost       *int          `json:"cost,omitempty"`
	CostPeriod string        `json:"cost_period"`
	Halal      *bool         `json:"halal,omitempty"`
	SortOrder  int           `json:"sort_order"`
}

func toMealPlanDTO(m domain.MealPlan) mealPlanDTO {
	return mealPlanDTO{ID: m.ID, MealType: m.MealType, Label: toBilingualDTOPtr(m.Label), Cost: m.Cost, CostPeriod: m.CostPeriod, Halal: m.Halal, SortOrder: m.SortOrder}
}

// newsArticleDTO — новость учреждения в кабинете (E3.x, дашборд учреждения) и на публичной
// странице /news/{id} (там же используется InstitutionID для ссылки «назад к профилю»).
type newsArticleDTO struct {
	ID            uuid.UUID      `json:"id"`
	InstitutionID uuid.UUID      `json:"institution_id"`
	Title         bilingualDTO   `json:"title"`
	Category      *bilingualDTO  `json:"category,omitempty"`
	CoverS3Key    *string        `json:"cover_s3_key,omitempty"`
	VideoURL      *string        `json:"video_url,omitempty"`
	Content       bilingualDTO   `json:"content"`
	Tags          []bilingualDTO `json:"tags,omitempty"`
	Status        string         `json:"status"`
	ViewsCount    int            `json:"views_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func toNewsArticleDTO(n domain.NewsArticle) newsArticleDTO {
	tags := make([]bilingualDTO, len(n.Tags))
	for i, tg := range n.Tags {
		tags[i] = toBilingualDTO(tg)
	}
	return newsArticleDTO{
		ID: n.ID, InstitutionID: n.InstitutionID, Title: toBilingualDTO(n.Title), Category: toBilingualDTOPtr(n.Category),
		CoverS3Key: n.CoverS3Key, VideoURL: n.VideoURL, Content: toBilingualDTO(n.Content),
		Tags: tags, Status: n.Status, ViewsCount: n.ViewsCount, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}
}

// institutionDTO — полная карточка учреждения (профильная страница, FR-06). Сознательно не
// включает ModerationStatus (внутреннее, Service.Get гарантирует только approved), Plan/PlanExpiresAt
// (биллинговые) и News (репозиторий её пока не заполняет — задача 24).
type institutionDTO struct {
	ID                uuid.UUID     `json:"id"`
	Name              bilingualDTO  `json:"name"`
	Types             []string      `json:"types"`
	Region            string        `json:"region"`
	City              *bilingualDTO `json:"city,omitempty"`
	District          *string       `json:"district,omitempty"`
	Description       *bilingualDTO `json:"description,omitempty"`
	Address           *bilingualDTO `json:"address,omitempty"`
	Lat               float64       `json:"lat"`
	Lng               float64       `json:"lng"`
	LocationLandmarks *string       `json:"location_landmarks,omitempty"`
	Phone             *string       `json:"phone,omitempty"`
	Email             *string       `json:"email,omitempty"`
	Website           *string       `json:"website,omitempty"`
	Socials           *socialsDTO   `json:"socials,omitempty"`
	CoverPhotoS3Key   *string       `json:"cover_photo_s3_key,omitempty"`
	AgeRange          *string       `json:"age_range,omitempty"`
	Tag               *bilingualDTO `json:"tag,omitempty"`
	LicenseNo         *string       `json:"license_no,omitempty"`
	Languages         []string      `json:"languages,omitempty"`
	ProgramLevel      []string      `json:"program_level,omitempty"`
	Curriculum        []string      `json:"curriculum,omitempty"`
	Price             *int          `json:"price,omitempty"`
	DiscountAvailable bool          `json:"discount_available"`
	DiscountType      []string      `json:"discount_type,omitempty"`
	DiscountDetails   *string       `json:"discount_details,omitempty"`
	Verified          bool          `json:"verified"`
	Founded           *int          `json:"founded,omitempty"`
	StudentsCount     *int          `json:"students_count,omitempty"`
	RatingAvg         *float64      `json:"rating_avg,omitempty"`
	ReviewCount       int           `json:"review_count"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`

	Staff           []staffMemberDTO    `json:"staff"`
	Achievements    []achievementDTO    `json:"achievements"`
	Gallery         []galleryItemDTO    `json:"gallery"`
	Alumni          []alumnusDTO        `json:"alumni"`
	TransportRoutes []transportRouteDTO `json:"transport_routes"`
	MealPlans       []mealPlanDTO       `json:"meal_plans"`
}

// toInstitutionDTO маппит полную Institution в публичную карточку, поле в поле.
func toInstitutionDTO(inst domain.Institution) institutionDTO {
	staff := make([]staffMemberDTO, len(inst.Staff))
	for i, m := range inst.Staff {
		staff[i] = toStaffMemberDTO(m)
	}
	achievements := make([]achievementDTO, len(inst.Achievements))
	for i, a := range inst.Achievements {
		achievements[i] = toAchievementDTO(a)
	}
	gallery := make([]galleryItemDTO, len(inst.Gallery))
	for i, g := range inst.Gallery {
		gallery[i] = toGalleryItemDTO(g)
	}
	alumni := make([]alumnusDTO, len(inst.Alumni))
	for i, a := range inst.Alumni {
		alumni[i] = toAlumnusDTO(a)
	}
	transportRoutes := make([]transportRouteDTO, len(inst.TransportRoutes))
	for i, r := range inst.TransportRoutes {
		transportRoutes[i] = toTransportRouteDTO(r)
	}
	mealPlans := make([]mealPlanDTO, len(inst.MealPlans))
	for i, m := range inst.MealPlans {
		mealPlans[i] = toMealPlanDTO(m)
	}

	return institutionDTO{
		ID:                inst.ID,
		Name:              toBilingualDTO(inst.Name),
		Types:             inst.Types,
		Region:            inst.Region,
		City:              toBilingualDTOPtr(inst.City),
		District:          inst.District,
		Description:       toBilingualDTOPtr(inst.Description),
		Address:           toBilingualDTOPtr(inst.Address),
		Lat:               inst.Lat,
		Lng:               inst.Lng,
		LocationLandmarks: inst.LocationLandmarks,
		Phone:             inst.Phone,
		Email:             inst.Email,
		Website:           inst.Website,
		Socials:           toSocialsDTOPtr(inst.Socials),
		CoverPhotoS3Key:   inst.CoverPhotoS3Key,
		AgeRange:          inst.AgeRange,
		Tag:               toBilingualDTOPtr(inst.Tag),
		LicenseNo:         inst.LicenseNo,
		Languages:         inst.Languages,
		ProgramLevel:      inst.ProgramLevel,
		Curriculum:        inst.Curriculum,
		Price:             inst.Price,
		DiscountAvailable: inst.DiscountAvailable,
		DiscountType:      inst.DiscountType,
		DiscountDetails:   inst.DiscountDetails,
		Verified:          inst.Verified,
		Founded:           inst.Founded,
		StudentsCount:     inst.StudentsCount,
		RatingAvg:         inst.RatingAvg,
		ReviewCount:       inst.ReviewCount,
		CreatedAt:         inst.CreatedAt,
		UpdatedAt:         inst.UpdatedAt,

		Staff:           staff,
		Achievements:    achievements,
		Gallery:         gallery,
		Alumni:          alumni,
		TransportRoutes: transportRoutes,
		MealPlans:       mealPlans,
	}
}

// moderationListItemDTO — карточка институции в очереди модератора/обзоре админа
// (GET /api/v1/moderation/institutions). В отличие от institutionListItemDTO, публичного каталога,
// включает ModerationStatus и профильные поля (адрес/телефон/email/описание) — этот эндпоинт
// доступен только moderator/admin (rbac.RequireRole), не публичному каталогу.
type moderationListItemDTO struct {
	ID               uuid.UUID     `json:"id"`
	Name             bilingualDTO  `json:"name"`
	Types            []string      `json:"types"`
	Region           string        `json:"region"`
	Address          *bilingualDTO `json:"address,omitempty"`
	Phone            *string       `json:"phone,omitempty"`
	Email            *string       `json:"email,omitempty"`
	Description      *bilingualDTO `json:"description,omitempty"`
	Price            *int          `json:"price,omitempty"`
	ModerationStatus string        `json:"moderation_status"`
	CreatedAt        time.Time     `json:"created_at"`
}

func toModerationListItemDTO(inst domain.Institution) moderationListItemDTO {
	return moderationListItemDTO{
		ID: inst.ID, Name: toBilingualDTO(inst.Name), Types: inst.Types, Region: inst.Region,
		Address: toBilingualDTOPtr(inst.Address), Phone: inst.Phone, Email: inst.Email,
		Description: toBilingualDTOPtr(inst.Description), Price: inst.Price,
		ModerationStatus: inst.ModerationStatus, CreatedAt: inst.CreatedAt,
	}
}

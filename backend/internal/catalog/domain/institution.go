package domain

import (
	"time"

	"github.com/google/uuid"
)

// Socials — ссылки на соцсети учреждения, зеркалит JSONB {instagram?, telegram?, facebook?}.
type Socials struct {
	Instagram *string
	Telegram  *string
	Facebook  *string
}

// StaffMember — педагог/персонал учреждения (catalog.institution_staff).
type StaffMember struct {
	ID        uuid.UUID
	Name      Bilingual
	RoleType  string
	RoleLabel Bilingual
	Subject   *Bilingual
	PhotoURL  *string
	Exp       *string
	Bio       *Bilingual
	Education []Bilingual
	Email     *string
	Phone     *string
	CreatedAt time.Time
}

// AchievementLink — ссылка в описании достижения ({label,url} в JSONB links).
type AchievementLink struct {
	Label string
	URL   string
}

// Achievement — достижение учреждения/сотрудника/ученика (catalog.achievements, полиморфная связь).
type Achievement struct {
	ID          uuid.UUID
	OwnerType   string
	OwnerID     uuid.UUID
	Title       Bilingual
	Year        int
	Category    string
	Description Bilingual
	Links       []AchievementLink
	CreatedAt   time.Time
}

// GalleryItem — фото/видео учреждения (catalog.institution_gallery).
type GalleryItem struct {
	ID        uuid.UUID
	S3Key     string
	Label     *Bilingual
	SortOrder int
	CreatedAt time.Time
}

// Alumnus — выпускник учреждения (catalog.institution_alumni).
type Alumnus struct {
	ID        uuid.UUID
	Name      Bilingual
	PhotoURL  *string
	GradYear  int
	NowLabel  *Bilingual
	CreatedAt time.Time
}

// TransportRoute — маршрут развозки учреждения (catalog.institution_transport_routes).
type TransportRoute struct {
	ID         uuid.UUID
	Type       string
	Label      *Bilingual
	Areas      []Bilingual
	Cost       *int
	CostPeriod string
	SortOrder  int
	CreatedAt  time.Time
}

// MealPlan — вариант питания учреждения (catalog.institution_meal_plans).
type MealPlan struct {
	ID         uuid.UUID
	MealType   string
	Label      *Bilingual
	Cost       *int
	CostPeriod string
	Halal      *bool
	SortOrder  int
	CreatedAt  time.Time
}

// NewsArticle — новость учреждения (catalog.news_articles).
type NewsArticle struct {
	ID         uuid.UUID
	Title      Bilingual
	Category   *Bilingual
	CoverS3Key *string
	VideoURL   *string
	Content    Bilingual
	Tags       []Bilingual
	Status     string
	ViewsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Institution — карточка учреждения, зеркалит catalog.institutions + вложенные коллекции-сателлиты.
type Institution struct {
	ID                uuid.UUID
	Name              Bilingual
	Types             []string
	Region            string
	City              *Bilingual
	District          *string
	Description       *Bilingual
	Address           *Bilingual
	Lat               float64
	Lng               float64
	LocationLandmarks *string
	Phone             *string
	Email             *string
	Website           *string
	Socials           *Socials
	CoverPhotoS3Key   *string
	AgeRange          *string
	Tag               *Bilingual
	LicenseNo         *string
	Languages         []string
	ProgramLevel      []string
	Curriculum        []string
	Price             *int
	DiscountAvailable bool
	DiscountType      []string
	DiscountDetails   *string
	Verified          bool
	ModerationStatus  string
	Plan              string
	PlanExpiresAt     *time.Time
	Founded           *int
	StudentsCount     *int
	RatingAvg         *float64
	ReviewCount       int
	CreatedAt         time.Time
	UpdatedAt         time.Time

	Staff           []StaffMember
	Achievements    []Achievement
	Gallery         []GalleryItem
	Alumni          []Alumnus
	TransportRoutes []TransportRoute
	MealPlans       []MealPlan
	News            []NewsArticle
}

// NewInstitution создаёт Institution с гарантированно non-nil коллекциями/слайсами
// (JSON-контракт: "gallery":[], не "gallery":null) — остальные поля repo-слой
// заполняет напрямую после конструктора (экспортируемые поля) и сателлиты (задача 24).
func NewInstitution(id uuid.UUID, name Bilingual, region string) *Institution {
	return &Institution{
		ID:     id,
		Name:   name,
		Region: region,

		Types:        []string{},
		Languages:    []string{},
		ProgramLevel: []string{},
		Curriculum:   []string{},
		DiscountType: []string{},

		Staff:           []StaffMember{},
		Achievements:    []Achievement{},
		Gallery:         []GalleryItem{},
		Alumni:          []Alumnus{},
		TransportRoutes: []TransportRoute{},
		MealPlans:       []MealPlan{},
		News:            []NewsArticle{},
	}
}

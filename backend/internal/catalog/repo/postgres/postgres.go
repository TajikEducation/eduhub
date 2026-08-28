// Package postgres — реализация internal/catalog/usecase.InstitutionRepo поверх PostgreSQL.
package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// querier — минимальный интерфейс доступа к БД, которому удовлетворяют и *pgxpool.Pool,
// и pgx.Tx. Нужен для тестируемости через транзакцию+rollback: если бы репозиторий требовал
// конкретно *pgxpool.Pool, фикстуры, вставленные в транзакцию теста, были бы не видны
// репозиторию (отдельное подключение из пула = отдельная сессия БД).
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// InstitutionRepo — репозиторий каталога институций поверх PostgreSQL.
type InstitutionRepo struct {
	db querier
}

// New создаёт InstitutionRepo поверх переданного querier (пул подключений или транзакция).
func New(db querier) *InstitutionRepo {
	return &InstitutionRepo{db: db}
}

// listColumns — скалярные колонки catalog.institutions в порядке, соответствующем scanInstitution.
// Без JOIN на сателлиты (задача 24). geo раскладывается на lat/lng через ST_Y/ST_X.
const listColumns = `
	id, name, types, region, city, district, description, address,
	ST_Y(geo::geometry), ST_X(geo::geometry), location_landmarks,
	phone, email, website, socials, cover_photo_s3_key, age_range, tag,
	license_no, languages, program_level, curriculum, price,
	discount_available, discount_type, discount_details, verified,
	moderation_status, plan, plan_expires_at, founded, students_count,
	rating_avg, review_count, created_at, updated_at
`

// cursorPayload — непрозрачное содержимое курсора keyset-пагинации, сериализуется в
// base64(JSON) и не должно интерпретироваться клиентом.
type cursorPayload struct {
	Sort      string  `json:"sort"`
	LastValue float64 `json:"last_value"` // цена или rating_avg — оба помещаются в float64
	LastID    string  `json:"last_id"`    // uuid.String()
}

// encodeCursor сериализует позицию последнего элемента страницы в непрозрачный курсор.
func encodeCursor(sort string, lastValue float64, lastID uuid.UUID) string {
	raw, _ := json.Marshal(cursorPayload{Sort: sort, LastValue: lastValue, LastID: lastID.String()})
	return base64.StdEncoding.EncodeToString(raw)
}

// decodeCursor разбирает непрозрачный курсор обратно в cursorPayload. Любая ошибка формата
// возвращается как доменная apperr.Invalid — курсор приходит от клиента, не должен приводить
// к панике/500.
func decodeCursor(raw string) (cursorPayload, error) {
	invalid := func() (cursorPayload, error) {
		return cursorPayload{}, apperr.Invalid(
			map[string]string{"cursor": "некорректный курсор пагинации"},
			"курсор не удалось разобрать",
		)
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return invalid()
	}

	var payload cursorPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return invalid()
	}
	if payload.Sort == "" || payload.LastID == "" {
		return invalid()
	}

	return payload, nil
}

// List возвращает институции, удовлетворяющие фильтру f, с keyset-пагинацией по f.Cursor/f.Limit.
func (r *InstitutionRepo) List(ctx context.Context, f domain.Filter) (domain.ListResult, error) {
	var conditions []string
	var args []any

	// addCond добавляет условие вида "col = $n", подставляя очередной позиционный плейсхолдер.
	addCond := func(condFmt string, arg any) {
		args = append(args, arg)
		conditions = append(conditions, fmt.Sprintf(condFmt, len(args)))
	}

	if len(f.Statuses) > 0 {
		addCond("moderation_status = ANY($%d)", f.Statuses)
	}
	if f.Region != nil {
		addCond("region = $%d", *f.Region)
	}
	if len(f.Types) > 0 {
		addCond("types && $%d", f.Types)
	}
	if f.Q != nil && *f.Q != "" {
		args = append(args, "%"+*f.Q+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(name->>'ru' ILIKE $%d OR name->>'tg' ILIKE $%d)", n, n))
	}
	if f.MinRating != nil {
		// NULL rating_avg естественно исключается 3-значной SQL-логикой (NULL >= x → UNKNOWN).
		addCond("rating_avg >= $%d", *f.MinRating)
	}
	if len(f.Curriculum) > 0 {
		addCond("curriculum && $%d", f.Curriculum)
	}
	if len(f.ProgramLevel) > 0 {
		addCond("program_level && $%d", f.ProgramLevel)
	}
	if f.Discount != nil {
		addCond("discount_available = $%d", *f.Discount)
	}
	if f.MinPrice != nil {
		addCond("price >= $%d", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		addCond("price <= $%d", *f.MaxPrice)
	}
	if f.Verified != nil {
		addCond("verified = $%d", *f.Verified)
	}

	if f.Cursor != nil && *f.Cursor != "" {
		payload, err := decodeCursor(*f.Cursor)
		if err != nil {
			return domain.ListResult{}, err
		}
		if payload.Sort != f.Sort {
			return domain.ListResult{}, apperr.Invalid(
				map[string]string{"cursor": "sort не совпадает с текущим запросом"},
				"курсор выдан для другого значения sort",
			)
		}

		lastID, err := uuid.Parse(payload.LastID)
		if err != nil {
			return domain.ListResult{}, apperr.Invalid(
				map[string]string{"cursor": "некорректный курсор пагинации"},
				"курсор содержит некорректный id",
			)
		}

		switch f.Sort {
		case "price_asc":
			args = append(args, int(payload.LastValue), lastID)
			valN, idN := len(args)-1, len(args)
			conditions = append(conditions, fmt.Sprintf("(price, id) > ($%d, $%d)", valN, idN))
		case "score":
			args = append(args, payload.LastValue, lastID)
			valN, idN := len(args)-1, len(args)
			conditions = append(conditions, fmt.Sprintf("(rating_avg, id) < ($%d, $%d)", valN, idN))
		}
	}

	geoSet := f.Lat != nil && f.Lng != nil
	if geoSet && f.RadiusKm != nil {
		args = append(args, *f.Lng, *f.Lat, *f.RadiusKm*1000)
		lngN, latN, radiusN := len(args)-2, len(args)-1, len(args)
		conditions = append(conditions, fmt.Sprintf(
			"ST_DWithin(geo, ST_MakePoint($%d,$%d)::geography, $%d)", lngN, latN, radiusN,
		))
	}

	// distanceExpr — вычисляемая колонка расстояния в SELECT. При отсутствии гео в фильтре
	// $n/$n+1 всё равно попадут в args, но выражение всегда возвращает NULL (CASE WHEN),
	// поэтому Scan в *float64 естественно даёт nil — контракт DistanceM (см. domain.Institution).
	distanceExpr := "NULL::float8"
	if geoSet {
		args = append(args, *f.Lng, *f.Lat)
		lngN, latN := len(args)-1, len(args)
		distanceExpr = fmt.Sprintf("ST_Distance(geo, ST_MakePoint($%d,$%d)::geography)", lngN, latN)
	}

	query := "SELECT " + listColumns + ", " + distanceExpr + " FROM catalog.institutions"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch f.Sort {
	case "price_asc":
		query += " ORDER BY price ASC NULLS LAST, id"
	case "score":
		query += " ORDER BY rating_avg DESC NULLS LAST, id"
	default:
		if geoSet {
			args = append(args, *f.Lng, *f.Lat)
			lngN, latN := len(args)-1, len(args)
			query += fmt.Sprintf(" ORDER BY geo <-> ST_MakePoint($%d,$%d)::geography, id", lngN, latN)
		}
	}

	// LIMIT f.Limit+1 — запрашиваем на одну строку больше, чтобы узнать, есть ли следующая
	// страница, без отдельного COUNT-запроса. f.Limit==0 (не нормализован вызывающей стороной,
	// как в прямых репозиторных тестах без пагинации) — лимит не применяется, ведём себя как
	// до задачи 23.
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit+1)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return domain.ListResult{}, fmt.Errorf("postgres: list institutions: %w", err)
	}
	defer rows.Close()

	var out []domain.Institution
	for rows.Next() {
		inst, err := scanInstitution(rows)
		if err != nil {
			return domain.ListResult{}, fmt.Errorf("postgres: scan institution: %w", err)
		}
		out = append(out, *inst)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, fmt.Errorf("postgres: list institutions rows: %w", err)
	}

	result := domain.ListResult{Items: out}
	if f.Limit > 0 && len(out) > f.Limit {
		result.Items = out[:f.Limit]

		last := result.Items[len(result.Items)-1]
		switch f.Sort {
		case "price_asc":
			if last.Price != nil {
				cursor := encodeCursor(f.Sort, float64(*last.Price), last.ID)
				result.NextCursor = &cursor
			}
		case "score":
			if last.RatingAvg != nil {
				cursor := encodeCursor(f.Sort, *last.RatingAvg, last.ID)
				result.NextCursor = &cursor
			}
		}
	}

	return result, nil
}

// getByIDQuery — карточка институции одним запросом: скалярные колонки catalog.institutions
// (те же, что listColumns) + 6 JSON-агрегатов сателлитных коллекций через LATERAL-подобные
// подзапросы (json_agg(row_to_json(...))). ORDER BY внутри json_agg(...) — единственное место,
// где он строго обязателен для стабильного порядка; во вложенном SELECT дублируется для ясности.
const getByIDQuery = `
	SELECT ` + listColumns + `,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, name, role_type, role_label, subject, photo_url, exp, bio, education, email, phone, created_at
			FROM catalog.institution_staff
			WHERE institution_id = i.id
			ORDER BY created_at, id
		 ) t
		) AS staff,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, owner_type, owner_id, title, year, category, description, links, created_at
			FROM catalog.achievements
			WHERE owner_type = 'institution' AND owner_id = i.id
			ORDER BY created_at, id
		 ) t
		) AS achievements,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.sort_order, t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, s3_key, label, sort_order, created_at
			FROM catalog.institution_gallery
			WHERE institution_id = i.id
			ORDER BY sort_order, created_at, id
		 ) t
		) AS gallery,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, name, photo_url, grad_year, now_label, created_at
			FROM catalog.institution_alumni
			WHERE institution_id = i.id
			ORDER BY created_at, id
		 ) t
		) AS alumni,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.sort_order, t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, type, label, areas, cost, cost_period, sort_order, created_at
			FROM catalog.institution_transport_routes
			WHERE institution_id = i.id
			ORDER BY sort_order, created_at, id
		 ) t
		) AS transport_routes,
		(SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.sort_order, t.created_at, t.id), '[]'::json)
		 FROM (
			SELECT id, meal_type, label, cost, cost_period, halal, sort_order, created_at
			FROM catalog.institution_meal_plans
			WHERE institution_id = i.id
			ORDER BY sort_order, created_at, id
		 ) t
		) AS meal_plans
	FROM catalog.institutions i
	WHERE i.id = $1
`

// GetByID возвращает полную карточку институции (скаляры + 6 сателлитных коллекций) одним
// SQL-запросом. apperr.ErrNotFound — если институция с id не найдена.
func (r *InstitutionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Institution, error) {
	row := r.db.QueryRow(ctx, getByIDQuery, id)

	inst, err := scanInstitutionFull(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Institution{}, apperr.NotFound("institution", id.String())
	}
	if err != nil {
		return domain.Institution{}, fmt.Errorf("postgres: get institution by id: %w", err)
	}

	return *inst, nil
}

// bilingualWire — форма вложенного JSON-объекта {ru,tg} внутри агрегатов сателлитных коллекций
// (не путать с bilingualJSON, который разбирает колонки самой catalog.institutions).
type bilingualWire struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func (b bilingualWire) toDomain() domain.Bilingual {
	return domain.Bilingual{RU: b.RU, TG: b.TG}
}

// bilingualWireSlice конвертирует срез вложенных {ru,tg}-объектов в domain.Bilingual.
func bilingualWireSlice(in []bilingualWire) []domain.Bilingual {
	out := make([]domain.Bilingual, len(in))
	for i, v := range in {
		out[i] = v.toDomain()
	}
	return out
}

// bilingualWirePtr конвертирует nullable вложенный {ru,tg}-объект в *domain.Bilingual.
func bilingualWirePtr(in *bilingualWire) *domain.Bilingual {
	if in == nil {
		return nil
	}
	b := in.toDomain()
	return &b
}

// staffWire — форма одного элемента агрегата staff.
type staffWire struct {
	ID        uuid.UUID       `json:"id"`
	Name      bilingualWire   `json:"name"`
	RoleType  string          `json:"role_type"`
	RoleLabel bilingualWire   `json:"role_label"`
	Subject   *bilingualWire  `json:"subject"`
	PhotoURL  *string         `json:"photo_url"`
	Exp       *string         `json:"exp"`
	Bio       *bilingualWire  `json:"bio"`
	Education []bilingualWire `json:"education"`
	Email     *string         `json:"email"`
	Phone     *string         `json:"phone"`
	CreatedAt time.Time       `json:"created_at"`
}

func (w staffWire) toDomain() domain.StaffMember {
	return domain.StaffMember{
		ID:        w.ID,
		Name:      w.Name.toDomain(),
		RoleType:  w.RoleType,
		RoleLabel: w.RoleLabel.toDomain(),
		Subject:   bilingualWirePtr(w.Subject),
		PhotoURL:  w.PhotoURL,
		Exp:       w.Exp,
		Bio:       bilingualWirePtr(w.Bio),
		Education: bilingualWireSlice(w.Education),
		Email:     w.Email,
		Phone:     w.Phone,
		CreatedAt: w.CreatedAt,
	}
}

// achievementLinkWire — форма одной ссылки во вложенном JSON-массиве links.
type achievementLinkWire struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// achievementWire — форма одного элемента агрегата achievements.
type achievementWire struct {
	ID          uuid.UUID             `json:"id"`
	OwnerType   string                `json:"owner_type"`
	OwnerID     uuid.UUID             `json:"owner_id"`
	Title       bilingualWire         `json:"title"`
	Year        int                   `json:"year"`
	Category    string                `json:"category"`
	Description bilingualWire         `json:"description"`
	Links       []achievementLinkWire `json:"links"`
	CreatedAt   time.Time             `json:"created_at"`
}

func (w achievementWire) toDomain() domain.Achievement {
	links := make([]domain.AchievementLink, len(w.Links))
	for i, l := range w.Links {
		links[i] = domain.AchievementLink{Label: l.Label, URL: l.URL}
	}
	return domain.Achievement{
		ID:          w.ID,
		OwnerType:   w.OwnerType,
		OwnerID:     w.OwnerID,
		Title:       w.Title.toDomain(),
		Year:        w.Year,
		Category:    w.Category,
		Description: w.Description.toDomain(),
		Links:       links,
		CreatedAt:   w.CreatedAt,
	}
}

// galleryItemWire — форма одного элемента агрегата gallery.
type galleryItemWire struct {
	ID        uuid.UUID      `json:"id"`
	S3Key     string         `json:"s3_key"`
	Label     *bilingualWire `json:"label"`
	SortOrder int            `json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
}

func (w galleryItemWire) toDomain() domain.GalleryItem {
	return domain.GalleryItem{
		ID:        w.ID,
		S3Key:     w.S3Key,
		Label:     bilingualWirePtr(w.Label),
		SortOrder: w.SortOrder,
		CreatedAt: w.CreatedAt,
	}
}

// alumnusWire — форма одного элемента агрегата alumni.
type alumnusWire struct {
	ID        uuid.UUID      `json:"id"`
	Name      bilingualWire  `json:"name"`
	PhotoURL  *string        `json:"photo_url"`
	GradYear  int            `json:"grad_year"`
	NowLabel  *bilingualWire `json:"now_label"`
	CreatedAt time.Time      `json:"created_at"`
}

func (w alumnusWire) toDomain() domain.Alumnus {
	return domain.Alumnus{
		ID:        w.ID,
		Name:      w.Name.toDomain(),
		PhotoURL:  w.PhotoURL,
		GradYear:  w.GradYear,
		NowLabel:  bilingualWirePtr(w.NowLabel),
		CreatedAt: w.CreatedAt,
	}
}

// transportRouteWire — форма одного элемента агрегата transport_routes.
type transportRouteWire struct {
	ID         uuid.UUID       `json:"id"`
	Type       string          `json:"type"`
	Label      *bilingualWire  `json:"label"`
	Areas      []bilingualWire `json:"areas"`
	Cost       *int            `json:"cost"`
	CostPeriod string          `json:"cost_period"`
	SortOrder  int             `json:"sort_order"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (w transportRouteWire) toDomain() domain.TransportRoute {
	return domain.TransportRoute{
		ID:         w.ID,
		Type:       w.Type,
		Label:      bilingualWirePtr(w.Label),
		Areas:      bilingualWireSlice(w.Areas),
		Cost:       w.Cost,
		CostPeriod: w.CostPeriod,
		SortOrder:  w.SortOrder,
		CreatedAt:  w.CreatedAt,
	}
}

// mealPlanWire — форма одного элемента агрегата meal_plans.
type mealPlanWire struct {
	ID         uuid.UUID      `json:"id"`
	MealType   string         `json:"meal_type"`
	Label      *bilingualWire `json:"label"`
	Cost       *int           `json:"cost"`
	CostPeriod string         `json:"cost_period"`
	Halal      *bool          `json:"halal"`
	SortOrder  int            `json:"sort_order"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (w mealPlanWire) toDomain() domain.MealPlan {
	return domain.MealPlan{
		ID:         w.ID,
		MealType:   w.MealType,
		Label:      bilingualWirePtr(w.Label),
		Cost:       w.Cost,
		CostPeriod: w.CostPeriod,
		Halal:      w.Halal,
		SortOrder:  w.SortOrder,
		CreatedAt:  w.CreatedAt,
	}
}

// scanInstitutionFull сканирует одну строку getByIDQuery (скаляры catalog.institutions
// в порядке listColumns, плюс 6 JSON-агрегатов сателлитных коллекций последними) в
// domain.Institution. pgx.Row (не pgx.Rows) — .Scan(...) сразу, без .Next().
func scanInstitutionFull(row pgx.Row) (*domain.Institution, error) {
	var (
		id                                         uuid.UUID
		nameRaw, cityRaw, descRaw, addrRaw         []byte
		socialsRaw, tagRaw                         []byte
		region                                     string
		district                                   *string
		lat, lng                                   float64
		locationLandmarks, phone, email, website   *string
		coverPhotoS3Key, ageRange, licenseNo       *string
		types, languages, programLevel, curriculum []string
		price                                      *int
		discountAvailable                          bool
		discountType                               []string
		discountDetails                            *string
		verified                                   bool
		moderationStatus, plan                     string
		planExpiresAt                              *time.Time
		founded, studentsCount                     *int
		ratingAvg                                  *float64
		reviewCount                                int
		createdAt, updatedAt                       time.Time

		staffRaw, achievementsRaw, galleryRaw   []byte
		alumniRaw, transportRoutesRaw, mealsRaw []byte
	)

	if err := row.Scan(
		&id, &nameRaw, &types, &region, &cityRaw, &district, &descRaw, &addrRaw,
		&lat, &lng, &locationLandmarks,
		&phone, &email, &website, &socialsRaw, &coverPhotoS3Key, &ageRange, &tagRaw,
		&licenseNo, &languages, &programLevel, &curriculum, &price,
		&discountAvailable, &discountType, &discountDetails, &verified,
		&moderationStatus, &plan, &planExpiresAt, &founded, &studentsCount,
		&ratingAvg, &reviewCount, &createdAt, &updatedAt,
		&staffRaw, &achievementsRaw, &galleryRaw, &alumniRaw, &transportRoutesRaw, &mealsRaw,
	); err != nil {
		return nil, err
	}

	name, err := scanBilingual(nameRaw)
	if err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}

	inst := domain.NewInstitution(id, name, region)
	inst.Types = types

	if inst.City, err = scanBilingualPtr(cityRaw); err != nil {
		return nil, fmt.Errorf("city: %w", err)
	}
	inst.District = district
	if inst.Description, err = scanBilingualPtr(descRaw); err != nil {
		return nil, fmt.Errorf("description: %w", err)
	}
	if inst.Address, err = scanBilingualPtr(addrRaw); err != nil {
		return nil, fmt.Errorf("address: %w", err)
	}
	inst.Lat = lat
	inst.Lng = lng
	inst.LocationLandmarks = locationLandmarks
	inst.Phone = phone
	inst.Email = email
	inst.Website = website
	if inst.Socials, err = scanSocials(socialsRaw); err != nil {
		return nil, fmt.Errorf("socials: %w", err)
	}
	inst.CoverPhotoS3Key = coverPhotoS3Key
	inst.AgeRange = ageRange
	if inst.Tag, err = scanBilingualPtr(tagRaw); err != nil {
		return nil, fmt.Errorf("tag: %w", err)
	}
	inst.LicenseNo = licenseNo
	inst.Languages = nonNilStrings(languages)
	inst.ProgramLevel = nonNilStrings(programLevel)
	inst.Curriculum = nonNilStrings(curriculum)
	inst.Price = price
	inst.DiscountAvailable = discountAvailable
	inst.DiscountType = nonNilStrings(discountType)
	inst.DiscountDetails = discountDetails
	inst.Verified = verified
	inst.ModerationStatus = moderationStatus
	inst.Plan = plan
	inst.PlanExpiresAt = planExpiresAt
	inst.Founded = founded
	inst.StudentsCount = studentsCount
	inst.RatingAvg = ratingAvg
	inst.ReviewCount = reviewCount
	inst.CreatedAt = createdAt
	inst.UpdatedAt = updatedAt

	var staffWires []staffWire
	if err := json.Unmarshal(staffRaw, &staffWires); err != nil {
		return nil, fmt.Errorf("staff: %w", err)
	}
	inst.Staff = make([]domain.StaffMember, len(staffWires))
	for i, w := range staffWires {
		inst.Staff[i] = w.toDomain()
	}

	var achievementWires []achievementWire
	if err := json.Unmarshal(achievementsRaw, &achievementWires); err != nil {
		return nil, fmt.Errorf("achievements: %w", err)
	}
	inst.Achievements = make([]domain.Achievement, len(achievementWires))
	for i, w := range achievementWires {
		inst.Achievements[i] = w.toDomain()
	}

	var galleryWires []galleryItemWire
	if err := json.Unmarshal(galleryRaw, &galleryWires); err != nil {
		return nil, fmt.Errorf("gallery: %w", err)
	}
	inst.Gallery = make([]domain.GalleryItem, len(galleryWires))
	for i, w := range galleryWires {
		inst.Gallery[i] = w.toDomain()
	}

	var alumniWires []alumnusWire
	if err := json.Unmarshal(alumniRaw, &alumniWires); err != nil {
		return nil, fmt.Errorf("alumni: %w", err)
	}
	inst.Alumni = make([]domain.Alumnus, len(alumniWires))
	for i, w := range alumniWires {
		inst.Alumni[i] = w.toDomain()
	}

	var transportWires []transportRouteWire
	if err := json.Unmarshal(transportRoutesRaw, &transportWires); err != nil {
		return nil, fmt.Errorf("transport_routes: %w", err)
	}
	inst.TransportRoutes = make([]domain.TransportRoute, len(transportWires))
	for i, w := range transportWires {
		inst.TransportRoutes[i] = w.toDomain()
	}

	var mealWires []mealPlanWire
	if err := json.Unmarshal(mealsRaw, &mealWires); err != nil {
		return nil, fmt.Errorf("meal_plans: %w", err)
	}
	inst.MealPlans = make([]domain.MealPlan, len(mealWires))
	for i, w := range mealWires {
		inst.MealPlans[i] = w.toDomain()
	}

	return inst, nil
}

// bilingualJSON — форма JSONB-колонок {ru,tg} на границе БД.
type bilingualJSON struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

// scanBilingual разбирает обязательную (NOT NULL) JSONB-колонку {ru,tg}.
func scanBilingual(raw []byte) (domain.Bilingual, error) {
	if len(raw) == 0 {
		return domain.Bilingual{}, nil
	}
	var v bilingualJSON
	if err := json.Unmarshal(raw, &v); err != nil {
		return domain.Bilingual{}, fmt.Errorf("unmarshal bilingual: %w", err)
	}
	return domain.Bilingual{RU: v.RU, TG: v.TG}, nil
}

// scanBilingualPtr разбирает nullable JSONB-колонку {ru,tg}, возвращая nil для NULL.
func scanBilingualPtr(raw []byte) (*domain.Bilingual, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	b, err := scanBilingual(raw)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// socialsJSON — форма JSONB-колонки socials на границе БД.
type socialsJSON struct {
	Instagram *string `json:"instagram,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	Facebook  *string `json:"facebook,omitempty"`
}

// scanSocials разбирает nullable JSONB-колонку socials, возвращая nil для NULL.
func scanSocials(raw []byte) (*domain.Socials, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var v socialsJSON
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("unmarshal socials: %w", err)
	}
	return &domain.Socials{Instagram: v.Instagram, Telegram: v.Telegram, Facebook: v.Facebook}, nil
}

// nonNilStrings нормализует nil-слайс в пустой — nullable TEXT[]-колонки (languages,
// program_level, curriculum, discount_type) без DEFAULT сканируются pgx как nil при SQL NULL,
// что нарушило бы non-nil гарантию domain.NewInstitution (JSON-контракт: "curriculum":[], не null).
func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// scanInstitution сканирует одну строку catalog.institutions (в порядке listColumns,
// плюс вычисляемая колонка distance последней) в domain.Institution.
func scanInstitution(rows pgx.Rows) (*domain.Institution, error) {
	var (
		id                                         uuid.UUID
		nameRaw, cityRaw, descRaw, addrRaw         []byte
		socialsRaw, tagRaw                         []byte
		region                                     string
		district                                   *string
		lat, lng                                   float64
		locationLandmarks, phone, email, website   *string
		coverPhotoS3Key, ageRange, licenseNo       *string
		types, languages, programLevel, curriculum []string
		price                                      *int
		discountAvailable                          bool
		discountType                               []string
		discountDetails                            *string
		verified                                   bool
		moderationStatus, plan                     string
		planExpiresAt                              *time.Time
		founded, studentsCount                     *int
		ratingAvg                                  *float64
		reviewCount                                int
		createdAt, updatedAt                       time.Time
		distanceM                                  *float64
	)

	if err := rows.Scan(
		&id, &nameRaw, &types, &region, &cityRaw, &district, &descRaw, &addrRaw,
		&lat, &lng, &locationLandmarks,
		&phone, &email, &website, &socialsRaw, &coverPhotoS3Key, &ageRange, &tagRaw,
		&licenseNo, &languages, &programLevel, &curriculum, &price,
		&discountAvailable, &discountType, &discountDetails, &verified,
		&moderationStatus, &plan, &planExpiresAt, &founded, &studentsCount,
		&ratingAvg, &reviewCount, &createdAt, &updatedAt,
		&distanceM,
	); err != nil {
		return nil, err
	}

	name, err := scanBilingual(nameRaw)
	if err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}

	inst := domain.NewInstitution(id, name, region)
	inst.Types = types

	if inst.City, err = scanBilingualPtr(cityRaw); err != nil {
		return nil, fmt.Errorf("city: %w", err)
	}
	inst.District = district
	if inst.Description, err = scanBilingualPtr(descRaw); err != nil {
		return nil, fmt.Errorf("description: %w", err)
	}
	if inst.Address, err = scanBilingualPtr(addrRaw); err != nil {
		return nil, fmt.Errorf("address: %w", err)
	}
	inst.Lat = lat
	inst.Lng = lng
	inst.LocationLandmarks = locationLandmarks
	inst.Phone = phone
	inst.Email = email
	inst.Website = website
	if inst.Socials, err = scanSocials(socialsRaw); err != nil {
		return nil, fmt.Errorf("socials: %w", err)
	}
	inst.CoverPhotoS3Key = coverPhotoS3Key
	inst.AgeRange = ageRange
	if inst.Tag, err = scanBilingualPtr(tagRaw); err != nil {
		return nil, fmt.Errorf("tag: %w", err)
	}
	inst.LicenseNo = licenseNo
	inst.Languages = nonNilStrings(languages)
	inst.ProgramLevel = nonNilStrings(programLevel)
	inst.Curriculum = nonNilStrings(curriculum)
	inst.Price = price
	inst.DiscountAvailable = discountAvailable
	inst.DiscountType = nonNilStrings(discountType)
	inst.DiscountDetails = discountDetails
	inst.Verified = verified
	inst.ModerationStatus = moderationStatus
	inst.Plan = plan
	inst.PlanExpiresAt = planExpiresAt
	inst.Founded = founded
	inst.StudentsCount = studentsCount
	inst.RatingAvg = ratingAvg
	inst.ReviewCount = reviewCount
	inst.CreatedAt = createdAt
	inst.UpdatedAt = updatedAt
	inst.DistanceM = distanceM

	return inst, nil
}

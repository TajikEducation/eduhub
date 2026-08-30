package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// biWire — форма JSONB {ru,tg} на границе БД (см. .claude/rules/db.md, двуязычные поля).
type biWire struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func biJSON(b bi) ([]byte, error) {
	return json.Marshal(biWire{RU: b.RU, TG: b.TG})
}

func biJSONPtr(b *bi) ([]byte, error) {
	if b == nil {
		return nil, nil
	}
	return biJSON(*b)
}

func biSliceJSON(bs []bi) ([]byte, error) {
	wire := make([]biWire, len(bs))
	for i, b := range bs {
		wire[i] = biWire{RU: b.RU, TG: b.TG}
	}
	return json.Marshal(wire)
}

type linkWire struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// linksJSON — сериализует ссылки достижения; пустой список даёт SQL NULL (links nullable).
func linksJSON(links []link) ([]byte, error) {
	if len(links) == 0 {
		return nil, nil
	}
	wire := make([]linkWire, len(links))
	for i, l := range links {
		wire[i] = linkWire{Label: l.Label, URL: l.URL}
	}
	return json.Marshal(wire)
}

type socialsWire struct {
	Instagram *string `json:"instagram,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	Facebook  *string `json:"facebook,omitempty"`
}

func socialsJSON(s *socialsSeed) ([]byte, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(socialsWire{Instagram: s.Instagram, Telegram: s.Telegram, Facebook: s.Facebook})
}

// ensureEntity — идемпотентное создание сущности через platform.seed_refs: если seed_ref уже
// записан (для entityType), переиспользует существующий entity_id и НЕ вызывает insertFn;
// иначе вызывает insertFn, вставляет запись и записывает связку в seed_refs — всё в одной tx.
func ensureEntity(ctx context.Context, tx pgx.Tx, seedRef, entityType string, insertFn func(ctx context.Context) (uuid.UUID, error)) (uuid.UUID, error) {
	var existing uuid.UUID
	err := tx.QueryRow(ctx,
		`SELECT entity_id FROM platform.seed_refs WHERE seed_ref = $1 AND entity_type = $2`,
		seedRef, entityType,
	).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("devseed: lookup seed_ref %s/%s: %w", entityType, seedRef, err)
	}

	id, err := insertFn(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO platform.seed_refs (seed_ref, entity_type, entity_id) VALUES ($1, $2, $3)`,
		seedRef, entityType, id,
	); err != nil {
		return uuid.Nil, fmt.Errorf("devseed: record seed_ref %s/%s: %w", entityType, seedRef, err)
	}

	return id, nil
}

// insertInstitution вставляет одну строку catalog.institutions и возвращает её id.
func insertInstitution(ctx context.Context, tx pgx.Tx, s institutionSeed) (uuid.UUID, error) {
	nameJSON, err := biJSON(s.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution name: %w", err)
	}
	cityJSON, err := biJSON(s.City)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution city: %w", err)
	}
	descJSON, err := biJSON(s.Description)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution description: %w", err)
	}
	addrJSON, err := biJSON(s.Address)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution address: %w", err)
	}
	tagJSON, err := biJSONPtr(s.Tag)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution tag: %w", err)
	}
	socialsJSONBytes, err := socialsJSON(s.Socials)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal institution socials: %w", err)
	}

	if s.Status != "approved" {
		return uuid.Nil, fmt.Errorf("devseed: institution %d has unexpected moderation_status %q, want approved", s.ID, s.Status)
	}

	const q = `
		INSERT INTO catalog.institutions (
			name, types, region, city, district, description, address, geo,
			phone, email, website, socials, cover_photo_s3_key, age_range, tag,
			curriculum, price, verified, moderation_status, founded, students_count,
			rating_avg, review_count
		) VALUES (
			$1::jsonb, $2, $3, $4::jsonb, $5, $6::jsonb, $7::jsonb,
			ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography,
			$10, $11, $12, $13::jsonb, $14, $15, $16::jsonb,
			$17, $18, $19, $20, $21, $22,
			$23, $24
		)
		RETURNING id
	`

	var id uuid.UUID
	err = tx.QueryRow(ctx, q,
		nameJSON, []string{s.Type}, s.Region, cityJSON, s.Area, descJSON, addrJSON,
		s.Lng, s.Lat,
		s.Phone, s.Email, s.Website, socialsJSONBytes, s.CoverPhoto, s.Age, tagJSON,
		s.Curriculum, s.Price, s.Ver, s.Status, s.Founded, s.Students,
		s.Score, s.Rev,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert institution %d: %w", s.ID, err)
	}

	return id, nil
}

// insertStaff вставляет одну строку catalog.institution_staff и возвращает её id.
func insertStaff(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID, s staffSeed) (uuid.UUID, error) {
	nameJSON, err := biJSON(s.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal staff %s name: %w", s.ID, err)
	}
	roleLabelJSON, err := biJSON(s.RoleLabel)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal staff %s role_label: %w", s.ID, err)
	}
	subjectJSON, err := biJSONPtr(s.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal staff %s subject: %w", s.ID, err)
	}
	bioJSON, err := biJSON(s.Bio)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal staff %s bio: %w", s.ID, err)
	}
	educationJSON, err := biSliceJSON(s.Education)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal staff %s education: %w", s.ID, err)
	}

	const q = `
		INSERT INTO catalog.institution_staff (
			institution_id, name, role_type, role_label, subject, photo_url, exp, bio, education, email, phone
		) VALUES (
			$1, $2::jsonb, $3, $4::jsonb, $5::jsonb, $6, $7, $8::jsonb, $9::jsonb, $10, $11
		)
		RETURNING id
	`

	var id uuid.UUID
	err = tx.QueryRow(ctx, q,
		institutionID, nameJSON, s.RoleType, roleLabelJSON, subjectJSON, s.Photo, s.Exp, bioJSON, educationJSON, s.Email, s.Phone,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert staff %s: %w", s.ID, err)
	}

	return id, nil
}

// insertAchievement вставляет одну строку catalog.achievements (полиморфная связь
// owner_type/owner_id) и возвращает её id.
func insertAchievement(ctx context.Context, tx pgx.Tx, ownerType string, ownerID uuid.UUID, a achievementSeed) (uuid.UUID, error) {
	titleJSON, err := biJSON(a.Title)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal achievement %s title: %w", a.ID, err)
	}
	descJSON, err := biJSON(a.Desc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal achievement %s desc: %w", a.ID, err)
	}
	linksJSONBytes, err := linksJSON(a.Links)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal achievement %s links: %w", a.ID, err)
	}

	const q = `
		INSERT INTO catalog.achievements (owner_type, owner_id, title, year, category, description, links)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6::jsonb, $7::jsonb)
		RETURNING id
	`

	var id uuid.UUID
	err = tx.QueryRow(ctx, q, ownerType, ownerID, titleJSON, a.Year, a.Category, descJSON, linksJSONBytes).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert achievement %s: %w", a.ID, err)
	}

	return id, nil
}

// insertGalleryItem вставляет одну строку catalog.institution_gallery.
func insertGalleryItem(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID, g galleryItemSeed, sortOrder int) (uuid.UUID, error) {
	labelJSON, err := biJSONPtr(g.Label)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal gallery item label: %w", err)
	}

	const q = `
		INSERT INTO catalog.institution_gallery (institution_id, s3_key, label, sort_order)
		VALUES ($1, $2, $3::jsonb, $4)
		RETURNING id
	`

	var id uuid.UUID
	if err := tx.QueryRow(ctx, q, institutionID, g.URL, labelJSON, sortOrder).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert gallery item: %w", err)
	}

	return id, nil
}

// insertAlumnus вставляет одну строку catalog.institution_alumni.
func insertAlumnus(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID, a alumnusSeed) (uuid.UUID, error) {
	nameJSON, err := biJSON(a.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal alumnus %s name: %w", a.ID, err)
	}
	nowJSON, err := biJSON(a.Now)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal alumnus %s now: %w", a.ID, err)
	}

	const q = `
		INSERT INTO catalog.institution_alumni (institution_id, name, photo_url, grad_year, now_label)
		VALUES ($1, $2::jsonb, $3, $4, $5::jsonb)
		RETURNING id
	`

	var id uuid.UUID
	if err := tx.QueryRow(ctx, q, institutionID, nameJSON, a.Photo, a.GradYear, nowJSON).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert alumnus %s: %w", a.ID, err)
	}

	return id, nil
}

// insertTransportRoute вставляет одну строку catalog.institution_transport_routes.
// type всегда 'other' — свободный текст мока (transportInfo.types[i]) нельзя надёжно
// классифицировать в enum own_bus|minibus|taxi|parent_coop|other.
func insertTransportRoute(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID, label bi, areas []bi, cost, sortOrder int) (uuid.UUID, error) {
	labelJSON, err := biJSON(label)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal transport route label: %w", err)
	}
	areasJSON, err := biSliceJSON(areas)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal transport route areas: %w", err)
	}

	const q = `
		INSERT INTO catalog.institution_transport_routes (institution_id, type, label, areas, cost, sort_order)
		VALUES ($1, 'other', $2::jsonb, $3::jsonb, $4, $5)
		RETURNING id
	`

	var id uuid.UUID
	if err := tx.QueryRow(ctx, q, institutionID, labelJSON, areasJSON, cost, sortOrder).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert transport route: %w", err)
	}

	return id, nil
}

// insertMealPlan вставляет единственную строку catalog.institution_meal_plans институции
// (meal_type='other' — mock даёт только булев признак наличия питания, без деталей).
func insertMealPlan(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID) (uuid.UUID, error) {
	const q = `
		INSERT INTO catalog.institution_meal_plans (institution_id, meal_type)
		VALUES ($1, 'other')
		RETURNING id
	`

	var id uuid.UUID
	if err := tx.QueryRow(ctx, q, institutionID).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert meal plan: %w", err)
	}

	return id, nil
}

// insertNews вставляет одну строку catalog.news_articles.
func insertNews(ctx context.Context, tx pgx.Tx, institutionID uuid.UUID, n newsSeed) (uuid.UUID, error) {
	titleJSON, err := biJSON(n.Title)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal news %s title: %w", n.ID, err)
	}
	categoryJSON, err := biJSONPtr(&n.Category)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal news %s category: %w", n.ID, err)
	}
	contentJSON, err := biJSON(n.Content)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: marshal news %s content: %w", n.ID, err)
	}
	var tagsJSON []byte
	if n.Tags != nil {
		tagsJSON, err = biSliceJSON(n.Tags)
		if err != nil {
			return uuid.Nil, fmt.Errorf("devseed: marshal news %s tags: %w", n.ID, err)
		}
	}

	const q = `
		INSERT INTO catalog.news_articles (
			institution_id, title, category, cover_s3_key, video_url, content, tags, status, views_count
		) VALUES (
			$1, $2::jsonb, $3::jsonb, $4, $5, $6::jsonb, $7::jsonb, $8, $9
		)
		RETURNING id
	`

	var id uuid.UUID
	err = tx.QueryRow(ctx, q,
		institutionID, titleJSON, categoryJSON, n.CoverURL, n.VideoURL, contentJSON, tagsJSON, n.Status, n.Views,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("devseed: insert news %s: %w", n.ID, err)
	}

	return id, nil
}

// seedAchievements идемпотентно вставляет достижения owner'а (институции или сотрудника)
// в общую полиморфную таблицу catalog.achievements.
func seedAchievements(ctx context.Context, tx pgx.Tx, ownerType string, ownerID uuid.UUID, achievements []achievementSeed) error {
	for _, a := range achievements {
		a := a
		if _, err := ensureEntity(ctx, tx, a.ID, "achievement", func(ctx context.Context) (uuid.UUID, error) {
			return insertAchievement(ctx, tx, ownerType, ownerID, a)
		}); err != nil {
			return err
		}
	}
	return nil
}

// SeedInstitution идемпотентно сидирует одну институцию и все её сателлиты (персонал,
// достижения, галерею, выпускников, маршруты транспорта, план питания, новости) в
// одной переданной транзакции. Порядок вставки обязателен из-за FK: сначала институция
// (нужен её UUID), затем сателлиты, затем достижения персонала (нужен UUID сотрудника).
func SeedInstitution(ctx context.Context, tx pgx.Tx, s institutionSeed) error {
	institutionSeedRef := fmt.Sprintf("web-inst-%d", s.ID)

	institutionID, err := ensureEntity(ctx, tx, institutionSeedRef, "institution", func(ctx context.Context) (uuid.UUID, error) {
		return insertInstitution(ctx, tx, s)
	})
	if err != nil {
		return fmt.Errorf("devseed: seed institution %d: %w", s.ID, err)
	}

	if err := seedAchievements(ctx, tx, "institution", institutionID, s.Achievements); err != nil {
		return fmt.Errorf("devseed: seed achievements for institution %d: %w", s.ID, err)
	}

	for i, g := range s.Gallery {
		g, i := g, i
		seedRef := fmt.Sprintf("web-inst-%d-gallery-%d", s.ID, i)
		if _, err := ensureEntity(ctx, tx, seedRef, "gallery_item", func(ctx context.Context) (uuid.UUID, error) {
			return insertGalleryItem(ctx, tx, institutionID, g, i)
		}); err != nil {
			return fmt.Errorf("devseed: seed gallery item %d for institution %d: %w", i, s.ID, err)
		}
	}

	for _, a := range s.Alumni {
		a := a
		if _, err := ensureEntity(ctx, tx, a.ID, "alumnus", func(ctx context.Context) (uuid.UUID, error) {
			return insertAlumnus(ctx, tx, institutionID, a)
		}); err != nil {
			return fmt.Errorf("devseed: seed alumnus %s for institution %d: %w", a.ID, s.ID, err)
		}
	}

	// Маршруты транспорта — только если Transport==true И TransportInfo присутствует;
	// по одной строке на каждый transportInfo.types[i], area/cost — общие для всех строк.
	if s.Transport && s.TransportInfo != nil {
		for i, t := range s.TransportInfo.Types {
			t, i := t, i
			seedRef := fmt.Sprintf("web-inst-%d-transport-%d", s.ID, i)
			areas := s.TransportInfo.Areas
			cost := s.TransportInfo.Cost
			if _, err := ensureEntity(ctx, tx, seedRef, "transport_route", func(ctx context.Context) (uuid.UUID, error) {
				return insertTransportRoute(ctx, tx, institutionID, t, areas, cost, i)
			}); err != nil {
				return fmt.Errorf("devseed: seed transport route %d for institution %d: %w", i, s.ID, err)
			}
		}
	}

	// План питания — максимум один на институцию, только если Food==true.
	if s.Food {
		seedRef := fmt.Sprintf("web-inst-%d-meal", s.ID)
		if _, err := ensureEntity(ctx, tx, seedRef, "meal_plan", func(ctx context.Context) (uuid.UUID, error) {
			return insertMealPlan(ctx, tx, institutionID)
		}); err != nil {
			return fmt.Errorf("devseed: seed meal plan for institution %d: %w", s.ID, err)
		}
	}

	for _, staff := range staffForInst(s.ID) {
		staff := staff
		staffID, err := ensureEntity(ctx, tx, staff.ID, "staff", func(ctx context.Context) (uuid.UUID, error) {
			return insertStaff(ctx, tx, institutionID, staff)
		})
		if err != nil {
			return fmt.Errorf("devseed: seed staff %s for institution %d: %w", staff.ID, s.ID, err)
		}

		if err := seedAchievements(ctx, tx, "staff", staffID, staff.Achievements); err != nil {
			return fmt.Errorf("devseed: seed achievements for staff %s: %w", staff.ID, err)
		}
	}

	for _, n := range newsForInst(s.ID) {
		n := n
		if _, err := ensureEntity(ctx, tx, n.ID, "news_article", func(ctx context.Context) (uuid.UUID, error) {
			return insertNews(ctx, tx, institutionID, n)
		}); err != nil {
			return fmt.Errorf("devseed: seed news %s for institution %d: %w", n.ID, s.ID, err)
		}
	}

	return nil
}

// SeedAll сидирует все 9 демо-институций (docs/EduHub_Backend_Development_Plan.md,
// задача 30) и их сателлиты. Каждая институция — отдельная транзакция (BEGIN/COMMIT),
// чтобы сбой на одной институции не откатывал уже успешно засеянные.
func SeedAll(ctx context.Context, db interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}) error {
	for _, inst := range institutions {
		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("devseed: begin tx for institution %d: %w", inst.ID, err)
		}

		if err := SeedInstitution(ctx, tx, inst); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("devseed: commit tx for institution %d: %w", inst.ID, err)
		}
	}

	return nil
}

// Package postgres — pgx-реализация internal/applicants/usecase.{ApplicantRepo,AchievementRepo,
// ApplicationRepo}.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/applicants/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// querier — минимальный интерфейс доступа к БД, тот же паттерн, что у остальных модулей.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

const foreignKeyViolationCode = "23503"

type bilingualJSON struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func bilingualJSONOf(b domain.Bilingual) []byte {
	raw, _ := json.Marshal(bilingualJSON{RU: b.RU, TG: b.TG})
	return raw
}

func bilingualJSONPtrOf(b *domain.Bilingual) []byte {
	if b == nil {
		return nil
	}
	return bilingualJSONOf(*b)
}

func bilingualSliceJSONOf(items []domain.Bilingual) []byte {
	if items == nil {
		return nil
	}
	wire := make([]bilingualJSON, len(items))
	for i, b := range items {
		wire[i] = bilingualJSON{RU: b.RU, TG: b.TG}
	}
	raw, _ := json.Marshal(wire)
	return raw
}

func scanBilingual(raw []byte) (domain.Bilingual, error) {
	var wire bilingualJSON
	if err := json.Unmarshal(raw, &wire); err != nil {
		return domain.Bilingual{}, err
	}
	return domain.Bilingual{RU: wire.RU, TG: wire.TG}, nil
}

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

func scanBilingualSlice(raw []byte) ([]domain.Bilingual, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var wire []bilingualJSON
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	out := make([]domain.Bilingual, len(wire))
	for i, w := range wire {
		out[i] = domain.Bilingual{RU: w.RU, TG: w.TG}
	}
	return out, nil
}

// ApplicantRepo — pgx-реализация usecase.ApplicantRepo.
type ApplicantRepo struct {
	db querier
}

func NewApplicantRepo(db querier) *ApplicantRepo {
	return &ApplicantRepo{db: db}
}

const applicantColumns = `id, user_id, name, photo_url, position, bio, education, experience, skills, email, phone, cv_s3_key, visibility, hide_contacts, created_at, updated_at`

func scanApplicant(row interface{ Scan(dest ...any) error }) (domain.Applicant, error) {
	var a domain.Applicant
	var nameRaw, positionRaw, bioRaw, educationRaw, experienceRaw, skillsRaw []byte
	if err := row.Scan(&a.ID, &a.UserID, &nameRaw, &a.PhotoURL, &positionRaw, &bioRaw, &educationRaw, &experienceRaw, &skillsRaw, &a.Email, &a.Phone, &a.CVS3Key, &a.Visibility, &a.HideContacts, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return domain.Applicant{}, err
	}
	name, err := scanBilingual(nameRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant name: %w", err)
	}
	a.Name = name
	position, err := scanBilingual(positionRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant position: %w", err)
	}
	a.Position = position
	bio, err := scanBilingualPtr(bioRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant bio: %w", err)
	}
	a.Bio = bio
	education, err := scanBilingualSlice(educationRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant education: %w", err)
	}
	a.Education = education
	experience, err := scanBilingualSlice(experienceRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant experience: %w", err)
	}
	a.Experience = experience
	skills, err := scanBilingualSlice(skillsRaw)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("unmarshal applicant skills: %w", err)
	}
	a.Skills = skills
	return a, nil
}

// Upsert создаёт профиль соискателя или полностью заменяет существующий (UNIQUE(user_id)).
func (r *ApplicantRepo) Upsert(ctx context.Context, userID uuid.UUID, in domain.ApplicantInput) (domain.Applicant, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.applicants (user_id, name, photo_url, position, bio, education, experience, skills, email, phone, cv_s3_key, visibility, hide_contacts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id) DO UPDATE SET
			name = EXCLUDED.name, photo_url = EXCLUDED.photo_url, position = EXCLUDED.position, bio = EXCLUDED.bio,
			education = EXCLUDED.education, experience = EXCLUDED.experience, skills = EXCLUDED.skills,
			email = EXCLUDED.email, phone = EXCLUDED.phone, cv_s3_key = EXCLUDED.cv_s3_key,
			visibility = EXCLUDED.visibility, hide_contacts = EXCLUDED.hide_contacts, updated_at = now()
		RETURNING `+applicantColumns,
		userID, bilingualJSONOf(in.Name), in.PhotoURL, bilingualJSONOf(in.Position), bilingualJSONPtrOf(in.Bio),
		bilingualSliceJSONOf(in.Education), bilingualSliceJSONOf(in.Experience), bilingualSliceJSONOf(in.Skills),
		in.Email, in.Phone, in.CVS3Key, in.Visibility, in.HideContacts)

	a, err := scanApplicant(row)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("postgres: upsert applicant: %w", err)
	}
	return a, nil
}

// GetByUserID возвращает профиль соискателя по user_id.
func (r *ApplicantRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Applicant, error) {
	row := r.db.QueryRow(ctx, `SELECT `+applicantColumns+` FROM communications.applicants WHERE user_id = $1`, userID)
	a, err := scanApplicant(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Applicant{}, apperr.NotFound("applicant", userID.String())
		}
		return domain.Applicant{}, fmt.Errorf("postgres: get applicant by user id: %w", err)
	}
	return a, nil
}

// GetByID возвращает профиль соискателя по id.
func (r *ApplicantRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Applicant, error) {
	row := r.db.QueryRow(ctx, `SELECT `+applicantColumns+` FROM communications.applicants WHERE id = $1`, id)
	a, err := scanApplicant(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Applicant{}, apperr.NotFound("applicant", id.String())
		}
		return domain.Applicant{}, fmt.Errorf("postgres: get applicant: %w", err)
	}
	return a, nil
}

// ListPublic возвращает полностью публичные профили, самые новые первыми.
func (r *ApplicantRepo) ListPublic(ctx context.Context) ([]domain.Applicant, error) {
	rows, err := r.db.Query(ctx, `SELECT `+applicantColumns+` FROM communications.applicants WHERE visibility = 'public' ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list public applicants: %w", err)
	}
	defer rows.Close()

	var out []domain.Applicant
	for rows.Next() {
		a, err := scanApplicant(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan applicant: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list public applicants rows: %w", err)
	}
	return out, nil
}

// AchievementRepo — pgx-реализация usecase.AchievementRepo.
type AchievementRepo struct {
	db querier
}

func NewAchievementRepo(db querier) *AchievementRepo {
	return &AchievementRepo{db: db}
}

// Create вставляет достижение соискателя.
func (r *AchievementRepo) Create(ctx context.Context, applicantID uuid.UUID, in domain.AchievementInput) (domain.Achievement, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.applicant_achievements (applicant_id, title, year, category, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, applicantID, in.Title, in.Year, in.Category, in.Description)

	a := domain.Achievement{ApplicantID: applicantID, Title: in.Title, Year: in.Year, Category: in.Category, Description: in.Description}
	if err := row.Scan(&a.ID, &a.CreatedAt); err != nil {
		return domain.Achievement{}, fmt.Errorf("postgres: insert applicant achievement: %w", err)
	}
	return a, nil
}

// Delete удаляет достижение соискателя.
func (r *AchievementRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM communications.applicant_achievements WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete applicant achievement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("applicant_achievement", id.String())
	}
	return nil
}

// ListByApplicant возвращает достижения соискателя.
func (r *AchievementRepo) ListByApplicant(ctx context.Context, applicantID uuid.UUID) ([]domain.Achievement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, applicant_id, title, year, category, description, created_at
		FROM communications.applicant_achievements
		WHERE applicant_id = $1
		ORDER BY created_at DESC
	`, applicantID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list applicant achievements: %w", err)
	}
	defer rows.Close()

	var out []domain.Achievement
	for rows.Next() {
		var a domain.Achievement
		if err := rows.Scan(&a.ID, &a.ApplicantID, &a.Title, &a.Year, &a.Category, &a.Description, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan applicant achievement: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list applicant achievements rows: %w", err)
	}
	return out, nil
}

// GetApplicantID возвращает applicant_id владеющей записи достижения.
func (r *AchievementRepo) GetApplicantID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT applicant_id FROM communications.applicant_achievements WHERE id = $1`, id)
	var applicantID uuid.UUID
	if err := row.Scan(&applicantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("applicant_achievement", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get applicant achievement applicant id: %w", err)
	}
	return applicantID, nil
}

// ApplicationRepo — pgx-реализация usecase.ApplicationRepo.
type ApplicationRepo struct {
	db querier
}

func NewApplicationRepo(db querier) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

// Create вставляет отклик или возвращает уже существующий (UNIQUE(applicant_id, vacancy_id) —
// тот же upsert-return паттерн, что internal/communications/repo/postgres.ConversationRepo.GetOrCreate).
// FK-нарушение (несуществующая vacancy_id) маппится в apperr.NotFound.
func (r *ApplicationRepo) Create(ctx context.Context, applicantID, vacancyID uuid.UUID) (domain.Application, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.applications (applicant_id, vacancy_id)
		VALUES ($1, $2)
		ON CONFLICT (applicant_id, vacancy_id) DO UPDATE SET applicant_id = EXCLUDED.applicant_id
		RETURNING id, applicant_id, vacancy_id, status, created_at
	`, applicantID, vacancyID)

	var a domain.Application
	if err := row.Scan(&a.ID, &a.ApplicantID, &a.VacancyID, &a.Status, &a.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationCode {
			return domain.Application{}, apperr.NotFound("vacancy", vacancyID.String())
		}
		return domain.Application{}, fmt.Errorf("postgres: create application: %w", err)
	}
	return a, nil
}

// ListVacancyIDsByApplicant возвращает id вакансий, на которые откликнулся соискатель.
func (r *ApplicationRepo) ListVacancyIDsByApplicant(ctx context.Context, applicantID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT vacancy_id FROM communications.applications WHERE applicant_id = $1`, applicantID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list applications: %w", err)
	}
	defer rows.Close()

	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("postgres: scan application vacancy id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list applications rows: %w", err)
	}
	return out, nil
}

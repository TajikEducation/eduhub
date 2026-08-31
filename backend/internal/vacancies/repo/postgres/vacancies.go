// Package postgres — pgx-реализация internal/vacancies/usecase.VacancyRepo.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
)

// querier — минимальный интерфейс доступа к БД (пул или транзакция), тот же паттерн, что
// internal/catalog и internal/reviews.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// VacancyRepo — pgx-реализация usecase.VacancyRepo.
type VacancyRepo struct {
	db querier
}

// New создаёт VacancyRepo.
func New(db querier) *VacancyRepo {
	return &VacancyRepo{db: db}
}

type bilingualJSON struct {
	RU string `json:"ru"`
	TG string `json:"tg"`
}

func bilingualJSONOf(b domain.Bilingual) []byte {
	raw, _ := json.Marshal(bilingualJSON{RU: b.RU, TG: b.TG})
	return raw
}

func bilingualSliceJSONOf(items []domain.Bilingual) []byte {
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

const vacancyColumns = `id, institution_id, title, description, requirements, salary_from, salary_to, employment, status, created_at, updated_at`

func scanVacancy(row pgx.Row) (domain.Vacancy, error) {
	var v domain.Vacancy
	var titleRaw, descRaw, reqRaw, employmentRaw []byte
	if err := row.Scan(&v.ID, &v.InstitutionID, &titleRaw, &descRaw, &reqRaw, &v.SalaryFrom, &v.SalaryTo, &employmentRaw, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return domain.Vacancy{}, err
	}
	title, err := scanBilingual(titleRaw)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("unmarshal vacancy title: %w", err)
	}
	v.Title = title
	desc, err := scanBilingual(descRaw)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("unmarshal vacancy description: %w", err)
	}
	v.Description = desc
	reqs, err := scanBilingualSlice(reqRaw)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("unmarshal vacancy requirements: %w", err)
	}
	v.Requirements = reqs
	employment, err := scanBilingual(employmentRaw)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("unmarshal vacancy employment: %w", err)
	}
	v.Employment = employment
	return v, nil
}

// Create вставляет новую вакансию.
func (r *VacancyRepo) Create(ctx context.Context, institutionID uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO communications.vacancies (institution_id, title, description, requirements, salary_from, salary_to, employment, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING `+vacancyColumns, institutionID, bilingualJSONOf(in.Title), bilingualJSONOf(in.Description), bilingualSliceJSONOf(in.Requirements), in.SalaryFrom, in.SalaryTo, bilingualJSONOf(in.Employment), in.Status)

	v, err := scanVacancy(row)
	if err != nil {
		return domain.Vacancy{}, fmt.Errorf("postgres: insert vacancy: %w", err)
	}
	return v, nil
}

// Update заменяет данные вакансии (полная замена, как catalog.UpdateNews).
func (r *VacancyRepo) Update(ctx context.Context, id uuid.UUID, in domain.VacancyInput) (domain.Vacancy, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE communications.vacancies SET
			title = $2, description = $3, requirements = $4, salary_from = $5, salary_to = $6, employment = $7, status = $8, updated_at = now()
		WHERE id = $1
		RETURNING `+vacancyColumns, id, bilingualJSONOf(in.Title), bilingualJSONOf(in.Description), bilingualSliceJSONOf(in.Requirements), in.SalaryFrom, in.SalaryTo, bilingualJSONOf(in.Employment), in.Status)

	v, err := scanVacancy(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Vacancy{}, apperr.NotFound("vacancy", id.String())
		}
		return domain.Vacancy{}, fmt.Errorf("postgres: update vacancy: %w", err)
	}
	return v, nil
}

// Delete удаляет вакансию.
func (r *VacancyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM communications.vacancies WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete vacancy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("vacancy", id.String())
	}
	return nil
}

// GetByID возвращает вакансию по id, любого статуса.
func (r *VacancyRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Vacancy, error) {
	row := r.db.QueryRow(ctx, `SELECT `+vacancyColumns+` FROM communications.vacancies WHERE id = $1`, id)
	v, err := scanVacancy(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Vacancy{}, apperr.NotFound("vacancy", id.String())
		}
		return domain.Vacancy{}, fmt.Errorf("postgres: get vacancy: %w", err)
	}
	return v, nil
}

// GetInstitutionID возвращает institution_id вакансии.
func (r *VacancyRepo) GetInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT institution_id FROM communications.vacancies WHERE id = $1`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("vacancy", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get vacancy institution id: %w", err)
	}
	return instID, nil
}

func (r *VacancyRepo) queryList(ctx context.Context, query string, args ...any) ([]domain.Vacancy, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: list vacancies: %w", err)
	}
	defer rows.Close()

	var out []domain.Vacancy
	for rows.Next() {
		v, err := scanVacancy(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan vacancy: %w", err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list vacancies rows: %w", err)
	}
	return out, nil
}

// ListByInstitution возвращает вакансии институции; onlyPublished фильтрует по status='published'
// (для публичной вкладки профиля), иначе — все статусы (для кабинета учреждения).
func (r *VacancyRepo) ListByInstitution(ctx context.Context, institutionID uuid.UUID, onlyPublished bool) ([]domain.Vacancy, error) {
	if onlyPublished {
		return r.queryList(ctx, `SELECT `+vacancyColumns+` FROM communications.vacancies WHERE institution_id = $1 AND status = 'published' ORDER BY created_at DESC`, institutionID)
	}
	return r.queryList(ctx, `SELECT `+vacancyColumns+` FROM communications.vacancies WHERE institution_id = $1 ORDER BY created_at DESC`, institutionID)
}

// ListPublished возвращает опубликованные вакансии всех институций, самые новые первыми
// (для /api/v1/vacancies — регион/тип фильтруются на фронте по встроенной сводке об
// учреждении, см. пакетный комментарий internal/vacancies/transport/http).
func (r *VacancyRepo) ListPublished(ctx context.Context, limit int) ([]domain.Vacancy, error) {
	return r.queryList(ctx, `SELECT `+vacancyColumns+` FROM communications.vacancies WHERE status = 'published' ORDER BY created_at DESC LIMIT $1`, limit)
}

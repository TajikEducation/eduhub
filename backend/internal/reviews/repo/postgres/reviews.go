// Package postgres — реализация internal/reviews/usecase.ReviewRepo поверх PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/reviews/domain"
)

const uniqueViolationCode = "23505"

// querier — минимальный интерфейс доступа к БД (пул или транзакция).
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// ReviewRepo — репозиторий отзывов поверх PostgreSQL.
type ReviewRepo struct {
	db querier
}

// New создаёт ReviewRepo поверх переданного querier.
func New(db querier) *ReviewRepo {
	return &ReviewRepo{db: db}
}

const reviewColumns = `id, institution_id, user_id, rating, text, reply, replied_at, status, created_at, updated_at`

func scanReview(row pgx.Row) (domain.Review, error) {
	var r domain.Review
	var status string
	if err := row.Scan(&r.ID, &r.InstitutionID, &r.UserID, &r.Rating, &r.Text, &r.Reply, &r.RepliedAt, &status, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return domain.Review{}, err
	}
	r.Status = domain.Status(status)
	return r, nil
}

// Create вставляет новый отзыв. Повтор (institution_id,user_id) → apperr.Conflict.
func (r *ReviewRepo) Create(ctx context.Context, rev domain.Review) (domain.Review, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO reviews.reviews (institution_id, user_id, rating, text, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+reviewColumns,
		rev.InstitutionID, rev.UserID, rev.Rating, rev.Text, string(rev.Status),
	)

	created, err := scanReview(row)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return domain.Review{}, apperr.Conflict("review_already_exists")
	}
	if err != nil {
		return domain.Review{}, fmt.Errorf("postgres: create review: %w", err)
	}
	return created, nil
}

// GetByID ищет отзыв по id.
func (r *ReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Review, error) {
	row := r.db.QueryRow(ctx, `SELECT `+reviewColumns+` FROM reviews.reviews WHERE id = $1`, id)
	rev, err := scanReview(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Review{}, apperr.NotFound("review", id.String())
	}
	if err != nil {
		return domain.Review{}, fmt.Errorf("postgres: get review by id: %w", err)
	}
	return rev, nil
}

// ListByInstitution возвращает отзывы институции, опционально только approved.
func (r *ReviewRepo) ListByInstitution(ctx context.Context, institutionID uuid.UUID, onlyApproved bool) ([]domain.Review, error) {
	query := `SELECT ` + reviewColumns + ` FROM reviews.reviews WHERE institution_id = $1`
	if onlyApproved {
		query += ` AND status = 'approved'`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, institutionID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list reviews: %w", err)
	}
	defer rows.Close()

	var out []domain.Review
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan review: %w", err)
		}
		out = append(out, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list reviews rows: %w", err)
	}
	return out, nil
}

// SetReply записывает ответ учреждения.
func (r *ReviewRepo) SetReply(ctx context.Context, id uuid.UUID, reply string) (domain.Review, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE reviews.reviews SET reply = $2, replied_at = now(), updated_at = now() WHERE id = $1
		RETURNING `+reviewColumns,
		id, reply,
	)
	updated, err := scanReview(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Review{}, apperr.NotFound("review", id.String())
	}
	if err != nil {
		return domain.Review{}, fmt.Errorf("postgres: set reply: %w", err)
	}
	return updated, nil
}

// SetStatus меняет статус модерации отзыва.
func (r *ReviewRepo) SetStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	tag, err := r.db.Exec(ctx, `UPDATE reviews.reviews SET status = $2, updated_at = now() WHERE id = $1`, id, string(status))
	if err != nil {
		return fmt.Errorf("postgres: set review status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("review", id.String())
	}
	return nil
}

// AggregateApproved возвращает среднюю оценку и количество approved-отзывов институции.
func (r *ReviewRepo) AggregateApproved(ctx context.Context, institutionID uuid.UUID) (float64, int, error) {
	row := r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews.reviews WHERE institution_id = $1 AND status = 'approved'
	`, institutionID)
	var avg float64
	var count int
	if err := row.Scan(&avg, &count); err != nil {
		return 0, 0, fmt.Errorf("postgres: aggregate approved reviews: %w", err)
	}
	return avg, count, nil
}

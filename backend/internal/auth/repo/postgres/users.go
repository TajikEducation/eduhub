// Package postgres — реализация internal/auth/usecase.UserRepo/RefreshTokenRepo поверх PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// uniqueViolationCode — код PostgreSQL для нарушения UNIQUE-ограничения.
const uniqueViolationCode = "23505"

// querier — минимальный интерфейс доступа к БД (пул или транзакция), см.
// internal/catalog/repo/postgres.querier — тот же паттерн ради тестируемости через rollback.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// UserRepo — репозиторий пользователей поверх PostgreSQL.
type UserRepo struct {
	db querier
}

// New создаёт UserRepo поверх переданного querier.
func New(db querier) *UserRepo {
	return &UserRepo{db: db}
}

const userColumns = `id, email, password_hash, role, status, display_name, created_at, updated_at`

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	var passwordHash *string
	var displayName *string
	if err := row.Scan(&u.ID, &u.Email, &passwordHash, &u.Role, &u.Status, &displayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return domain.User{}, err
	}
	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	if displayName != nil {
		u.DisplayName = *displayName
	}
	return u, nil
}

// Create вставляет нового пользователя. Email уже занят → apperr.Conflict("email_taken").
func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO auth.users (email, password_hash, role, status, display_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+userColumns,
		u.Email, nullIfEmpty(u.PasswordHash), u.Role, u.Status, nullIfEmpty(u.DisplayName),
	)

	created, err := scanUser(row)
	if isUniqueViolation(err) {
		return domain.User{}, apperr.Conflict("email_taken")
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("postgres: create user: %w", err)
	}
	return created, nil
}

// GetByEmail ищет пользователя регистронезависимо (см. миграцию 00006, lower(email) индекс).
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM auth.users WHERE lower(email) = lower($1)`, email)

	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, apperr.NotFound("user", email)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("postgres: get user by email: %w", err)
	}
	return user, nil
}

// GetByID ищет пользователя по id.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM auth.users WHERE id = $1`, id)

	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, apperr.NotFound("user", id.String())
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("postgres: get user by id: %w", err)
	}
	return user, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

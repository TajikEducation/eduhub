package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// uniqueViolationCode — код ошибки PostgreSQL 23505 (unique_violation).
const uniqueViolationCode = "23505"

// UserRepo — реализация usecase.UserRoleLookup поверх PostgreSQL. Названо UserRepo, не
// RoleRepo — единственный метод сейчас именно про роль, но домашняя схема для будущих auth.users
// операций (E2.4) будет расти в этом же репозитории, не отдельным типом на каждый метод.
type UserRepo struct {
	db querier
}

// NewUserRepo создаёт UserRepo поверх переданного querier (пул или транзакция).
func NewUserRepo(db querier) *UserRepo {
	return &UserRepo{db: db}
}

// RoleByUserID возвращает текущую роль пользователя. Не найден → apperr.NotFound.
func (r *UserRepo) RoleByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	const q = `SELECT role FROM auth.users WHERE id = $1`

	var role string
	if err := r.db.QueryRow(ctx, q, userID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.NotFound("user", userID.String())
		}
		return "", fmt.Errorf("postgres: role by user id: %w", err)
	}
	return role, nil
}

// Create вставляет нового пользователя и возвращает его с полями, сгенерированными БД
// (id, created_at, updated_at) — не только id: вызывающему (например DTO ответа /auth/register)
// нужна настоящая метка времени создания, не нулевое значение time.Time.
func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	const q = `
		INSERT INTO auth.users (email, display_name, locale, phone, password_hash, role, status,
			email_verified_at, consent_at, consent_version, failed_login_count, locked_until, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, q,
		u.Email, u.DisplayName, u.Locale, u.Phone, u.PasswordHash, u.Role, u.Status,
		u.EmailVerifiedAt, u.ConsentAt, u.ConsentVersion, u.FailedLoginCount, u.LockedUntil, u.DeletedAt,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domain.User{}, apperr.ConflictCode("email_taken", "email уже зарегистрирован")
		}
		return domain.User{}, fmt.Errorf("postgres: create user: %w", err)
	}
	return u, nil
}

// FindByEmail ищет пользователя по email как есть (нормализация — забота вызывающего usecase).
// Не найден → apperr.NotFound.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `
		SELECT id, email, display_name, locale, phone, password_hash, role, status,
			email_verified_at, consent_at, consent_version, failed_login_count, locked_until,
			deleted_at, created_at, updated_at
		FROM auth.users
		WHERE email = $1
	`
	return r.scanUser(r.db.QueryRow(ctx, q, email), "email", email)
}

// FindByID ищет пользователя по id. Не найден → apperr.NotFound.
func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const q = `
		SELECT id, email, display_name, locale, phone, password_hash, role, status,
			email_verified_at, consent_at, consent_version, failed_login_count, locked_until,
			deleted_at, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`
	return r.scanUser(r.db.QueryRow(ctx, q, id), "id", id.String())
}

// scanUser сканирует одну строку auth.users; resource/idValue — только для apperr.NotFound.
func (r *UserRepo) scanUser(row pgx.Row, idField, idValue string) (domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.Locale, &u.Phone, &u.PasswordHash, &u.Role, &u.Status,
		&u.EmailVerifiedAt, &u.ConsentAt, &u.ConsentVersion, &u.FailedLoginCount, &u.LockedUntil,
		&u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperr.NotFound("user", idField+"="+idValue)
		}
		return domain.User{}, fmt.Errorf("postgres: find user: %w", err)
	}
	return u, nil
}

// UpdateConsent обновляет consent_at/consent_version — переподтверждение при смене политики
// (POST /auth/consent), отдельно от начального consent при регистрации.
func (r *UserRepo) UpdateConsent(ctx context.Context, userID uuid.UUID, consentVersion string, consentAt time.Time) error {
	const q = `UPDATE auth.users SET consent_at = $2, consent_version = $3 WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, userID, consentAt, consentVersion); err != nil {
		return fmt.Errorf("postgres: update consent: %w", err)
	}
	return nil
}

// MarkEmailVerified ставит email_verified_at и переводит status в 'active' — успешная
// верификация email (POST /auth/verify) считается достаточной для активации аккаунта.
func (r *UserRepo) MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	const q = `UPDATE auth.users SET email_verified_at = $2, status = 'active' WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, userID, verifiedAt); err != nil {
		return fmt.Errorf("postgres: mark email verified: %w", err)
	}
	return nil
}

// UpdatePasswordHash — новый пароль после password-reset.
func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	const q = `UPDATE auth.users SET password_hash = $2 WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, userID, passwordHash); err != nil {
		return fmt.Errorf("postgres: update password hash: %w", err)
	}
	return nil
}

// SoftDelete — право на удаление аккаунта (закон РТ №1537): НЕ физический DELETE, статус
// 'deleted' + анонимизация PII-полей. anonymizedEmail — типично "deleted-<uuid>@eduhub.local",
// формирует вызывающий usecase, не репозиторий (это не забота техники хранения).
func (r *UserRepo) SoftDelete(ctx context.Context, userID uuid.UUID, anonymizedEmail string, deletedAt time.Time) error {
	const q = `UPDATE auth.users SET status = 'deleted', deleted_at = $2, email = $3, phone = NULL, password_hash = NULL WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, userID, deletedAt, anonymizedEmail); err != nil {
		return fmt.Errorf("postgres: soft delete user: %w", err)
	}
	return nil
}

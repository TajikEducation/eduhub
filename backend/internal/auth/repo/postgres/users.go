package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

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

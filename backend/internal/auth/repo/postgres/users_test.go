//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

func TestUserRepo_RoleByUserID_ReturnsCurrentRole(t *testing.T) {
	url := testDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() вернул ошибку: %v", err)
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO auth.users (email, role, consent_at, consent_version)
		VALUES ('moderator@example.test', 'moderator', now(), 'v1')
		RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("вставка тестового пользователя не удалась: %v", err)
	}

	repo := postgres.NewUserRepo(tx)

	role, err := repo.RoleByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("RoleByUserID() вернул ошибку: %v", err)
	}
	if role != "moderator" {
		t.Errorf("role = %q, want %q", role, "moderator")
	}
}

func TestUserRepo_RoleByUserID_Unknown_ReturnsNotFound(t *testing.T) {
	url := testDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() вернул ошибку: %v", err)
	}
	defer tx.Rollback(ctx)

	repo := postgres.NewUserRepo(tx)

	_, err = repo.RoleByUserID(ctx, uuid.New())
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

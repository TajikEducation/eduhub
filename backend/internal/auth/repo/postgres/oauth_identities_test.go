//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

func TestOAuthRepo_Create_FindByProvider_RoundTrip(t *testing.T) {
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
		INSERT INTO auth.users (email, role, status, consent_at, consent_version)
		VALUES ('oauth-roundtrip@example.test', 'user', 'active', now(), 'v1')
		RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("вставка тестового пользователя не удалась: %v", err)
	}

	repo := postgres.NewOAuthRepo(tx)

	oi := domain.OAuthIdentity{UserID: userID, Provider: "google", ProviderUserID: "google-sub-123"}
	if err := repo.Create(ctx, oi); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	found, err := repo.FindByProvider(ctx, "google", "google-sub-123")
	if err != nil {
		t.Fatalf("FindByProvider() вернул ошибку: %v", err)
	}
	if found.UserID != userID {
		t.Errorf("found.UserID = %v, want %v", found.UserID, userID)
	}
	if found.Provider != "google" {
		t.Errorf("found.Provider = %q, want %q", found.Provider, "google")
	}
	if found.ProviderUserID != "google-sub-123" {
		t.Errorf("found.ProviderUserID = %q, want %q", found.ProviderUserID, "google-sub-123")
	}
	if found.ID == uuid.Nil {
		t.Error("found.ID пуст")
	}
	if found.CreatedAt.IsZero() {
		t.Error("found.CreatedAt пуст")
	}
}

func TestOAuthRepo_FindByProvider_Unknown_ReturnsNotFound(t *testing.T) {
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

	repo := postgres.NewOAuthRepo(tx)

	_, err = repo.FindByProvider(ctx, "google", "unknown-sub")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

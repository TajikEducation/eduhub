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

func TestVerificationCodeRepo_Create_FindLatestActive_RoundTrip(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "verify1@example.test")

	repo := postgres.NewVerificationCodeRepo(tx)
	now := time.Now().Truncate(time.Microsecond)
	vc := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash1",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}

	if err := repo.Create(ctx, vc); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	found, err := repo.FindLatestActive(ctx, userID, "email", "register", now)
	if err != nil {
		t.Fatalf("FindLatestActive() вернул ошибку: %v", err)
	}
	if found.ID != vc.ID {
		t.Errorf("found.ID = %v, want %v", found.ID, vc.ID)
	}
	if found.CodeHash != vc.CodeHash {
		t.Errorf("found.CodeHash = %q, want %q", found.CodeHash, vc.CodeHash)
	}
}

func TestVerificationCodeRepo_FindLatestActive_Expired_ReturnsNotFound(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "verify2@example.test")

	repo := postgres.NewVerificationCodeRepo(tx)
	now := time.Now().Truncate(time.Microsecond)
	vc := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash-expired",
		ExpiresAt: now.Add(-time.Minute), // уже истёк
		CreatedAt: now.Add(-time.Hour),
	}
	if err := repo.Create(ctx, vc); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	_, err = repo.FindLatestActive(ctx, userID, "email", "register", now)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestVerificationCodeRepo_FindLatestActive_ReturnsMostRecent(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "verify3@example.test")

	repo := postgres.NewVerificationCodeRepo(tx)
	now := time.Now().Truncate(time.Microsecond)

	older := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash-older",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now.Add(-time.Minute),
	}
	newer := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash-newer",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	if err := repo.Create(ctx, older); err != nil {
		t.Fatalf("Create(older) вернул ошибку: %v", err)
	}
	if err := repo.Create(ctx, newer); err != nil {
		t.Fatalf("Create(newer) вернул ошибку: %v", err)
	}

	found, err := repo.FindLatestActive(ctx, userID, "email", "register", now)
	if err != nil {
		t.Fatalf("FindLatestActive() вернул ошибку: %v", err)
	}
	if found.ID != newer.ID {
		t.Errorf("found.ID = %v, want %v (самый свежий)", found.ID, newer.ID)
	}
}

func TestVerificationCodeRepo_IncrementAttempts_IncrementsByOne(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "verify4@example.test")

	repo := postgres.NewVerificationCodeRepo(tx)
	now := time.Now().Truncate(time.Microsecond)
	vc := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash-attempts",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	if err := repo.Create(ctx, vc); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	if err := repo.IncrementAttempts(ctx, vc.ID); err != nil {
		t.Fatalf("IncrementAttempts() вернул ошибку: %v", err)
	}

	found, err := repo.FindLatestActive(ctx, userID, "email", "register", now)
	if err != nil {
		t.Fatalf("FindLatestActive() вернул ошибку: %v", err)
	}
	if found.AttemptsCount != 1 {
		t.Errorf("AttemptsCount = %d, want 1", found.AttemptsCount)
	}
}

func TestVerificationCodeRepo_Delete_RemovesRow(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "verify5@example.test")

	repo := postgres.NewVerificationCodeRepo(tx)
	now := time.Now().Truncate(time.Microsecond)
	vc := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   "register",
		CodeHash:  "hash-delete",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	if err := repo.Create(ctx, vc); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	if err := repo.Delete(ctx, vc.ID); err != nil {
		t.Fatalf("Delete() вернул ошибку: %v", err)
	}

	_, err = repo.FindLatestActive(ctx, userID, "email", "register", now)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false после Delete, err = %v", err)
	}
}

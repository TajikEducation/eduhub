//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, пропускаем интеграционный тест")
	}
	return url
}

// insertTestUser вставляет минимально валидную строку auth.users (NOT NULL email/consent_at/
// consent_version из миграции 00006) и возвращает её id.
func insertTestUser(ctx context.Context, t *testing.T, tx pgx.Tx, email string) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO auth.users (email, consent_at, consent_version)
		VALUES ($1, now(), 'v1')
		RETURNING id`, email,
	).Scan(&id)
	if err != nil {
		t.Fatalf("вставка тестового пользователя не удалась: %v", err)
	}
	return id
}

func TestRefreshTokens_CreateAndFindByHash_RoundTrip(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "refresh-roundtrip@example.test")
	repo := postgres.NewRefreshTokenRepo(tx)

	rt := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: "hash-abc-123",
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Truncate(time.Microsecond),
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, rt); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	got, err := repo.FindByHash(ctx, "hash-abc-123")
	if err != nil {
		t.Fatalf("FindByHash() вернул ошибку: %v", err)
	}

	if got.ID != rt.ID {
		t.Errorf("ID = %v, want %v", got.ID, rt.ID)
	}
	if got.UserID != rt.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, rt.UserID)
	}
	if got.FamilyID != rt.FamilyID {
		t.Errorf("FamilyID = %v, want %v", got.FamilyID, rt.FamilyID)
	}
	if got.RevokedAt != nil {
		t.Errorf("RevokedAt = %v, want nil (свежий токен)", got.RevokedAt)
	}
	if got.ReplacedBy != nil {
		t.Errorf("ReplacedBy = %v, want nil", got.ReplacedBy)
	}
}

func TestRefreshTokens_FindByHash_Unknown_ReturnsNotFound(t *testing.T) {
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

	repo := postgres.NewRefreshTokenRepo(tx)

	_, err = repo.FindByHash(ctx, "неизвестный-хеш")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestRefreshTokens_Revoke_SetsRevokedAtAndReplacedBy(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "revoke@example.test")
	repo := postgres.NewRefreshTokenRepo(tx)

	family := uuid.New()
	rt := domain.RefreshToken{
		ID: uuid.New(), UserID: userID, TokenHash: "hash-revoke", FamilyID: family,
		ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, rt); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	// replaced_by — реальный FK на auth.refresh_tokens(id) (см. миграцию 00006) — нужна
	// настоящая строка-преемник, не синтетический id.
	successor := domain.RefreshToken{
		ID: uuid.New(), UserID: userID, TokenHash: "hash-revoke-successor", FamilyID: family,
		ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, successor); err != nil {
		t.Fatalf("Create() (successor) вернул ошибку: %v", err)
	}

	revokedAt := time.Now().Truncate(time.Microsecond)
	if err := repo.Revoke(ctx, rt.ID, revokedAt, &successor.ID); err != nil {
		t.Fatalf("Revoke() вернул ошибку: %v", err)
	}

	got, err := repo.FindByHash(ctx, "hash-revoke")
	if err != nil {
		t.Fatalf("FindByHash() вернул ошибку: %v", err)
	}
	if got.RevokedAt == nil || !got.RevokedAt.Equal(revokedAt) {
		t.Errorf("RevokedAt = %v, want %v", got.RevokedAt, revokedAt)
	}
	if got.ReplacedBy == nil || *got.ReplacedBy != successor.ID {
		t.Errorf("ReplacedBy = %v, want %v", got.ReplacedBy, successor.ID)
	}
}

func TestRefreshTokens_RevokeFamily_RevokesOnlyUnrevokedInSameFamily(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "family@example.test")
	repo := postgres.NewRefreshTokenRepo(tx)

	familyA := uuid.New()
	familyB := uuid.New()

	mustCreate := func(hash string, family uuid.UUID) {
		rt := domain.RefreshToken{
			ID: uuid.New(), UserID: userID, TokenHash: hash, FamilyID: family,
			ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
		}
		if err := repo.Create(ctx, rt); err != nil {
			t.Fatalf("Create(%s) вернул ошибку: %v", hash, err)
		}
	}

	mustCreate("family-a-1", familyA)
	mustCreate("family-a-2", familyA)
	mustCreate("family-b-1", familyB)

	// family-a-1 уже был отозван РАНЬШЕ (например легитимной ротацией) — своя метка времени.
	earlierRevoke := time.Now().Add(-time.Hour).Truncate(time.Microsecond)
	if err := repo.Revoke(ctx, mustFind(ctx, t, repo, "family-a-1").ID, earlierRevoke, nil); err != nil {
		t.Fatalf("Revoke(family-a-1) вернул ошибку: %v", err)
	}

	revokeFamilyAt := time.Now().Truncate(time.Microsecond)
	if err := repo.RevokeFamily(ctx, familyA, revokeFamilyAt); err != nil {
		t.Fatalf("RevokeFamily() вернул ошибку: %v", err)
	}

	a1 := mustFind(ctx, t, repo, "family-a-1")
	if a1.RevokedAt == nil || !a1.RevokedAt.Equal(earlierRevoke) {
		t.Errorf("family-a-1.RevokedAt = %v, want исходный %v (RevokeFamily не должен перезаписывать уже отозванные)", a1.RevokedAt, earlierRevoke)
	}

	a2 := mustFind(ctx, t, repo, "family-a-2")
	if a2.RevokedAt == nil || !a2.RevokedAt.Equal(revokeFamilyAt) {
		t.Errorf("family-a-2.RevokedAt = %v, want %v", a2.RevokedAt, revokeFamilyAt)
	}

	b1 := mustFind(ctx, t, repo, "family-b-1")
	if b1.RevokedAt != nil {
		t.Errorf("family-b-1.RevokedAt = %v, want nil (другая семья, не должна была затронута)", b1.RevokedAt)
	}
}

func mustFind(ctx context.Context, t *testing.T, repo *postgres.RefreshTokenRepo, hash string) domain.RefreshToken {
	t.Helper()
	rt, err := repo.FindByHash(ctx, hash)
	if err != nil {
		t.Fatalf("FindByHash(%s) вернул ошибку: %v", hash, err)
	}
	return rt
}

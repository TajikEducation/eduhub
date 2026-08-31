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

func TestUserRepo_Create_FindByEmail_FindByID_RoundTrip(t *testing.T) {
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

	displayName := "Тестовый Пользователь"
	phone := "+992900000000"
	hash := "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA"

	u := domain.User{
		Email:            "roundtrip@example.test",
		DisplayName:      &displayName,
		Locale:           "ru",
		Phone:            &phone,
		PasswordHash:     &hash,
		Role:             "user",
		Status:           "unverified",
		ConsentAt:        time.Now().Truncate(time.Microsecond),
		ConsentVersion:   "v1",
		FailedLoginCount: 0,
	}

	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("Create() вернул нулевой id")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Create() вернул нулевой CreatedAt — RETURNING не подтягивает сгенерированное БД значение")
	}
	id := created.ID

	byEmail, err := repo.FindByEmail(ctx, "roundtrip@example.test")
	if err != nil {
		t.Fatalf("FindByEmail() вернул ошибку: %v", err)
	}
	if byEmail.ID != id {
		t.Errorf("byEmail.ID = %v, want %v", byEmail.ID, id)
	}
	if byEmail.Email != u.Email {
		t.Errorf("byEmail.Email = %q, want %q", byEmail.Email, u.Email)
	}
	if byEmail.DisplayName == nil || *byEmail.DisplayName != displayName {
		t.Errorf("byEmail.DisplayName = %v, want %q", byEmail.DisplayName, displayName)
	}
	if byEmail.Phone == nil || *byEmail.Phone != phone {
		t.Errorf("byEmail.Phone = %v, want %q", byEmail.Phone, phone)
	}
	if byEmail.PasswordHash == nil || *byEmail.PasswordHash != hash {
		t.Errorf("byEmail.PasswordHash = %v, want %q", byEmail.PasswordHash, hash)
	}
	if byEmail.Role != "user" {
		t.Errorf("byEmail.Role = %q, want %q", byEmail.Role, "user")
	}
	if byEmail.Status != "unverified" {
		t.Errorf("byEmail.Status = %q, want %q", byEmail.Status, "unverified")
	}
	if !byEmail.ConsentAt.Equal(u.ConsentAt) {
		t.Errorf("byEmail.ConsentAt = %v, want %v", byEmail.ConsentAt, u.ConsentAt)
	}
	if byEmail.ConsentVersion != "v1" {
		t.Errorf("byEmail.ConsentVersion = %q, want %q", byEmail.ConsentVersion, "v1")
	}

	byID, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", err)
	}
	if byID.Email != u.Email {
		t.Errorf("byID.Email = %q, want %q", byID.Email, u.Email)
	}
}

func TestUserRepo_Create_DuplicateEmail_ReturnsConflictEmailTaken(t *testing.T) {
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

	u := domain.User{
		Email:          "duplicate@example.test",
		Locale:         "ru",
		Role:           "user",
		Status:         "unverified",
		ConsentAt:      time.Now(),
		ConsentVersion: "v1",
	}

	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() (первый) вернул ошибку: %v", err)
	}

	_, err = repo.Create(ctx, u)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "email_taken" {
		t.Errorf("Code() = %q, want %q", target.Code(), "email_taken")
	}
}

func TestUserRepo_FindByEmail_Unknown_ReturnsNotFound(t *testing.T) {
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

	_, err = repo.FindByEmail(ctx, "unknown@example.test")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestUserRepo_FindByID_Unknown_ReturnsNotFound(t *testing.T) {
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

	_, err = repo.FindByID(ctx, uuid.New())
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

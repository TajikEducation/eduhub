//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/repo/postgres"
	"github.com/abdulhalim/eduhub/backend/internal/moderation"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/pg"
)

// insertTestInstitution вставляет минимально валидную строку catalog.institutions и возвращает её id.
func insertTestInstitution(ctx context.Context, t *testing.T, tx pgx.Tx, name string) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO catalog.institutions (name, types, region, geo, moderation_status)
		VALUES (jsonb_build_object('ru', $1::text, 'tg', $1::text), '{kindergarten}', 'dushanbe',
			ST_MakePoint(68.78,38.56)::geography, 'approved')
		RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("вставка тестового учреждения не удалась: %v", err)
	}
	return id
}

func TestChildRepo_Create_Success(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "child-repo-success@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад")
	repo := postgres.NewChildRepo(tx, moderation.NewRecorder())

	c := domain.Child{
		ID:                 uuid.New(),
		UserID:             userID,
		InstitutionID:      institutionID,
		AgeGroup:           "primary",
		Status:             "current",
		ConfirmationStatus: "pending",
		CreatedAt:          time.Now().Truncate(time.Microsecond),
	}

	got, err := repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}
	if got.ConfirmationStatus != "pending" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "pending")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt пуст")
	}
}

func TestChildRepo_Create_DuplicatePair_ReturnsConflict(t *testing.T) {
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

	userID := insertTestUser(ctx, t, tx, "child-repo-dup@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 2")
	repo := postgres.NewChildRepo(tx, moderation.NewRecorder())

	mustCreate := func() {
		c := domain.Child{
			ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
			AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
			CreatedAt: time.Now(),
		}
		if _, err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create() вернул ошибку: %v", err)
		}
	}
	mustCreate()

	c2 := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "secondary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	_, err = repo.Create(ctx, c2)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_link_exists" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_link_exists")
	}
}

// setupChildRepo — общий фикстур-хелпер для тестов ListPendingByInstitution/Confirm/Reject:
// открывает пул, начинает внешнюю тестовую транзакцию (откатывается в t.Cleanup) и строит
// ChildRepo поверх неё.
func setupChildRepo(t *testing.T) (context.Context, *postgres.ChildRepo, pgx.Tx) {
	t.Helper()

	url := testDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	pool, err := pg.Open(ctx, url)
	if err != nil {
		t.Fatalf("pg.Open() вернул ошибку: %v", err)
	}
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() вернул ошибку: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	repo := postgres.NewChildRepo(tx, moderation.NewRecorder())
	return ctx, repo, tx
}

func TestChildRepo_ListPendingByInstitution_FiltersByInstitutionAndPending(t *testing.T) {
	ctx, repo, tx := setupChildRepo(t)

	userID := insertTestUser(ctx, t, tx, "child-repo-list-pending@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 3")
	otherInstitutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 4")

	pending := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, pending); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	otherInstitution := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: otherInstitutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, otherInstitution); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	got, err := repo.ListPendingByInstitution(ctx, institutionID)
	if err != nil {
		t.Fatalf("ListPendingByInstitution() вернул ошибку: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].ID != pending.ID {
		t.Errorf("got[0].ID = %v, want %v", got[0].ID, pending.ID)
	}
}

func TestChildRepo_Confirm_Success_UpdatesRowAndWritesAuditLog(t *testing.T) {
	ctx, repo, tx := setupChildRepo(t)

	userID := insertTestUser(ctx, t, tx, "child-repo-confirm@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 5")

	c := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	actorID := insertTestUser(ctx, t, tx, "child-repo-confirm-actor@example.test")
	got, err := repo.Confirm(ctx, c.ID, actorID, "moderator", "req-confirm-1")
	if err != nil {
		t.Fatalf("Confirm() вернул ошибку: %v", err)
	}
	if got.ConfirmationStatus != "confirmed" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "confirmed")
	}
	if got.ConfirmedBy == nil || *got.ConfirmedBy != actorID {
		t.Errorf("ConfirmedBy = %v, want %v", got.ConfirmedBy, actorID)
	}
	if got.ConfirmedAt == nil {
		t.Error("ConfirmedAt пуст")
	}

	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM moderation.audit_log WHERE target_id = $1 AND request_id = $2 AND action = 'child_confirmed'`,
		c.ID, "req-confirm-1").Scan(&count); err != nil {
		t.Fatalf("QueryRow() вернул ошибку: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (audit-запись должна появиться)", count)
	}
}

func TestChildRepo_Confirm_NotFound_ReturnsNotFound(t *testing.T) {
	ctx, repo, _ := setupChildRepo(t)

	_, err := repo.Confirm(ctx, uuid.New(), uuid.New(), "moderator", "req-confirm-notfound")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestChildRepo_Confirm_AlreadyConfirmed_ReturnsConflict(t *testing.T) {
	ctx, repo, tx := setupChildRepo(t)

	userID := insertTestUser(ctx, t, tx, "child-repo-confirm-twice@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 6")

	c := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	actorID := insertTestUser(ctx, t, tx, "child-repo-confirm-twice-actor@example.test")
	if _, err := repo.Confirm(ctx, c.ID, actorID, "moderator", "req-confirm-first"); err != nil {
		t.Fatalf("первый Confirm() вернул ошибку: %v", err)
	}

	_, err := repo.Confirm(ctx, c.ID, actorID, "moderator", "req-confirm-second")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_not_pending" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_not_pending")
	}
}

func TestChildRepo_Reject_Success_UpdatesRowAndWritesAuditLogWithReason(t *testing.T) {
	ctx, repo, tx := setupChildRepo(t)

	userID := insertTestUser(ctx, t, tx, "child-repo-reject@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 7")

	c := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	actorID := insertTestUser(ctx, t, tx, "child-repo-reject-actor@example.test")
	reasonText := "документ не подтверждает обучение"
	got, err := repo.Reject(ctx, c.ID, actorID, "moderator", "invalid_document", &reasonText, "req-reject-1")
	if err != nil {
		t.Fatalf("Reject() вернул ошибку: %v", err)
	}
	if got.ConfirmationStatus != "rejected" {
		t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "rejected")
	}

	var gotReasonCode, gotReasonText string
	if err := tx.QueryRow(ctx, `SELECT reason_code, reason_text FROM moderation.audit_log WHERE target_id = $1 AND request_id = $2 AND action = 'child_rejected'`,
		c.ID, "req-reject-1").Scan(&gotReasonCode, &gotReasonText); err != nil {
		t.Fatalf("QueryRow() вернул ошибку: %v", err)
	}
	if gotReasonCode != "invalid_document" {
		t.Errorf("reason_code = %q, want %q", gotReasonCode, "invalid_document")
	}
	if gotReasonText != reasonText {
		t.Errorf("reason_text = %q, want %q", gotReasonText, reasonText)
	}
}

func TestChildRepo_Reject_NotFound_ReturnsNotFound(t *testing.T) {
	ctx, repo, _ := setupChildRepo(t)

	_, err := repo.Reject(ctx, uuid.New(), uuid.New(), "moderator", "invalid_document", nil, "req-reject-notfound")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, err = %v", err)
	}
}

func TestChildRepo_Reject_AlreadyRejected_ReturnsConflict(t *testing.T) {
	ctx, repo, tx := setupChildRepo(t)

	userID := insertTestUser(ctx, t, tx, "child-repo-reject-twice@example.test")
	institutionID := insertTestInstitution(ctx, t, tx, "Тестовый сад 8")

	c := domain.Child{
		ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
		AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
		CreatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() вернул ошибку: %v", err)
	}

	actorID := insertTestUser(ctx, t, tx, "child-repo-reject-twice-actor@example.test")
	if _, err := repo.Reject(ctx, c.ID, actorID, "moderator", "invalid_document", nil, "req-reject-first"); err != nil {
		t.Fatalf("первый Reject() вернул ошибку: %v", err)
	}

	_, err := repo.Reject(ctx, c.ID, actorID, "moderator", "invalid_document", nil, "req-reject-second")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "child_not_pending" {
		t.Errorf("Code() = %q, want %q", target.Code(), "child_not_pending")
	}
}

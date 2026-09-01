//go:build integration

package moderation_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/moderation"
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

func TestRecorder_Record_CommittedTx_PersistsRow(t *testing.T) {
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

	targetID := uuid.New()
	recorder := moderation.NewRecorder()
	entry := moderation.Entry{
		ActorType:  "system",
		Action:     "institution.approve",
		TargetType: "institution",
		TargetID:   targetID,
		RequestID:  "req-commit-1",
	}

	if err := recorder.Record(ctx, tx, entry); err != nil {
		t.Fatalf("Record() вернул ошибку: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("tx.Commit() вернул ошибку: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM moderation.audit_log WHERE target_id = $1 AND request_id = $2`,
		targetID, "req-commit-1").Scan(&count); err != nil {
		t.Fatalf("QueryRow() вернул ошибку: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (запись должна быть видна после commit)", count)
	}
}

func TestRecorder_Record_RolledBackTx_LeavesNoRow(t *testing.T) {
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

	targetID := uuid.New()
	recorder := moderation.NewRecorder()
	entry := moderation.Entry{
		ActorType:  "system",
		Action:     "institution.approve",
		TargetType: "institution",
		TargetID:   targetID,
		RequestID:  "req-rollback-1",
	}

	if err := recorder.Record(ctx, tx, entry); err != nil {
		t.Fatalf("Record() вернул ошибку: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("tx.Rollback() вернул ошибку: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM moderation.audit_log WHERE target_id = $1 AND request_id = $2`,
		targetID, "req-rollback-1").Scan(&count); err != nil {
		t.Fatalf("QueryRow() вернул ошибку: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0 (rollback не должен оставлять строку)", count)
	}
}

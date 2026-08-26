package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// TestHealth_Healthz_ReturnsOKWithoutDependencies проверяет, что Healthz отвечает 200 без обращения
// к каким-либо зависимостям.
func TestHealth_Healthz_ReturnsOKWithoutDependencies(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	httpx.Healthz(log)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want %q", body["status"], "ok")
	}
}

// TestHealth_Readyz_FailingPingerReturns503WithName проверяет, что упавшая зависимость приводит к 503
// с её именем в списке failed.
func TestHealth_Readyz_FailingPingerReturns503WithName(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	dep := httpx.Dependency{
		Name: "db",
		Ping: func(ctx context.Context) error { return errors.New("connection refused") },
	}

	httpx.Readyz(log, time.Second, dep)(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body struct {
		Status string   `json:"status"`
		Failed []string `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if body.Status != "unavailable" {
		t.Fatalf("status field = %q, want %q", body.Status, "unavailable")
	}
	if len(body.Failed) != 1 || body.Failed[0] != "db" {
		t.Fatalf("failed = %v, want [\"db\"]", body.Failed)
	}
}

// TestHealth_Readyz_SuccessfulPingerReturns200 проверяет, что успешный Ping приводит к 200 ok.
func TestHealth_Readyz_SuccessfulPingerReturns200(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	dep := httpx.Dependency{
		Name: "db",
		Ping: func(ctx context.Context) error { return nil },
	}

	httpx.Readyz(log, time.Second, dep)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want %q", body["status"], "ok")
	}
}

// TestHealth_Readyz_RespectsTimeoutEvenWhenPingerIgnoresContext проверяет, что Readyz не зависает
// дольше timeout, даже когда Ping игнорирует контекст (например, блокируется на time.Sleep).
func TestHealth_Readyz_RespectsTimeoutEvenWhenPingerIgnoresContext(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	dep := httpx.Dependency{
		Name: "spin",
		Ping: func(ctx context.Context) error {
			time.Sleep(5 * time.Second)
			return nil
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		httpx.Readyz(log, 2*time.Second, dep)(rec, req)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("readyz завис дольше timeout")
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body struct {
		Status string   `json:"status"`
		Failed []string `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if len(body.Failed) != 1 || body.Failed[0] != "spin" {
		t.Fatalf("failed = %v, want [\"spin\"]", body.Failed)
	}
}

// TestHealth_Readyz_MultipleDependenciesOnlyFailingListed проверяет, что при нескольких зависимостях
// в failed попадает только упавшая, а не все.
func TestHealth_Readyz_MultipleDependenciesOnlyFailingListed(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	ok := httpx.Dependency{
		Name: "cache",
		Ping: func(ctx context.Context) error { return nil },
	}
	bad := httpx.Dependency{
		Name: "db",
		Ping: func(ctx context.Context) error { return errors.New("timeout") },
	}

	httpx.Readyz(log, time.Second, ok, bad)(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body struct {
		Status string   `json:"status"`
		Failed []string `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if len(body.Failed) != 1 || body.Failed[0] != "db" {
		t.Fatalf("failed = %v, want [\"db\"]", body.Failed)
	}
}

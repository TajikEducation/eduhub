package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// TestAccessLog_WritesExpectedFields проверяет, что мидлварь AccessLog
// пишет одну лог-запись с method/path/status/duration_ms/request_id.
func TestAccessLog_WritesExpectedFields(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.WithRequestID(httpx.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))

	req := httptest.NewRequest(http.MethodPost, "/search", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logBuf.String())
	}

	if entry["method"] != http.MethodPost {
		t.Errorf("method = %v, want %q", entry["method"], http.MethodPost)
	}
	if entry["path"] != "/search" {
		t.Errorf("path = %v, want %q", entry["path"], "/search")
	}
	// slog кодирует числа как float64 через encoding/json.
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("status = %v, want %d", entry["status"], http.StatusCreated)
	}
	if _, ok := entry["duration_ms"]; !ok {
		t.Error("log entry missing key \"duration_ms\"")
	}
	if _, ok := entry["request_id"]; !ok {
		t.Error("log entry missing key \"request_id\"")
	}
}

// TestAccessLog_DoesNotLeakQueryString проверяет кейс (б): access-log
// мидлварь пишет method/path/status/duration_ms/request_id и НЕ содержит
// query-строку целиком — GET /search?q=Саидахон не должен оставить текст
// "Саидахон" нигде в итоговой лог-записи (PII: поисковый запрос может
// содержать имя ребёнка/родителя).
func TestAccessLog_DoesNotLeakQueryString(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.WithRequestID(httpx.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/search?q=%D0%A1%D0%B0%D0%B8%D0%B4%D0%B0%D1%85%D0%BE%D0%BD", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logged := logBuf.String()
	if strings.Contains(logged, "Саидахон") {
		t.Errorf("log output contains PII query string %q: log = %q", "Саидахон", logged)
	}
	if strings.Contains(logged, "q=") {
		t.Errorf("log output contains raw query string fragment %q: log = %q", "q=", logged)
	}

	var entry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logged)
	}
	if entry["path"] != "/search" {
		t.Errorf("path = %v, want %q (without query string)", entry["path"], "/search")
	}
}

// TestAccessLog_RequestIDMatchesContext проверяет, что request_id в
// лог-записи совпадает со значением, установленным WithRequestID.
func TestAccessLog_RequestIDMatchesContext(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	const incoming = "test-request-id-99"

	handler := httpx.WithRequestID(httpx.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logBuf.String())
	}
	if entry["request_id"] != incoming {
		t.Errorf("request_id = %v, want %q", entry["request_id"], incoming)
	}
}

// TestAccessLog_DefaultStatusIsOK проверяет, что если хендлер не вызывает
// WriteHeader явно (стандартное поведение net/http — неявные 200), в лог
// попадает статус 200, а не 0.
func TestAccessLog_DefaultStatusIsOK(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.WithRequestID(httpx.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logBuf.String())
	}
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("status = %v, want %d", entry["status"], http.StatusOK)
	}
}

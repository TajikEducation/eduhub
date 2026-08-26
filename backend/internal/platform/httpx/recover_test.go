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

// TestRecover_CatchesPanicAndReturns500 проверяет кейс (а): паника внутри
// хендлера не роняет сервер — мидлварь Recover перехватывает её, отвечает
// 500 с JSON-телом {"error":{"code":"internal",...}}, а сам стек паники
// попадает в лог, не в HTTP-ответ.
func TestRecover_CatchesPanicAndReturns500(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.Recover(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Паника не должна вылезти наружу из ServeHTTP — если Recover не
	// перехватит её, test-раннер сам упадёт паникой на этой строке.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal(body) error = %v, body = %q", err, rec.Body.String())
	}
	if body.Error.Code != "internal" {
		t.Errorf("body.error.code = %q, want %q", body.Error.Code, "internal")
	}

	// Стек паники (само сообщение "boom" и данные из runtime/debug.Stack)
	// не должен попасть в тело HTTP-ответа — это внутренняя информация.
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("body contains panic message %q, want it only in logs: body = %q", "boom", rec.Body.String())
	}

	// Стек паники должен попасть в лог на уровне Error.
	logged := logBuf.String()
	if !strings.Contains(logged, "boom") {
		t.Errorf("log output does not contain panic message %q: log = %q", "boom", logged)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logged)
	}
	if logEntry["level"] != "ERROR" {
		t.Errorf("log level = %v, want %q", logEntry["level"], "ERROR")
	}
}

// TestRecover_RequestIDInLog проверяет, что паника логируется вместе с
// request_id из контекста (для корреляции с access-log) — Recover
// встраивается после WithRequestID в цепочке.
func TestRecover_RequestIDInLog(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	const incoming = "test-request-id-42"

	handler := httpx.WithRequestID(httpx.Recover(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var logEntry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logBuf.String())
	}
	if logEntry["request_id"] != incoming {
		t.Errorf("log request_id = %v, want %q", logEntry["request_id"], incoming)
	}
}

// TestRecover_NoPanicPassesThrough проверяет, что при отсутствии паники
// Recover не искажает нормальный ответ хендлера.
func TestRecover_NoPanicPassesThrough(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.Recover(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
	if logBuf.Len() != 0 {
		t.Errorf("log output = %q, want empty (no panic occurred)", logBuf.String())
	}
}

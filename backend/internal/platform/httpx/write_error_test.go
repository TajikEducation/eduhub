package httpx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// errorBody — структура тела ответа httpx.WriteError, см. контракт в
// docs/EduHub_Backend_Architecture.md, раздел 8.
type errorBody struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields,omitempty"`
	} `json:"error"`
	RequestID string `json:"request_id"`
}

// callWriteError прогоняет err через httpx.WriteError внутри цепочки
// WithRequestID (реальная мидлварь кладёт request_id в контекст запроса),
// возвращает итоговый рекордер.
func callWriteError(t *testing.T, err error) (*httptest.ResponseRecorder, *bytes.Buffer) {
	t.Helper()

	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	handler := httpx.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, log, err)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "test-request-id-42")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	return rec, &logBuf
}

func decodeErrorBody(t *testing.T, rec *httptest.ResponseRecorder) errorBody {
	t.Helper()

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal(body) error = %v, body = %q", err, rec.Body.String())
	}
	return body
}

// TestWriteError_NotFound проверяет маппинг apperr.NotFound → 404/not_found.
func TestWriteError_NotFound(t *testing.T) {
	rec, _ := callWriteError(t, apperr.NotFound("institution", "42"))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "not_found" {
		t.Errorf("code = %q, want %q", body.Error.Code, "not_found")
	}
	if body.Error.Message != "institution not found: id=42" {
		t.Errorf("message = %q, want %q", body.Error.Message, "institution not found: id=42")
	}
	if body.Error.Fields != nil {
		t.Errorf("fields = %v, want nil (omitempty)", body.Error.Fields)
	}
	if body.RequestID != "test-request-id-42" {
		t.Errorf("request_id = %q, want %q", body.RequestID, "test-request-id-42")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}
}

// TestWriteError_Invalid проверяет маппинг apperr.Invalid → 400/invalid_input
// с полями fields.
func TestWriteError_Invalid(t *testing.T) {
	fields := map[string]string{"min_price": "must be non-negative"}
	rec, _ := callWriteError(t, apperr.Invalid(fields, "validation failed"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "invalid_input" {
		t.Errorf("code = %q, want %q", body.Error.Code, "invalid_input")
	}
	if body.Error.Message != "validation failed" {
		t.Errorf("message = %q, want %q", body.Error.Message, "validation failed")
	}
	if body.Error.Fields["min_price"] != "must be non-negative" {
		t.Errorf("fields[min_price] = %q, want %q", body.Error.Fields["min_price"], "must be non-negative")
	}
}

// TestWriteError_Unauthorized проверяет маппинг apperr.Unauthorized → 401/unauthorized.
func TestWriteError_Unauthorized(t *testing.T) {
	rec, _ := callWriteError(t, apperr.Unauthorized("токен истёк"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "unauthorized" {
		t.Errorf("code = %q, want %q", body.Error.Code, "unauthorized")
	}
	if body.Error.Message != "токен истёк" {
		t.Errorf("message = %q, want %q", body.Error.Message, "токен истёк")
	}
	if body.Error.Fields != nil {
		t.Errorf("fields = %v, want nil (omitempty)", body.Error.Fields)
	}
}

// TestWriteError_Forbidden проверяет маппинг apperr.Forbidden → 403/forbidden.
func TestWriteError_Forbidden(t *testing.T) {
	rec, _ := callWriteError(t, apperr.Forbidden("недостаточно прав"))

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "forbidden" {
		t.Errorf("code = %q, want %q", body.Error.Code, "forbidden")
	}
	if body.Error.Message != "недостаточно прав" {
		t.Errorf("message = %q, want %q", body.Error.Message, "недостаточно прав")
	}
}

// TestWriteError_Conflict проверяет маппинг apperr.Conflict → 409/conflict.
func TestWriteError_Conflict(t *testing.T) {
	rec, _ := callWriteError(t, apperr.Conflict("отзыв уже существует"))

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "conflict" {
		t.Errorf("code = %q, want %q", body.Error.Code, "conflict")
	}
	if body.Error.Message != "отзыв уже существует" {
		t.Errorf("message = %q, want %q", body.Error.Message, "отзыв уже существует")
	}
}

// TestWriteError_ConflictCode проверяет, что apperr.ConflictCode переопределяет code в теле
// ответа, оставляя статус 409.
func TestWriteError_ConflictCode(t *testing.T) {
	rec, _ := callWriteError(t, apperr.ConflictCode("email_taken", "email уже занят"))

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "email_taken" {
		t.Errorf("code = %q, want %q", body.Error.Code, "email_taken")
	}
	if body.Error.Message != "email уже занят" {
		t.Errorf("message = %q, want %q", body.Error.Message, "email уже занят")
	}
}

// TestWriteError_UnauthorizedCode проверяет, что apperr.UnauthorizedCode переопределяет code
// в теле ответа, оставляя статус 401.
func TestWriteError_UnauthorizedCode(t *testing.T) {
	rec, _ := callWriteError(t, apperr.UnauthorizedCode("google_account_no_password", "войдите через Google"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "google_account_no_password" {
		t.Errorf("code = %q, want %q", body.Error.Code, "google_account_no_password")
	}
}

// TestWriteError_RateLimited проверяет маппинг apperr.RateLimited →
// 429/rate_limited с заголовком Retry-After: 60.
func TestWriteError_RateLimited(t *testing.T) {
	rec, _ := callWriteError(t, apperr.RateLimited("слишком много попыток"))

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "60" {
		t.Errorf("Retry-After = %q, want %q", got, "60")
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "rate_limited" {
		t.Errorf("code = %q, want %q", body.Error.Code, "rate_limited")
	}
	if body.Error.Message != "слишком много попыток" {
		t.Errorf("message = %q, want %q", body.Error.Message, "слишком много попыток")
	}
}

// TestWriteError_Internal проверяет маппинг apperr.Internal → 500/internal:
// клиенту уходит фиксированное сообщение (не cause), cause логируется на Error.
func TestWriteError_Internal(t *testing.T) {
	cause := errors.New("dial tcp: connection refused")
	rec, logBuf := callWriteError(t, apperr.Internal(cause))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "internal" {
		t.Errorf("code = %q, want %q", body.Error.Code, "internal")
	}
	if body.Error.Message != "internal server error" {
		t.Errorf("message = %q, want %q", body.Error.Message, "internal server error")
	}
	if body.Error.Fields != nil {
		t.Errorf("fields = %v, want nil (omitempty)", body.Error.Fields)
	}

	// apperr.Internal хранит cause отдельно для Unwrap/errors.Is — Error() его в
	// текст не подмешивает (см. apperr.go), поэтому лог сверяем с err.Error() дословно,
	// как того требует контракт логирования ("полный err.Error()"), а не с сырым cause.
	logged := logBuf.String()
	err := apperr.Internal(cause)
	if !bytes.Contains([]byte(logged), []byte(err.Error())) {
		t.Errorf("log output does not contain err.Error() %q: log = %q", err.Error(), logged)
	}

	var logEntry map[string]any
	if unmarshalErr := json.Unmarshal(logBuf.Bytes(), &logEntry); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", unmarshalErr, logged)
	}
	if logEntry["level"] != "ERROR" {
		t.Errorf("log level = %v, want %q", logEntry["level"], "ERROR")
	}
	if logEntry["request_id"] != "test-request-id-42" {
		t.Errorf("log request_id = %v, want %q", logEntry["request_id"], "test-request-id-42")
	}
}

// TestWriteError_UnknownError проверяет маппинг произвольной ошибки (не
// *apperr.Error, не матчится ни через errors.Is) → 500/internal, тем же
// фиксированным сообщением, что и Internal, плюс лог на Error.
func TestWriteError_UnknownError(t *testing.T) {
	cause := errors.New("unexpected panic-adjacent failure")
	rec, logBuf := callWriteError(t, cause)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	body := decodeErrorBody(t, rec)
	if body.Error.Code != "internal" {
		t.Errorf("code = %q, want %q", body.Error.Code, "internal")
	}
	if body.Error.Message != "internal server error" {
		t.Errorf("message = %q, want %q", body.Error.Message, "internal server error")
	}
	if body.Error.Fields != nil {
		t.Errorf("fields = %v, want nil (omitempty)", body.Error.Fields)
	}

	logged := logBuf.String()
	if !bytes.Contains([]byte(logged), []byte("unexpected panic-adjacent failure")) {
		t.Errorf("log output does not contain cause %q: log = %q", "unexpected panic-adjacent failure", logged)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("json.Unmarshal(log entry) error = %v, log = %q", err, logged)
	}
	if logEntry["level"] != "ERROR" {
		t.Errorf("log level = %v, want %q", logEntry["level"], "ERROR")
	}
}

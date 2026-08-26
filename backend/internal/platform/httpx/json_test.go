package httpx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

type jsonTestPayload struct {
	Name string `json:"name"`
}

// TestJSON_WriteJSON_ContentType проверяет, что WriteJSON ставит корректный Content-Type.
func TestJSON_WriteJSON_ContentType(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()

	if err := httpx.WriteJSON(rec, log, http.StatusOK, jsonTestPayload{Name: "test"}); err != nil {
		t.Fatalf("WriteJSON() error = %v, want nil", err)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestJSON_DecodeJSON_BodyTooLarge проверяет, что тело >1MB даёт apperr.Invalid с
// Fields["body"]=="body_too_large" и не паникует.
func TestJSON_DecodeJSON_BodyTooLarge(t *testing.T) {
	big := strings.Repeat("a", (1<<20)+1)
	body := `{"name":"` + big + `"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var dst jsonTestPayload
	err := httpx.DecodeJSON(rec, req, &dst)

	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want error")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, want true; err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err = %v", err)
	}
	if target.Fields["body"] != "body_too_large" {
		t.Errorf("Fields[body] = %q, want %q", target.Fields["body"], "body_too_large")
	}
}

// TestJSON_DecodeJSON_UnknownField проверяет, что лишнее поле в JSON даёт apperr.Invalid
// с Fields["body"]=="unknown_field".
func TestJSON_DecodeJSON_UnknownField(t *testing.T) {
	body := `{"name":"test","extra":"field"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var dst jsonTestPayload
	err := httpx.DecodeJSON(rec, req, &dst)

	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want error")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, want true; err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err = %v", err)
	}
	if target.Fields["body"] != "unknown_field" {
		t.Errorf("Fields[body] = %q, want %q", target.Fields["body"], "unknown_field")
	}
}

// TestJSON_DecodeJSON_MalformedJSON проверяет, что синтаксически битый JSON даёт
// apperr.Invalid с Fields["body"]=="malformed_json" и не паникует.
func TestJSON_DecodeJSON_MalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	rec := httptest.NewRecorder()

	var dst jsonTestPayload
	err := httpx.DecodeJSON(rec, req, &dst)

	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want error")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, want true; err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err = %v", err)
	}
	if target.Fields["body"] != "malformed_json" {
		t.Errorf("Fields[body] = %q, want %q", target.Fields["body"], "malformed_json")
	}
}

// TestJSON_DecodeJSON_EmptyBody проверяет, что пустое тело (EOF) классифицируется как
// malformed_json, а не паникует decoder'ом.
func TestJSON_DecodeJSON_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()

	var dst jsonTestPayload
	err := httpx.DecodeJSON(rec, req, &dst)

	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want error")
	}
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Errorf("errors.Is(err, apperr.ErrInvalid) = false, want true; err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err = %v", err)
	}
	if target.Fields["body"] != "malformed_json" {
		t.Errorf("Fields[body] = %q, want %q", target.Fields["body"], "malformed_json")
	}
}

// TestJSON_DecodeJSON_HappyPath проверяет корректное декодирование валидного JSON.
func TestJSON_DecodeJSON_HappyPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
	rec := httptest.NewRecorder()

	var dst jsonTestPayload
	if err := httpx.DecodeJSON(rec, req, &dst); err != nil {
		t.Fatalf("DecodeJSON() error = %v, want nil", err)
	}
	if dst.Name != "test" {
		t.Errorf("dst.Name = %q, want %q", dst.Name, "test")
	}
}

// TestJSON_WriteJSON_EmptySliceNotNull документирует контракт stdlib: non-nil пустой
// слайс сериализуется как [], не null — хендлерам важно инициализировать слайсы, а не
// оставлять их nil.
func TestJSON_WriteJSON_EmptySliceNotNull(t *testing.T) {
	type withItems struct {
		Items []string `json:"items"`
	}

	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	rec := httptest.NewRecorder()

	if err := httpx.WriteJSON(rec, log, http.StatusOK, withItems{Items: []string{}}); err != nil {
		t.Fatalf("WriteJSON() error = %v, want nil", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %q", err, rec.Body.String())
	}
	if string(decoded["items"]) != "[]" {
		t.Errorf("items = %s, want %s", decoded["items"], "[]")
	}
}

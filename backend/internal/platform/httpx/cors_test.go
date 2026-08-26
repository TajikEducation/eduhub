package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// TestCORS_PreflightAllowedOriginReturns204WithHeaders проверяет preflight-ответ для разрешённого
// Origin: 204 и Access-Control-Allow-Origin с точным значением Origin запроса.
func TestCORS_PreflightAllowedOriginReturns204WithHeaders(t *testing.T) {
	mw := httpx.CORS([]string{"https://eduhub.tj"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next.ServeHTTP не должен вызываться на preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/institutions", nil)
	req.Header.Set("Origin", "https://eduhub.tj")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://eduhub.tj" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://eduhub.tj")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods отсутствует")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers отсутствует")
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got == "" {
		t.Error("Access-Control-Max-Age отсутствует")
	}
}

// TestCORS_PreflightDisallowedOriginNoAllowHeader проверяет, что для Origin не из allowlist
// preflight всё равно отвечает 204, но без Access-Control-Allow-Origin (браузер сам заблокирует).
func TestCORS_PreflightDisallowedOriginNoAllowHeader(t *testing.T) {
	mw := httpx.CORS([]string{"https://eduhub.tj"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next.ServeHTTP не должен вызываться на preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/institutions", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestCORS_RegularRequestAllowedOriginSetsHeaderAndCallsNext проверяет, что обычный (не preflight)
// запрос с разрешённым Origin получает заголовок и next вызывается.
func TestCORS_RegularRequestAllowedOriginSetsHeaderAndCallsNext(t *testing.T) {
	mw := httpx.CORS([]string{"https://eduhub.tj"})
	var nextCalled bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/institutions", nil)
	req.Header.Set("Origin", "https://eduhub.tj")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("next.ServeHTTP не был вызван")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://eduhub.tj" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://eduhub.tj")
	}
}

// TestCORS_RegularRequestDisallowedOriginNoHeaderButCallsNext проверяет, что обычный запрос
// с Origin не из allowlist не получает CORS-заголовков, но next всё равно вызывается.
func TestCORS_RegularRequestDisallowedOriginNoHeaderButCallsNext(t *testing.T) {
	mw := httpx.CORS([]string{"https://eduhub.tj"})
	var nextCalled bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/institutions", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("next.ServeHTTP не был вызван")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

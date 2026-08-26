package httpx_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// TestRouter_ChainCallsInDeclaredOrder доказывает, что Chain не переворачивает и не путает
// порядок: первый middleware — самый внешний слой (выполняется первым и последним).
func TestRouter_ChainCallsInDeclaredOrder(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chained := httpx.Chain(mw1, mw2)(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	want := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("order = %v, want %v", order, want)
	}
}

// TestRouter_UnknownPathReturns404JSON проверяет, что незарегистрированный путь отдаёт
// JSON-контракт ошибки (не стандартный text/plain от http.ServeMux).
func TestRouter_UnknownPathReturns404JSON(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	rt := httpx.NewRouter(log)
	rt.Handle("GET /items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal(body) error = %v, body = %q", err, rec.Body.String())
	}
	if body.Error.Code != "not_found" {
		t.Errorf("body.error.code = %q, want %q", body.Error.Code, "not_found")
	}
}

// TestRouter_WrongMethodReturns405JSON проверяет, что несовпадение метода на зарегистрированном
// пути отдаёт JSON 405, а не стандартный text/plain "Method Not Allowed" от http.ServeMux.
func TestRouter_WrongMethodReturns405JSON(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	rt := httpx.NewRouter(log)
	rt.Handle("GET /items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal(body) error = %v, body = %q", err, rec.Body.String())
	}
	if body.Error.Code != "method_not_allowed" {
		t.Errorf("body.error.code = %q, want %q", body.Error.Code, "method_not_allowed")
	}
}

// TestRouter_MatchedRouteServesNormally проверяет, что при совпавшем маршруте Router
// не вмешивается — реально вызывается зарегистрированный хендлер.
func TestRouter_MatchedRouteServesNormally(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)

	rt := httpx.NewRouter(log)
	rt.Handle("GET /items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

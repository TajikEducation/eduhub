package httpx_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// doRateLimitedRequest прогоняет один запрос через переданный handler с фиксированным RemoteAddr
// одного и того же клиента.
func doRateLimitedRequest(handler http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/institutions", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// TestRateLimit_AllowsUpToLimitThenBlocks проверяет, что первые limit запросов проходят,
// а limit+1-й получает 429 с заголовком Retry-After.
func TestRateLimit_AllowsUpToLimitThenBlocks(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	fakeClock := clock.NewFake(time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))

	mw := httpx.RateLimit(log, fakeClock, 5, time.Minute)
	handler := httpx.WithRequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	for i := 1; i <= 5; i++ {
		rec := doRateLimitedRequest(handler)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}

	rec := doRateLimitedRequest(handler)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request 6: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Fatal("Retry-After header is empty")
	}
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		t.Fatalf("Retry-After = %q, not an integer: %v", retryAfter, err)
	}
	if seconds <= 0 {
		t.Errorf("Retry-After = %d, want > 0", seconds)
	}
}

// TestRateLimit_ResetsAfterWindowAdvance проверяет, что после продвижения fakeClock на window
// счётчик сбрасывается и следующий запрос снова проходит.
func TestRateLimit_ResetsAfterWindowAdvance(t *testing.T) {
	var logBuf bytes.Buffer
	log := logger.New("info", "eduhub-test", &logBuf)
	fakeClock := clock.NewFake(time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))

	mw := httpx.RateLimit(log, fakeClock, 5, time.Minute)
	handler := httpx.WithRequestID(mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	for i := 1; i <= 5; i++ {
		doRateLimitedRequest(handler)
	}

	blocked := doRateLimitedRequest(handler)
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("request 6: status = %d, want %d", blocked.Code, http.StatusTooManyRequests)
	}

	fakeClock.Advance(time.Minute)

	rec := doRateLimitedRequest(handler)
	if rec.Code != http.StatusOK {
		t.Errorf("request after window advance: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

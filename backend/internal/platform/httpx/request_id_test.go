package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// uuidV4Pattern — простая проверка формата UUID v4 (8-4-4-4-12 hex,
// версия "4" в третьей группе, вариант "8/9/a/b" в четвёртой).
var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// TestRequestID_GeneratesWhenHeaderAbsent проверяет кейс (а): без
// входящего заголовка X-Request-ID мидлварь генерирует UUID, кладёт его и
// в заголовок ответа X-Request-ID, и в контекст запроса (доступен внутри
// хендлера через httpx.RequestID).
func TestRequestID_GeneratesWhenHeaderAbsent(t *testing.T) {
	var gotFromCtx string

	handler := httpx.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFromCtx = httpx.RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	gotHeader := rec.Header().Get("X-Request-ID")
	if gotHeader == "" {
		t.Fatal("X-Request-ID header is empty, want generated UUID")
	}
	if !uuidV4Pattern.MatchString(gotHeader) {
		t.Errorf("X-Request-ID header = %q, want UUID v4 format", gotHeader)
	}
	if gotFromCtx != gotHeader {
		t.Errorf("RequestID(ctx) = %q, want it to match response header %q", gotFromCtx, gotHeader)
	}
}

// TestRequestID_PreservesIncomingHeader проверяет кейс (б): входящий
// заголовок X-Request-ID сохраняется как есть, не перегенерируется.
func TestRequestID_PreservesIncomingHeader(t *testing.T) {
	const incoming = "client-supplied-id-123"

	var gotFromCtx string

	handler := httpx.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFromCtx = httpx.RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != incoming {
		t.Errorf("X-Request-ID header = %q, want preserved %q", got, incoming)
	}
	if gotFromCtx != incoming {
		t.Errorf("RequestID(ctx) = %q, want preserved %q", gotFromCtx, incoming)
	}
}

// TestRequestID_EmptyOnBareContext проверяет кейс (в): RequestID на
// контексте без установленного request_id возвращает пустую строку, не
// паникует.
func TestRequestID_EmptyOnBareContext(t *testing.T) {
	got := httpx.RequestID(context.Background())
	if got != "" {
		t.Errorf("RequestID(context.Background()) = %q, want empty string", got)
	}
}

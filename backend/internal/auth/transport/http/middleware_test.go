package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

func TestFromContext_EmptyContext_ReturnsFalse(t *testing.T) {
	_, ok := FromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context())
	if ok {
		t.Error("FromContext(пустой контекст) = true, want false")
	}
}

func TestRequireAuth(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC))
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	userID := uuid.New()

	var capturedPrincipal Principal
	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		capturedPrincipal, _ = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	t.Run("валидный Bearer вызывает хендлер за миддлварой и кладёт Principal в контекст", func(t *testing.T) {
		nextCalled = false
		token, err := issuer.Issue(userID, "user")
		if err != nil {
			t.Fatalf("Issue() вернул ошибку: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		RequireAuth(issuer, testLogger())(next).ServeHTTP(rec, req)

		if !nextCalled {
			t.Fatal("хендлер за миддлварой не вызван")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if capturedPrincipal.UserID != userID {
			t.Errorf("Principal.UserID = %v, want %v", capturedPrincipal.UserID, userID)
		}
		if capturedPrincipal.Role != "user" {
			t.Errorf("Principal.Role = %q, want %q", capturedPrincipal.Role, "user")
		}
	})

	t.Run("отсутствующий заголовок Authorization возвращает 401, хендлер не вызывается", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		rec := httptest.NewRecorder()

		RequireAuth(issuer, testLogger())(next).ServeHTTP(rec, req)

		if nextCalled {
			t.Fatal("хендлер за миддлварой вызван, не должен")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("битый токен возвращает 401, хендлер не вызывается", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer not-a-real-token")
		rec := httptest.NewRecorder()

		RequireAuth(issuer, testLogger())(next).ServeHTTP(rec, req)

		if nextCalled {
			t.Fatal("хендлер за миддлварой вызван, не должен")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("просроченный токен возвращает 401, хендлер не вызывается", func(t *testing.T) {
		nextCalled = false
		token, err := issuer.Issue(userID, "user")
		if err != nil {
			t.Fatalf("Issue() вернул ошибку: %v", err)
		}
		clk.Advance(16 * time.Minute)
		defer clk.Advance(-16 * time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		RequireAuth(issuer, testLogger())(next).ServeHTTP(rec, req)

		if nextCalled {
			t.Fatal("хендлер за миддлварой вызван, не должен")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

// TestRequireRole_Matrix — табличный тест матрицы «роль × доступ» (E2.5, план прямо требует
// такой тест). Пока в реальных роутах нет ни одного ролевого эндпоинта (появятся с вехи 3 —
// модерация, вехи 6 — админ-действия), таблица растёт по мере их добавления, сейчас проверяет
// сам механизм RequireRole независимо от конкретного бизнес-роута.
func TestRequireRole_Matrix(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC))
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)

	tests := []struct {
		name          string
		allowedRoles  []string
		principalRole string
		wantAllowed   bool
	}{
		{"admin-роут, вызывающий admin — пропущен", []string{"admin"}, "admin", true},
		{"admin-роут, вызывающий moderator — отказ", []string{"admin"}, "moderator", false},
		{"admin-роут, вызывающий user — отказ", []string{"admin"}, "user", false},
		{"multi-role роут (moderator|admin), вызывающий moderator — пропущен", []string{"moderator", "admin"}, "moderator", true},
		{"multi-role роут (moderator|admin), вызывающий admin — пропущен", []string{"moderator", "admin"}, "admin", true},
		{"multi-role роут (moderator|admin), вызывающий user — отказ", []string{"moderator", "admin"}, "user", false},
		{"institution-роут, вызывающий institution — пропущен", []string{"institution"}, "institution", true},
		{"institution-роут, вызывающий guest — отказ", []string{"institution"}, "guest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := issuer.Issue(uuid.New(), tt.principalRole)
			if err != nil {
				t.Fatalf("Issue() вернул ошибку: %v", err)
			}

			var nextCalled bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/some/protected/route", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			handler := RequireAuth(issuer, testLogger())(RequireRole(testLogger(), tt.allowedRoles...)(next))
			handler.ServeHTTP(rec, req)

			if nextCalled != tt.wantAllowed {
				t.Errorf("nextCalled = %v, want %v", nextCalled, tt.wantAllowed)
			}
			wantStatus := http.StatusOK
			if !tt.wantAllowed {
				wantStatus = http.StatusForbidden
			}
			if rec.Code != wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, wantStatus)
			}
		})
	}
}

func TestRequireRole_NoPrincipalInContext_ReturnsInternalNotPanic(t *testing.T) {
	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// RequireRole вызван БЕЗ предшествующего RequireAuth — Principal в контексте нет.
	req := httptest.NewRequest(http.MethodGet, "/some/protected/route", nil)
	rec := httptest.NewRecorder()

	RequireRole(testLogger(), "admin")(next).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("хендлер за миддлварой вызван, не должен")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (баг wiring — RequireRole без RequireAuth — не должен ронять процесс паникой)", rec.Code, http.StatusInternalServerError)
	}
}

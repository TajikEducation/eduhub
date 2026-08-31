package rbac_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_MissingHeaderReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	handler := rbac.RequireAuth("secret", log)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_InvalidTokenReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	handler := rbac.RequireAuth("secret", log)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidTokenPutsPrincipalInContext(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	userID := uuid.New()
	token, err := security.IssueAccessToken("secret", userID, "moderator", time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken() unexpected error: %v", err)
	}

	var gotPrincipal rbac.Principal
	var gotOK bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPrincipal, gotOK = rbac.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := rbac.RequireAuth("secret", log)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !gotOK {
		t.Fatal("FromContext() ok = false, want true")
	}
	if gotPrincipal.UserID != userID {
		t.Errorf("Principal.UserID = %v, want %v", gotPrincipal.UserID, userID)
	}
	if gotPrincipal.Role != "moderator" {
		t.Errorf("Principal.Role = %q, want %q", gotPrincipal.Role, "moderator")
	}
}

func TestRequireRole_MatchingRolePasses(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	handler := rbac.RequireRole(log, "admin", "moderator")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "moderator"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireRole_WrongRoleReturns403(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	handler := rbac.RequireRole(log, "admin")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: uuid.New(), Role: "user"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireRole_NoPrincipalReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	handler := rbac.RequireRole(log, "admin")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (RequireRole without RequireAuth is a routing bug, not a 403)", rec.Code, http.StatusUnauthorized)
	}
}

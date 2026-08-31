package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// fakeAuthService — тестовая реализация authService: возвращает заранее заданные результаты/ошибки,
// фиксирует последние полученные аргументы.
type fakeAuthService struct {
	registerInput domain.RegisterInput
	registerUser  domain.User
	registerPair  usecase.TokenPair
	registerErr   error

	loginEmail    string
	loginPassword string
	loginUser     domain.User
	loginPair     usecase.TokenPair
	loginErr      error

	refreshRawToken string
	refreshPair     usecase.TokenPair
	refreshErr      error

	meUserID uuid.UUID
	meUser   domain.User
	meErr    error
}

func (f *fakeAuthService) Register(_ context.Context, in domain.RegisterInput) (domain.User, usecase.TokenPair, error) {
	f.registerInput = in
	return f.registerUser, f.registerPair, f.registerErr
}

func (f *fakeAuthService) Login(_ context.Context, email, password string) (domain.User, usecase.TokenPair, error) {
	f.loginEmail, f.loginPassword = email, password
	return f.loginUser, f.loginPair, f.loginErr
}

func (f *fakeAuthService) Refresh(_ context.Context, rawToken string) (usecase.TokenPair, error) {
	f.refreshRawToken = rawToken
	return f.refreshPair, f.refreshErr
}

func (f *fakeAuthService) Me(_ context.Context, userID uuid.UUID) (domain.User, error) {
	f.meUserID = userID
	return f.meUser, f.meErr
}

func TestRegisterHandler_ValidBodyReturns201(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{
		registerUser: domain.User{ID: uuid.New(), Email: "new@example.com", Role: domain.RoleUser, Status: domain.StatusActive},
		registerPair: usecase.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900},
	}

	body := bytes.NewBufferString(`{"email":"new@example.com","password":"password123","display_name":"New"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	rec := httptest.NewRecorder()

	RegisterHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if fake.registerInput.Email != "new@example.com" {
		t.Errorf("registerInput.Email = %q, want %q", fake.registerInput.Email, "new@example.com")
	}

	var resp authResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Tokens.AccessToken != "access" {
		t.Errorf("resp.Tokens.AccessToken = %q, want %q", resp.Tokens.AccessToken, "access")
	}
}

func TestRegisterHandler_ServiceErrorPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{registerErr: apperr.Conflict("email_taken")}

	body := bytes.NewBufferString(`{"email":"dup@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	rec := httptest.NewRecorder()

	RegisterHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRegisterHandler_MalformedBodyReturns400(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`not json`))
	rec := httptest.NewRecorder()

	RegisterHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoginHandler_ValidCredentialsReturns200(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{
		loginUser: domain.User{ID: uuid.New(), Email: "u@example.com"},
		loginPair: usecase.TokenPair{AccessToken: "access", RefreshToken: "refresh"},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"u@example.com","password":"pw"}`))
	rec := httptest.NewRecorder()

	LoginHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if fake.loginEmail != "u@example.com" || fake.loginPassword != "pw" {
		t.Errorf("login called with (%q,%q), want (%q,%q)", fake.loginEmail, fake.loginPassword, "u@example.com", "pw")
	}
}

func TestLoginHandler_UnauthorizedPropagates(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{loginErr: apperr.Unauthorized("неверный email или пароль")}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"u@example.com","password":"wrong"}`))
	rec := httptest.NewRecorder()

	LoginHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRefreshHandler_ValidTokenReturns200(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{refreshPair: usecase.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"old-refresh"}`))
	rec := httptest.NewRecorder()

	RefreshHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if fake.refreshRawToken != "old-refresh" {
		t.Errorf("refreshRawToken = %q, want %q", fake.refreshRawToken, "old-refresh")
	}
}

func TestMeHandler_WithPrincipalReturns200(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	userID := uuid.New()
	fake := &fakeAuthService{meUser: domain.User{ID: userID, Email: "me@example.com"}}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(rbac.NewContext(req.Context(), rbac.Principal{UserID: userID, Role: "user"}))
	rec := httptest.NewRecorder()

	MeHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if fake.meUserID != userID {
		t.Errorf("meUserID = %v, want %v", fake.meUserID, userID)
	}
}

func TestMeHandler_WithoutPrincipalReturns401(t *testing.T) {
	log := logger.New("info", "test", io.Discard)
	fake := &fakeAuthService{}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	MeHandler(fake, log).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

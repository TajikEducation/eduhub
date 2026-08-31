package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

// fakeAccountService — тестовый двойник accountService.
type fakeAccountService struct {
	registerUser domain.User
	registerErr  error

	loginAccess  string
	loginRefresh string
	loginErr     error

	meUser domain.User
	meErr  error

	googleAccess  string
	googleRefresh string
	googleErr     error

	updateConsentErr error
	deleteMeErr      error
}

func (f *fakeAccountService) Register(_ context.Context, _, _, _ string) (domain.User, error) {
	return f.registerUser, f.registerErr
}

func (f *fakeAccountService) Login(_ context.Context, _, _ string) (string, string, error) {
	return f.loginAccess, f.loginRefresh, f.loginErr
}

func (f *fakeAccountService) Me(_ context.Context, _ uuid.UUID) (domain.User, error) {
	return f.meUser, f.meErr
}

func (f *fakeAccountService) LoginWithGoogle(_ context.Context, _, _ string) (string, string, error) {
	return f.googleAccess, f.googleRefresh, f.googleErr
}

func (f *fakeAccountService) UpdateConsent(_ context.Context, _ uuid.UUID, _ string) error {
	return f.updateConsentErr
}

func (f *fakeAccountService) DeleteMe(_ context.Context, _ uuid.UUID) error {
	return f.deleteMeErr
}

// fakeSessionService — тестовый двойник sessionService.
type fakeSessionService struct {
	rotateAccess  string
	rotateRefresh string
	rotateErr     error

	logoutErr error
}

func (f *fakeSessionService) Rotate(_ context.Context, _ string) (string, string, error) {
	return f.rotateAccess, f.rotateRefresh, f.rotateErr
}

func (f *fakeSessionService) Logout(_ context.Context, _ string) error {
	return f.logoutErr
}

// fakeVerificationService — тестовый двойник verificationService.
type fakeVerificationService struct {
	requestEmailVerificationErr error
	verifyEmailErr              error
	requestPasswordResetErr     error
	confirmPasswordResetErr     error
}

func (f *fakeVerificationService) RequestEmailVerification(_ context.Context, _ string) error {
	return f.requestEmailVerificationErr
}

func (f *fakeVerificationService) VerifyEmail(_ context.Context, _, _ string) error {
	return f.verifyEmailErr
}

func (f *fakeVerificationService) RequestPasswordReset(_ context.Context, _ string) error {
	return f.requestPasswordResetErr
}

func (f *fakeVerificationService) ConfirmPasswordReset(_ context.Context, _, _, _ string) error {
	return f.confirmPasswordResetErr
}

func TestRegisterHandler(t *testing.T) {
	t.Run("невалидный email возвращает 400 с полем email", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/register", jsonBody(t, registerRequest{
			Email: "not-an-email", Password: "password123", ConsentVersion: "v1",
		}))
		rec := httptest.NewRecorder()

		RegisterHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := body["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["email"]; !ok {
			t.Errorf("fields не содержит email: %v", fields)
		}
	})

	t.Run("короткий пароль возвращает 400 с полем password", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/register", jsonBody(t, registerRequest{
			Email: "test@example.com", Password: "short", ConsentVersion: "v1",
		}))
		rec := httptest.NewRecorder()

		RegisterHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := body["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["password"]; !ok {
			t.Errorf("fields не содержит password: %v", fields)
		}
	})

	t.Run("пустой consent_version возвращает 400 с полем consent_version", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/register", jsonBody(t, registerRequest{
			Email: "test@example.com", Password: "password123", ConsentVersion: "",
		}))
		rec := httptest.NewRecorder()

		RegisterHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := body["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["consent_version"]; !ok {
			t.Errorf("fields не содержит consent_version: %v", fields)
		}
	})

	t.Run("успех возвращает 201 с телом", func(t *testing.T) {
		id := uuid.New()
		createdAt := time.Now()
		svc := &fakeAccountService{registerUser: domain.User{
			ID: id, Email: "test@example.com", Role: "user", Status: "unverified", CreatedAt: createdAt,
		}}
		req := httptest.NewRequest(http.MethodPost, "/auth/register", jsonBody(t, registerRequest{
			Email: "test@example.com", Password: "password123", ConsentVersion: "v1",
		}))
		rec := httptest.NewRecorder()

		RegisterHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
		}
		var parsed registerResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.ID != id {
			t.Errorf("ID = %v, want %v", parsed.ID, id)
		}
		if parsed.Email != "test@example.com" {
			t.Errorf("Email = %q, want %q", parsed.Email, "test@example.com")
		}
	})

	t.Run("конфликт email возвращает 409 с кодом email_taken", func(t *testing.T) {
		svc := &fakeAccountService{registerErr: apperr.ConflictCode("email_taken", "email уже зарегистрирован")}
		req := httptest.NewRequest(http.MethodPost, "/auth/register", jsonBody(t, registerRequest{
			Email: "test@example.com", Password: "password123", ConsentVersion: "v1",
		}))
		rec := httptest.NewRecorder()

		RegisterHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := body["error"].(map[string]any)
		if errObj["code"] != "email_taken" {
			t.Errorf("code = %v, want %q", errObj["code"], "email_taken")
		}
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("успех возвращает 200 с токенами", func(t *testing.T) {
		svc := &fakeAccountService{loginAccess: "access-tok", loginRefresh: "refresh-tok"}
		req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, loginRequest{
			Email: "test@example.com", Password: "password123",
		}))
		rec := httptest.NewRecorder()

		LoginHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var parsed tokenResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.AccessToken != "access-tok" || parsed.RefreshToken != "refresh-tok" {
			t.Errorf("tokens = %+v, want access-tok/refresh-tok", parsed)
		}
		if parsed.TokenType != "Bearer" {
			t.Errorf("TokenType = %q, want %q", parsed.TokenType, "Bearer")
		}
	})

	t.Run("ошибка сервиса пробрасывается как есть через httpx.WriteError", func(t *testing.T) {
		svc := &fakeAccountService{loginErr: apperr.Unauthorized("неверный email или пароль")}
		req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, loginRequest{
			Email: "test@example.com", Password: "wrong",
		}))
		rec := httptest.NewRecorder()

		LoginHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})

	t.Run("пустой email/password возвращает 400", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/login", jsonBody(t, loginRequest{}))
		rec := httptest.NewRecorder()

		LoginHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestRefreshHandler(t *testing.T) {
	t.Run("пустой refresh_token возвращает 400", func(t *testing.T) {
		svc := &fakeSessionService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", jsonBody(t, refreshRequest{}))
		rec := httptest.NewRecorder()

		RefreshHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("успех возвращает 200 tokenResponse", func(t *testing.T) {
		svc := &fakeSessionService{rotateAccess: "new-access", rotateRefresh: "new-refresh"}
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", jsonBody(t, refreshRequest{RefreshToken: "old-refresh"}))
		rec := httptest.NewRecorder()

		RefreshHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var parsed tokenResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.AccessToken != "new-access" || parsed.RefreshToken != "new-refresh" {
			t.Errorf("tokens = %+v", parsed)
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	t.Run("пустой refresh_token возвращает 400", func(t *testing.T) {
		svc := &fakeSessionService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", jsonBody(t, logoutRequest{}))
		rec := httptest.NewRecorder()

		LogoutHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("успех возвращает 200 status ok", func(t *testing.T) {
		svc := &fakeSessionService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", jsonBody(t, logoutRequest{RefreshToken: "some-token"}))
		rec := httptest.NewRecorder()

		LogoutHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var parsed logoutResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.Status != "ok" {
			t.Errorf("Status = %q, want %q", parsed.Status, "ok")
		}
	})
}

func TestOAuthGoogleHandler(t *testing.T) {
	t.Run("пустой id_token возвращает 400", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/google", jsonBody(t, oauthGoogleRequest{}))
		rec := httptest.NewRecorder()

		OAuthGoogleHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("успех возвращает 200 с токенами", func(t *testing.T) {
		svc := &fakeAccountService{googleAccess: "g-access", googleRefresh: "g-refresh"}
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/google", jsonBody(t, oauthGoogleRequest{
			IDToken: "raw-id-token", ConsentVersion: "v1",
		}))
		rec := httptest.NewRecorder()

		OAuthGoogleHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var parsed tokenResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.AccessToken != "g-access" || parsed.RefreshToken != "g-refresh" {
			t.Errorf("tokens = %+v, want g-access/g-refresh", parsed)
		}
	})

	t.Run("ошибка сервиса пробрасывается как есть через httpx.WriteError", func(t *testing.T) {
		svc := &fakeAccountService{googleErr: apperr.Unauthorized("невалидный Google id-токен")}
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/google", jsonBody(t, oauthGoogleRequest{IDToken: "bad-token"}))
		rec := httptest.NewRecorder()

		OAuthGoogleHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})
}

func TestMeHandler(t *testing.T) {
	t.Run("без Principal в контексте не паникует, отдаёт 500", func(t *testing.T) {
		svc := &fakeAccountService{}
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		rec := httptest.NewRecorder()

		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("MeHandler запаниковал: %v", r)
				}
			}()
			MeHandler(svc, testLogger()).ServeHTTP(rec, req.WithContext(context.Background()))
		}()

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
		}
	})

	t.Run("с Principal возвращает 200 meResponse", func(t *testing.T) {
		id := uuid.New()
		svc := &fakeAccountService{meUser: domain.User{ID: id, Email: "test@example.com", Role: "user", Status: "active"}}
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		ctx := context.WithValue(req.Context(), principalContextKey{}, Principal{UserID: id, Role: "user"})
		rec := httptest.NewRecorder()

		MeHandler(svc, testLogger()).ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var parsed meResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.ID != id {
			t.Errorf("ID = %v, want %v", parsed.ID, id)
		}
	})
}

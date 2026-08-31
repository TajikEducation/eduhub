package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/googleoauth"
	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// fakeGoogleVerifier — тестовый двойник usecase.GoogleIDTokenVerifier.
type fakeGoogleVerifier struct {
	claims googleoauth.Claims
	err    error
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, _ string) (googleoauth.Claims, error) {
	return f.claims, f.err
}

// fakeOAuthRepo — тестовый двойник usecase.OAuthIdentityRepo, in-memory по (provider,providerUserID).
type fakeOAuthRepo struct {
	byKey map[string]domain.OAuthIdentity

	createCalls []domain.OAuthIdentity
}

func newFakeOAuthRepo() *fakeOAuthRepo {
	return &fakeOAuthRepo{byKey: map[string]domain.OAuthIdentity{}}
}

func (f *fakeOAuthRepo) FindByProvider(_ context.Context, provider, providerUserID string) (domain.OAuthIdentity, error) {
	oi, ok := f.byKey[provider+":"+providerUserID]
	if !ok {
		return domain.OAuthIdentity{}, apperr.NotFound("oauth_identity", provider+":"+providerUserID)
	}
	return oi, nil
}

func (f *fakeOAuthRepo) Create(_ context.Context, oi domain.OAuthIdentity) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	f.createCalls = append(f.createCalls, oi)
	f.byKey[oi.Provider+":"+oi.ProviderUserID] = oi
	return nil
}

// newTestAccountServiceWithGoogle — тот же набор фейков, что newTestAccountService, плюс
// настраиваемые googleVerifier/oauthRepo для тестов LoginWithGoogle.
func newTestAccountServiceWithGoogle(users *fakeUserRepo, clk *clock.Fake, googleVerifier usecase.GoogleIDTokenVerifier, oauthRepo *fakeOAuthRepo) *usecase.AccountService {
	sessionRepo := newFakeRefreshTokenRepo()
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	sessions := usecase.NewSessionService(sessionRepo, fakeUserRoleLookup{role: "user"}, issuer, clk, refreshTTL)
	hasher := password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	return usecase.NewAccountService(users, hasher, sessions, clk, googleVerifier, oauthRepo)
}

// (а) существующая связка identity → сессия для правильного пользователя, FindByEmail/Create
// не должны вызываться на UserRepo.
func TestLoginWithGoogle_ExistingIdentity_IssuesSessionForLinkedUser(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	existing := users.put(domain.User{Email: "known@example.com", Role: "user", Status: "active"})

	oauthRepo := newFakeOAuthRepo()
	oauthRepo.byKey["google:sub-1"] = domain.OAuthIdentity{ID: uuid.New(), UserID: existing.ID, Provider: "google", ProviderUserID: "sub-1"}

	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-1", Email: "known@example.com", EmailVerified: true}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	access, refresh, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "")
	if err != nil {
		t.Fatalf("LoginWithGoogle() вернул ошибку: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("токены пусты")
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
	if len(oauthRepo.createCalls) != 0 {
		t.Errorf("oauthRepo.Create() вызван %d раз, want 0", len(oauthRepo.createCalls))
	}
}

// (б) нет связки, есть пользователь с тем же email, EmailVerified=true → линковка, сессия для
// существующего пользователя.
func TestLoginWithGoogle_NoIdentityExistingEmailVerified_LinksAndIssuesSession(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	existing := users.put(domain.User{Email: "known@example.com", Role: "user", Status: "active"})

	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-2", Email: "known@example.com", EmailVerified: true}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	access, refresh, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "")
	if err != nil {
		t.Fatalf("LoginWithGoogle() вернул ошибку: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("токены пусты")
	}
	if len(oauthRepo.createCalls) != 1 {
		t.Fatalf("oauthRepo.Create() вызван %d раз, want 1", len(oauthRepo.createCalls))
	}
	linked := oauthRepo.createCalls[0]
	if linked.UserID != existing.ID {
		t.Errorf("linked.UserID = %v, want %v", linked.UserID, existing.ID)
	}
	if linked.ProviderUserID != "sub-2" {
		t.Errorf("linked.ProviderUserID = %q, want %q", linked.ProviderUserID, "sub-2")
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
}

// (в) нет связки, есть пользователь с тем же email, EmailVerified=false → отказ, ничего не создаётся.
func TestLoginWithGoogle_NoIdentityExistingEmailUnverified_Refuses(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	users.put(domain.User{Email: "known@example.com", Role: "user", Status: "active"})

	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-3", Email: "known@example.com", EmailVerified: false}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	_, _, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, apperr.ErrUnauthorized) = false, err = %v", err)
	}
	if len(oauthRepo.createCalls) != 0 {
		t.Errorf("oauthRepo.Create() вызван %d раз, want 0", len(oauthRepo.createCalls))
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
}

// (г) нет связки, нет пользователя с email, EmailVerified=true, consentVersion задан → новый
// пользователь создан со Status="active", PasswordHash=nil, EmailVerifiedAt не nil,
// oauthRepo.Create вызван, сессия выпущена.
func TestLoginWithGoogle_NewUser_CreatesActiveUserAndIssuesSession(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-4", Email: "new@example.com", EmailVerified: true}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	access, refresh, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "v1")
	if err != nil {
		t.Fatalf("LoginWithGoogle() вернул ошибку: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("токены пусты")
	}
	if len(users.createCalls) != 1 {
		t.Fatalf("users.Create() вызван %d раз, want 1", len(users.createCalls))
	}
	created := users.createCalls[0]
	if created.Status != "active" {
		t.Errorf("created.Status = %q, want %q", created.Status, "active")
	}
	if created.PasswordHash != nil {
		t.Errorf("created.PasswordHash = %v, want nil", created.PasswordHash)
	}
	if created.EmailVerifiedAt == nil {
		t.Error("created.EmailVerifiedAt == nil, want не-nil")
	}
	if created.Email != "new@example.com" {
		t.Errorf("created.Email = %q, want %q", created.Email, "new@example.com")
	}
	if len(oauthRepo.createCalls) != 1 {
		t.Fatalf("oauthRepo.Create() вызван %d раз, want 1", len(oauthRepo.createCalls))
	}
	if oauthRepo.createCalls[0].ProviderUserID != "sub-4" {
		t.Errorf("linked.ProviderUserID = %q, want %q", oauthRepo.createCalls[0].ProviderUserID, "sub-4")
	}
}

// (д) тот же кейс (г), но consentVersion="" → apperr.Invalid, ничего не создано.
func TestLoginWithGoogle_NewUserWithoutConsentVersion_ReturnsInvalid(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-5", Email: "new2@example.com", EmailVerified: true}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	_, _, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
	if len(oauthRepo.createCalls) != 0 {
		t.Errorf("oauthRepo.Create() вызван %d раз, want 0", len(oauthRepo.createCalls))
	}
}

// (е) нет связки, нет пользователя, EmailVerified=false → отказ, ничего не создано (даже если
// consentVersion задан).
func TestLoginWithGoogle_NewUserEmailUnverified_RefusesEvenWithConsentVersion(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{claims: googleoauth.Claims{Subject: "sub-6", Email: "new3@example.com", EmailVerified: false}}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	_, _, err := svc.LoginWithGoogle(context.Background(), "irrelevant-raw-token", "v1")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, apperr.ErrUnauthorized) = false, err = %v", err)
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
	if len(oauthRepo.createCalls) != 0 {
		t.Errorf("oauthRepo.Create() вызван %d раз, want 0", len(oauthRepo.createCalls))
	}
}

// (ж) googleVerifier.Verify возвращает ошибку (битый токен) → apperr.Unauthorized, репозитории
// не трогаются.
func TestLoginWithGoogle_VerifierError_ReturnsUnauthorized(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	verifier := &fakeGoogleVerifier{err: googleoauth.ErrInvalidIDToken}
	svc := newTestAccountServiceWithGoogle(users, clk, verifier, oauthRepo)

	_, _, err := svc.LoginWithGoogle(context.Background(), "битый-токен", "v1")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, apperr.ErrUnauthorized) = false, err = %v", err)
	}
	if len(users.createCalls) != 0 {
		t.Errorf("users.Create() вызван %d раз, want 0", len(users.createCalls))
	}
	if len(oauthRepo.createCalls) != 0 {
		t.Errorf("oauthRepo.Create() вызван %d раз, want 0", len(oauthRepo.createCalls))
	}
}

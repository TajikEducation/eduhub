package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// fakeUserRepo — тестовый двойник usecase.UserRepo поверх in-memory map.
type fakeUserRepo struct {
	byEmail map[string]domain.User
	byID    map[uuid.UUID]domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]domain.User{}, byID: map[uuid.UUID]domain.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u domain.User) (domain.User, error) {
	if _, exists := f.byEmail[u.Email]; exists {
		return domain.User{}, apperr.Conflict("email_taken")
	}
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, apperr.NotFound("user", email)
	}
	return u, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return domain.User{}, apperr.NotFound("user", id.String())
	}
	return u, nil
}

// fakeRefreshTokenRepo — тестовый двойник usecase.RefreshTokenRepo поверх in-memory map.
type fakeRefreshTokenRepo struct {
	byHash map[string]usecase.RefreshTokenRecord
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{byHash: map[string]usecase.RefreshTokenRecord{}}
}

func (f *fakeRefreshTokenRepo) Store(_ context.Context, t usecase.RefreshTokenRecord) error {
	f.byHash[t.TokenHash] = t
	return nil
}

func (f *fakeRefreshTokenRepo) GetByHash(_ context.Context, hash string) (usecase.RefreshTokenRecord, error) {
	rec, ok := f.byHash[hash]
	if !ok {
		return usecase.RefreshTokenRecord{}, apperr.Unauthorized("refresh token not found")
	}
	return rec, nil
}

func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id uuid.UUID, replacedBy uuid.UUID) error {
	for hash, rec := range f.byHash {
		if rec.ID == id {
			now := time.Now()
			rec.RevokedAt = &now
			rec.ReplacedBy = &replacedBy
			f.byHash[hash] = rec
		}
	}
	return nil
}

func newTestService(users usecase.UserRepo, tokens usecase.RefreshTokenRepo, clk clock.Clock) *usecase.Service {
	return usecase.New(users, tokens, clk, usecase.Config{
		JWTSecret:  "test-secret",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	})
}

func TestService_Register_CreatesUserAndIssuesTokens(t *testing.T) {
	svc := newTestService(newFakeUserRepo(), newFakeRefreshTokenRepo(), clock.New())

	user, tokens, err := svc.Register(context.Background(), domain.RegisterInput{
		Email: "new@example.com", Password: "password123", DisplayName: "New User",
	})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	if user.Email != "new@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "new@example.com")
	}
	if user.Role != domain.RoleUser {
		t.Errorf("user.Role = %q, want %q", user.Role, domain.RoleUser)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("Register() returned empty tokens")
	}

	claims, err := security.ParseAccessToken("test-secret", tokens.AccessToken)
	if err != nil {
		t.Fatalf("issued access token does not parse: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("access token UserID = %v, want %v", claims.UserID, user.ID)
	}
}

func TestService_Register_DuplicateEmailReturnsConflict(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	in := domain.RegisterInput{Email: "dup@example.com", Password: "password123"}
	if _, _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("first Register() unexpected error: %v", err)
	}

	_, _, err := svc.Register(context.Background(), in)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("second Register() error = %v, want apperr.ErrConflict", err)
	}
}

func TestService_Register_InvalidInputRejectedBeforeRepo(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	_, _, err := svc.Register(context.Background(), domain.RegisterInput{Email: "bad", Password: "short"})
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Register() error = %v, want apperr.ErrInvalid", err)
	}
	if len(users.byEmail) != 0 {
		t.Error("Register() with invalid input must not touch the repo")
	}
}

func TestService_Login_CorrectPasswordSucceeds(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	_, _, err := svc.Register(context.Background(), domain.RegisterInput{Email: "login@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	user, tokens, err := svc.Login(context.Background(), "login@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if user.Email != "login@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "login@example.com")
	}
	if tokens.AccessToken == "" {
		t.Fatal("Login() returned empty access token")
	}
}

func TestService_Login_WrongPasswordReturnsUnauthorized(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	if _, _, err := svc.Register(context.Background(), domain.RegisterInput{Email: "u@example.com", Password: "password123"}); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, _, err := svc.Login(context.Background(), "u@example.com", "wrong-password")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("Login() error = %v, want apperr.ErrUnauthorized", err)
	}
}

// TestService_Login_UnknownEmailAndWrongPassword_SameErrorMessage — anti-enumeration
// (SRS Веха 2, E2.4): по тексту ошибки нельзя понять, существует ли аккаунт.
func TestService_Login_UnknownEmailAndWrongPassword_SameErrorMessage(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	if _, _, err := svc.Register(context.Background(), domain.RegisterInput{Email: "known@example.com", Password: "password123"}); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, _, errUnknownEmail := svc.Login(context.Background(), "unknown@example.com", "password123")
	_, _, errWrongPassword := svc.Login(context.Background(), "known@example.com", "wrong-password")

	if errUnknownEmail == nil || errWrongPassword == nil {
		t.Fatal("expected both Login() calls to fail")
	}
	var e1, e2 *apperr.Error
	if !errors.As(errUnknownEmail, &e1) || !errors.As(errWrongPassword, &e2) {
		t.Fatal("expected *apperr.Error for both")
	}
	if e1.Message() != e2.Message() {
		t.Errorf("error messages differ: %q vs %q — leaks whether email exists", e1.Message(), e2.Message())
	}
}

func TestService_Refresh_RotatesTokenAndRevokesOld(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	clk := clock.NewFake(time.Now())
	svc := newTestService(users, tokens, clk)

	_, pair, err := svc.Register(context.Background(), domain.RegisterInput{Email: "r@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	newPair, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() unexpected error: %v", err)
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Error("Refresh() returned the same refresh token — must rotate")
	}
	if newPair.AccessToken == "" {
		t.Fatal("Refresh() returned empty access token")
	}

	// Старый токен отозван — повторное использование должно провалиться.
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken); err == nil {
		t.Fatal("Refresh() with an already-rotated token expected to fail, got nil error")
	}
}

func TestService_Refresh_ExpiredTokenFails(t *testing.T) {
	users := newFakeUserRepo()
	tokens := newFakeRefreshTokenRepo()
	clk := clock.NewFake(time.Now())
	svc := newTestService(users, tokens, clk)

	_, pair, err := svc.Register(context.Background(), domain.RegisterInput{Email: "exp@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	clk.Advance(31 * 24 * time.Hour) // за пределами RefreshTTL=30 дней

	if _, err := svc.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("Refresh() error = %v, want apperr.ErrUnauthorized", err)
	}
}

func TestService_Me_ReturnsUserByID(t *testing.T) {
	users := newFakeUserRepo()
	svc := newTestService(users, newFakeRefreshTokenRepo(), clock.New())

	created, _, err := svc.Register(context.Background(), domain.RegisterInput{Email: "me@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	got, err := svc.Me(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Me() unexpected error: %v", err)
	}
	if got.Email != created.Email {
		t.Errorf("Me().Email = %q, want %q", got.Email, created.Email)
	}
}

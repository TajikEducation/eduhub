package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// fakeUserRepo — тестовый двойник usecase.UserRepo, in-memory по email.
type fakeUserRepo struct {
	byEmail map[string]domain.User
	byID    map[uuid.UUID]domain.User

	createCalls []domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]domain.User{}, byID: map[uuid.UUID]domain.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u domain.User) (domain.User, error) {
	if _, exists := f.byEmail[u.Email]; exists {
		return domain.User{}, apperr.ConflictCode("email_taken", "email уже зарегистрирован")
	}
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	f.createCalls = append(f.createCalls, u)
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, apperr.NotFound("user", email)
	}
	return u, nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return domain.User{}, apperr.NotFound("user", id.String())
	}
	return u, nil
}

func (f *fakeUserRepo) put(u domain.User) domain.User {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return u
}

func (f *fakeUserRepo) UpdateConsent(_ context.Context, userID uuid.UUID, consentVersion string, consentAt time.Time) error {
	u, ok := f.byID[userID]
	if !ok {
		return apperr.NotFound("user", userID.String())
	}
	u.ConsentVersion = consentVersion
	u.ConsentAt = consentAt
	f.byID[userID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) MarkEmailVerified(_ context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	u, ok := f.byID[userID]
	if !ok {
		return apperr.NotFound("user", userID.String())
	}
	u.EmailVerifiedAt = &verifiedAt
	u.Status = "active"
	f.byID[userID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) UpdatePasswordHash(_ context.Context, userID uuid.UUID, passwordHash string) error {
	u, ok := f.byID[userID]
	if !ok {
		return apperr.NotFound("user", userID.String())
	}
	u.PasswordHash = &passwordHash
	f.byID[userID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) SoftDelete(_ context.Context, userID uuid.UUID, anonymizedEmail string, deletedAt time.Time) error {
	u, ok := f.byID[userID]
	if !ok {
		return apperr.NotFound("user", userID.String())
	}
	delete(f.byEmail, u.Email)
	u.Status = "deleted"
	u.DeletedAt = &deletedAt
	u.Email = anonymizedEmail
	u.Phone = nil
	u.PasswordHash = nil
	f.byID[userID] = u
	f.byEmail[u.Email] = u
	return nil
}

// newTestAccountService собирает AccountService с фейковыми зависимостями поверх общих
// fakeRefreshTokenRepo/fakeUserRoleLookup из session_test.go — не дублирует их. Google-порты
// подставлены нулевыми фейками (не используются тестами, не завязанными на LoginWithGoogle) —
// см. newTestAccountServiceWithGoogle в account_google_test.go для тестов Google-логина.
func newTestAccountService(users *fakeUserRepo, clk *clock.Fake) *usecase.AccountService {
	return newTestAccountServiceWithGoogle(users, clk, &fakeGoogleVerifier{}, newFakeOAuthRepo())
}

func TestRegister_Success_CreatesUnverifiedUserWithNormalizedEmail(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	_, err := svc.Register(context.Background(), "  Test@Example.COM  ", "password123", "v1")
	if err != nil {
		t.Fatalf("Register() вернул ошибку: %v", err)
	}

	if len(users.createCalls) != 1 {
		t.Fatalf("Create() вызван %d раз, want 1", len(users.createCalls))
	}
	created := users.createCalls[0]

	if created.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", created.Email, "test@example.com")
	}
	if created.PasswordHash == nil || *created.PasswordHash == "" {
		t.Error("PasswordHash пуст")
	}
	if created.Role != "user" {
		t.Errorf("Role = %q, want %q", created.Role, "user")
	}
	if created.Status != "unverified" {
		t.Errorf("Status = %q, want %q", created.Status, "unverified")
	}
	if !created.ConsentAt.Equal(start) {
		t.Errorf("ConsentAt = %v, want %v", created.ConsentAt, start)
	}
	if created.ConsentVersion != "v1" {
		t.Errorf("ConsentVersion = %q, want %q", created.ConsentVersion, "v1")
	}
}

func TestRegister_DuplicateEmail_PropagatesConflict(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	if _, err := svc.Register(context.Background(), "dup@example.com", "password123", "v1"); err != nil {
		t.Fatalf("первый Register() вернул ошибку: %v", err)
	}

	_, err := svc.Register(context.Background(), "dup@example.com", "password123", "v1")
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("errors.Is(err, apperr.ErrConflict) = false, err = %v", err)
	}
}

func TestLogin_UnknownEmail_ReturnsUnauthorized(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	_, _, err := svc.Login(context.Background(), "unknown@example.com", "password123")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, apperr.ErrUnauthorized) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Message() != "неверный email или пароль" {
		t.Errorf("Message() = %q, want %q", target.Message(), "неверный email или пароль")
	}
}

func TestLogin_WrongPassword_ReturnsSameUnauthorizedAsUnknownEmail(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	hasher := password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}
	users.put(domain.User{Email: "known@example.com", PasswordHash: &hash, Role: "user", Status: "active"})

	_, _, unknownErr := svc.Login(context.Background(), "unknown@example.com", "whatever")
	_, _, wrongErr := svc.Login(context.Background(), "known@example.com", "wrong-password")

	if !errors.Is(wrongErr, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(wrongErr, apperr.ErrUnauthorized) = false, err = %v", wrongErr)
	}

	var unknownTarget, wrongTarget *apperr.Error
	if !errors.As(unknownErr, &unknownTarget) || !errors.As(wrongErr, &wrongTarget) {
		t.Fatalf("errors.As не сработал: unknownErr=%v, wrongErr=%v", unknownErr, wrongErr)
	}
	if unknownTarget.Message() != wrongTarget.Message() {
		t.Errorf("сообщения различаются: unknown=%q, wrong=%q (анти-энумерация нарушена)", unknownTarget.Message(), wrongTarget.Message())
	}
}

func TestLogin_GoogleOnlyAccount_ReturnsGoogleAccountNoPasswordCode(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	users.put(domain.User{Email: "google@example.com", PasswordHash: nil, Role: "user", Status: "active"})

	_, _, err := svc.Login(context.Background(), "google@example.com", "irrelevant")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, apperr.ErrUnauthorized) = false, err = %v", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", err)
	}
	if target.Code() != "google_account_no_password" {
		t.Errorf("Code() = %q, want %q", target.Code(), "google_account_no_password")
	}
}

func TestLogin_BannedAccountWithCorrectPassword_StillReturnsUnauthorized(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	hasher := password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}
	users.put(domain.User{Email: "banned@example.com", PasswordHash: &hash, Role: "user", Status: "banned"})

	_, _, loginErr := svc.Login(context.Background(), "banned@example.com", "correct-password")
	if !errors.Is(loginErr, apperr.ErrUnauthorized) {
		t.Fatalf("errors.Is(loginErr, apperr.ErrUnauthorized) = false, err = %v", loginErr)
	}

	var target *apperr.Error
	if !errors.As(loginErr, &target) {
		t.Fatalf("errors.As(err, &target) = false, err = %v", loginErr)
	}
	if target.Code() != "" {
		t.Errorf("Code() = %q, want пустой (обычный unauthorized, не отдельный код на статус аккаунта)", target.Code())
	}
}

func TestLogin_Success_ReturnsNonEmptyTokens(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	svc := newTestAccountService(users, clk)

	hasher := password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}
	users.put(domain.User{Email: "active@example.com", PasswordHash: &hash, Role: "user", Status: "active"})

	access, refresh, err := svc.Login(context.Background(), "active@example.com", "correct-password")
	if err != nil {
		t.Fatalf("Login() вернул ошибку: %v", err)
	}
	if access == "" {
		t.Error("access-токен пуст")
	}
	if refresh == "" {
		t.Error("refresh-токен пуст")
	}
}

// newTestAccountServiceWithSessions — тот же набор фейков, что newTestAccountService, но также
// возвращает fakeRefreshTokenRepo — нужен тестам DeleteMe, чтобы проверить, что все сессии
// пользователя реально отозваны.
func newTestAccountServiceWithSessions(users *fakeUserRepo, clk *clock.Fake) (*usecase.AccountService, *fakeRefreshTokenRepo) {
	sessionRepo := newFakeRefreshTokenRepo()
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	sessions := usecase.NewSessionService(sessionRepo, fakeUserRoleLookup{role: "user"}, issuer, clk, refreshTTL)
	hasher := password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	svc := usecase.NewAccountService(users, hasher, sessions, clk, &fakeGoogleVerifier{}, newFakeOAuthRepo())
	return svc, sessionRepo
}

func TestUpdateConsent_UpdatesVersionAndTimestamp(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "consent@example.com", Role: "user", Status: "active", ConsentVersion: "v1"})
	svc := newTestAccountService(users, clk)

	if err := svc.UpdateConsent(context.Background(), u.ID, "v2"); err != nil {
		t.Fatalf("UpdateConsent() вернул ошибку: %v", err)
	}

	updated, err := users.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", err)
	}
	if updated.ConsentVersion != "v2" {
		t.Errorf("ConsentVersion = %q, want %q", updated.ConsentVersion, "v2")
	}
	if !updated.ConsentAt.Equal(start) {
		t.Errorf("ConsentAt = %v, want %v", updated.ConsentAt, start)
	}
}

func TestDeleteMe_SoftDeletesAnonymizesAndRevokesAllSessions(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "delete-me@example.com", Role: "user", Status: "active"})
	svc, sessionRepo := newTestAccountServiceWithSessions(users, clk)

	if err := svc.DeleteMe(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteMe() вернул ошибку: %v", err)
	}

	updated, err := users.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", err)
	}
	if updated.Status != "deleted" {
		t.Errorf("Status = %q, want %q", updated.Status, "deleted")
	}
	wantEmail := "deleted-" + u.ID.String() + "@eduhub.local"
	if updated.Email != wantEmail {
		t.Errorf("Email = %q, want %q", updated.Email, wantEmail)
	}
	if len(sessionRepo.revokeAllForUserCalls) != 1 || sessionRepo.revokeAllForUserCalls[0] != u.ID {
		t.Errorf("revokeAllForUserCalls = %v, want [%v]", sessionRepo.revokeAllForUserCalls, u.ID)
	}
}

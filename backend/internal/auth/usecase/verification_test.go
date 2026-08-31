package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
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

func testHasher() *password.Hasher {
	return password.New(password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// fakeVerificationCodeRepo — тестовый двойник usecase.VerificationCodeRepo, in-memory по id.
type fakeVerificationCodeRepo struct {
	byID map[uuid.UUID]domain.VerificationCode

	createCalls            []domain.VerificationCode
	incrementAttemptsCalls []uuid.UUID
	deleteCalls            []uuid.UUID
}

func newFakeVerificationCodeRepo() *fakeVerificationCodeRepo {
	return &fakeVerificationCodeRepo{byID: map[uuid.UUID]domain.VerificationCode{}}
}

func (f *fakeVerificationCodeRepo) Create(_ context.Context, vc domain.VerificationCode) error {
	f.createCalls = append(f.createCalls, vc)
	f.byID[vc.ID] = vc
	return nil
}

func (f *fakeVerificationCodeRepo) FindLatestActive(_ context.Context, userID uuid.UUID, channel, purpose string, now time.Time) (domain.VerificationCode, error) {
	var latest domain.VerificationCode
	found := false
	for _, vc := range f.byID {
		if vc.UserID != userID || vc.Channel != channel || vc.Purpose != purpose {
			continue
		}
		if !vc.ExpiresAt.After(now) {
			continue
		}
		if !found || vc.CreatedAt.After(latest.CreatedAt) {
			latest = vc
			found = true
		}
	}
	if !found {
		return domain.VerificationCode{}, apperr.NotFound("verification_code", userID.String())
	}
	return latest, nil
}

func (f *fakeVerificationCodeRepo) IncrementAttempts(_ context.Context, id uuid.UUID) error {
	f.incrementAttemptsCalls = append(f.incrementAttemptsCalls, id)
	vc, ok := f.byID[id]
	if !ok {
		return apperr.NotFound("verification_code", id.String())
	}
	vc.AttemptsCount++
	f.byID[id] = vc
	return nil
}

func (f *fakeVerificationCodeRepo) Delete(_ context.Context, id uuid.UUID) error {
	f.deleteCalls = append(f.deleteCalls, id)
	delete(f.byID, id)
	return nil
}

func newTestVerificationService(users *fakeUserRepo, codes *fakeVerificationCodeRepo, clk *clock.Fake) (*usecase.VerificationService, *fakeRefreshTokenRepo) {
	sessionRepo := newFakeRefreshTokenRepo()
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	sessions := usecase.NewSessionService(sessionRepo, fakeUserRoleLookup{role: "user"}, issuer, clk, refreshTTL)
	svc := usecase.NewVerificationService(users, codes, sessions, testHasher(), clk, noopLogger())
	return svc, sessionRepo
}

func TestRequestEmailVerification_UnknownEmail_SilentSuccessWithoutCreate(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestEmailVerification(context.Background(), "unknown@example.com"); err != nil {
		t.Fatalf("RequestEmailVerification() вернул ошибку: %v", err)
	}
	if len(codes.createCalls) != 0 {
		t.Errorf("codes.Create() вызван %d раз, want 0", len(codes.createCalls))
	}
}

func TestRequestEmailVerification_AlreadyVerified_SilentSuccessWithoutCreate(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	verifiedAt := time.Now()
	users.put(domain.User{Email: "verified@example.com", Role: "user", Status: "active", EmailVerifiedAt: &verifiedAt})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestEmailVerification(context.Background(), "verified@example.com"); err != nil {
		t.Fatalf("RequestEmailVerification() вернул ошибку: %v", err)
	}
	if len(codes.createCalls) != 0 {
		t.Errorf("codes.Create() вызван %d раз, want 0", len(codes.createCalls))
	}
}

func TestRequestEmailVerification_Unverified_CreatesCodeWithRegisterPurpose(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	users.put(domain.User{Email: "unverified@example.com", Role: "user", Status: "unverified"})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestEmailVerification(context.Background(), "unverified@example.com"); err != nil {
		t.Fatalf("RequestEmailVerification() вернул ошибку: %v", err)
	}
	if len(codes.createCalls) != 1 {
		t.Fatalf("codes.Create() вызван %d раз, want 1", len(codes.createCalls))
	}
	if codes.createCalls[0].Purpose != "register" {
		t.Errorf("Purpose = %q, want %q", codes.createCalls[0].Purpose, "register")
	}
}

func TestVerifyEmail_WrongCode_ReturnsInvalidAndIncrementsAttempts(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "unverified"})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestEmailVerification(context.Background(), u.Email); err != nil {
		t.Fatalf("RequestEmailVerification() вернул ошибку: %v", err)
	}
	var codeID uuid.UUID
	for id := range codes.byID {
		codeID = id
	}

	err := svc.VerifyEmail(context.Background(), u.Email, "000000")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
	}
	if len(codes.incrementAttemptsCalls) != 1 || codes.incrementAttemptsCalls[0] != codeID {
		t.Errorf("incrementAttemptsCalls = %v, want [%v]", codes.incrementAttemptsCalls, codeID)
	}
}

func TestVerifyEmail_NoActiveCode_ReturnsInvalid(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "unverified"})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	err := svc.VerifyEmail(context.Background(), u.Email, "123456")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
	}
}

func TestVerifyEmail_TooManyAttempts_ReturnsInvalidAndDeletesCode(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "unverified"})
	codes := newFakeVerificationCodeRepo()

	vc := domain.VerificationCode{
		ID: uuid.New(), UserID: u.ID, Channel: "email", Purpose: "register",
		CodeHash: "irrelevant", AttemptsCount: 5, ExpiresAt: clk.Now().Add(time.Hour), CreatedAt: clk.Now(),
	}
	codes.byID[vc.ID] = vc

	svc, _ := newTestVerificationService(users, codes, clk)

	err := svc.VerifyEmail(context.Background(), u.Email, "123456")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
	}
	if len(codes.deleteCalls) != 1 || codes.deleteCalls[0] != vc.ID {
		t.Errorf("deleteCalls = %v, want [%v]", codes.deleteCalls, vc.ID)
	}
}

func TestVerifyEmail_CorrectCode_MarksVerifiedAndDeletesCode(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "unverified"})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestEmailVerification(context.Background(), u.Email); err != nil {
		t.Fatalf("RequestEmailVerification() вернул ошибку: %v", err)
	}
	var stored domain.VerificationCode
	for _, vc := range codes.byID {
		stored = vc
	}

	// Достаём "настоящий" код напрямую хешем нельзя — генерация случайна. Подменяем хеш на
	// известный код, чтобы детерминированно проверить happy path.
	stored.CodeHash = hashCodeForTest("111111")
	codes.byID[stored.ID] = stored

	if err := svc.VerifyEmail(context.Background(), u.Email, "111111"); err != nil {
		t.Fatalf("VerifyEmail() вернул ошибку: %v", err)
	}

	updated, err := users.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", err)
	}
	if updated.EmailVerifiedAt == nil {
		t.Error("EmailVerifiedAt не установлен после VerifyEmail()")
	}
	if updated.Status != "active" {
		t.Errorf("Status = %q, want %q", updated.Status, "active")
	}
	if len(codes.deleteCalls) != 1 || codes.deleteCalls[0] != stored.ID {
		t.Errorf("deleteCalls = %v, want [%v]", codes.deleteCalls, stored.ID)
	}
}

func TestRequestPasswordReset_UnknownEmail_SilentSuccessWithoutCreate(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestPasswordReset(context.Background(), "unknown@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset() вернул ошибку: %v", err)
	}
	if len(codes.createCalls) != 0 {
		t.Errorf("codes.Create() вызван %d раз, want 0", len(codes.createCalls))
	}
}

func TestConfirmPasswordReset_CorrectCode_UpdatesPasswordAndRevokesSessions(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "active"})
	codes := newFakeVerificationCodeRepo()
	svc, sessionRepo := newTestVerificationService(users, codes, clk)

	if err := svc.RequestPasswordReset(context.Background(), u.Email); err != nil {
		t.Fatalf("RequestPasswordReset() вернул ошибку: %v", err)
	}
	var stored domain.VerificationCode
	for _, vc := range codes.byID {
		stored = vc
	}
	stored.CodeHash = hashCodeForTest("222222")
	codes.byID[stored.ID] = stored

	if err := svc.ConfirmPasswordReset(context.Background(), u.Email, "222222", "new-password-123"); err != nil {
		t.Fatalf("ConfirmPasswordReset() вернул ошибку: %v", err)
	}

	updated, err := users.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", err)
	}
	if updated.PasswordHash == nil {
		t.Fatal("PasswordHash не установлен после ConfirmPasswordReset()")
	}
	if len(codes.deleteCalls) != 1 || codes.deleteCalls[0] != stored.ID {
		t.Errorf("deleteCalls = %v, want [%v]", codes.deleteCalls, stored.ID)
	}
	if len(sessionRepo.revokeAllForUserCalls) != 1 || sessionRepo.revokeAllForUserCalls[0] != u.ID {
		t.Errorf("revokeAllForUserCalls = %v, want [%v]", sessionRepo.revokeAllForUserCalls, u.ID)
	}
}

func TestConfirmPasswordReset_WrongCode_PasswordNotChanged(t *testing.T) {
	clk := clock.NewFake(time.Now())
	users := newFakeUserRepo()
	hasher := testHasher()
	originalHash, err := hasher.Hash("original-password")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}
	u := users.put(domain.User{Email: "test@example.com", Role: "user", Status: "active", PasswordHash: &originalHash})
	codes := newFakeVerificationCodeRepo()
	svc, _ := newTestVerificationService(users, codes, clk)

	if err := svc.RequestPasswordReset(context.Background(), u.Email); err != nil {
		t.Fatalf("RequestPasswordReset() вернул ошибку: %v", err)
	}

	err = svc.ConfirmPasswordReset(context.Background(), u.Email, "000000", "new-password-123")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("errors.Is(err, apperr.ErrInvalid) = false, err = %v", err)
	}

	updated, findErr := users.FindByID(context.Background(), u.ID)
	if findErr != nil {
		t.Fatalf("FindByID() вернул ошибку: %v", findErr)
	}
	if updated.PasswordHash == nil || *updated.PasswordHash != originalHash {
		t.Error("PasswordHash изменился после неверного кода — не должен")
	}
}

// hashCodeForTest дублирует internal-хеширование VerificationService.checkCode (обычный
// sha256, см. usecase.hashCode) — нужен, чтобы тесты могли детерминированно подменить CodeHash
// на известное значение, минуя недоступный извне сгенерированный код. Не тест-хук в
// production-коде — сам алгоритм тривиален и стабилен (sha256 hex), дублирование безопасно.
func hashCodeForTest(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

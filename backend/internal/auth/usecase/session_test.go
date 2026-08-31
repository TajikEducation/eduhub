package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

const refreshTTL = 30 * 24 * time.Hour

// fakeRefreshTokenRepo — тестовый двойник usecase.RefreshTokenRepo, in-memory по TokenHash.
type fakeRefreshTokenRepo struct {
	byHash map[string]domain.RefreshToken

	createCalls           []domain.RefreshToken
	revokeCalls           []revokeCall
	revokeFamilyCalls     []uuid.UUID
	revokeAllForUserCalls []uuid.UUID
}

type revokeCall struct {
	id         uuid.UUID
	revokedAt  time.Time
	replacedBy *uuid.UUID
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{byHash: map[string]domain.RefreshToken{}}
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, rt domain.RefreshToken) error {
	f.byHash[rt.TokenHash] = rt
	f.createCalls = append(f.createCalls, rt)
	return nil
}

func (f *fakeRefreshTokenRepo) FindByHash(_ context.Context, tokenHash string) (domain.RefreshToken, error) {
	rt, ok := f.byHash[tokenHash]
	if !ok {
		// apperr.NotFound — тот же контракт, что вернёт реальный internal/auth/repo/postgres.
		return domain.RefreshToken{}, apperr.NotFound("refresh_token", tokenHash)
	}
	return rt, nil
}

func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id uuid.UUID, revokedAt time.Time, replacedBy *uuid.UUID) error {
	f.revokeCalls = append(f.revokeCalls, revokeCall{id: id, revokedAt: revokedAt, replacedBy: replacedBy})
	for hash, rt := range f.byHash {
		if rt.ID == id {
			rt.RevokedAt = &revokedAt
			rt.ReplacedBy = replacedBy
			f.byHash[hash] = rt
		}
	}
	return nil
}

func (f *fakeRefreshTokenRepo) RevokeFamily(_ context.Context, familyID uuid.UUID, revokedAt time.Time) error {
	f.revokeFamilyCalls = append(f.revokeFamilyCalls, familyID)
	for hash, rt := range f.byHash {
		if rt.FamilyID == familyID && rt.RevokedAt == nil {
			rt.RevokedAt = &revokedAt
			f.byHash[hash] = rt
		}
	}
	return nil
}

func (f *fakeRefreshTokenRepo) RevokeAllForUser(_ context.Context, userID uuid.UUID, revokedAt time.Time) error {
	f.revokeAllForUserCalls = append(f.revokeAllForUserCalls, userID)
	for hash, rt := range f.byHash {
		if rt.UserID == userID && rt.RevokedAt == nil {
			rt.RevokedAt = &revokedAt
			f.byHash[hash] = rt
		}
	}
	return nil
}

func hashOf(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// fakeUserRoleLookup — тестовый двойник usecase.UserRoleLookup, фиксированная роль на всех.
type fakeUserRoleLookup struct{ role string }

func (f fakeUserRoleLookup) RoleByUserID(context.Context, uuid.UUID) (string, error) {
	return f.role, nil
}

func newTestService(repo usecase.RefreshTokenRepo, clk *clock.Fake) *usecase.SessionService {
	issuer := jwt.NewIssuer([]byte("test-secret"), 15*time.Minute, clk)
	return usecase.NewSessionService(repo, fakeUserRoleLookup{role: "user"}, issuer, clk, refreshTTL)
}

func TestIssue_CreatesRefreshTokenRowAndReturnsValidAccessToken(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)
	userID := uuid.New()

	access, refresh, err := svc.Issue(context.Background(), userID, "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("Issue() вернул пустой access или refresh токен")
	}

	if len(repo.createCalls) != 1 {
		t.Fatalf("Create() вызван %d раз, want 1", len(repo.createCalls))
	}
	created := repo.createCalls[0]

	if created.UserID != userID {
		t.Errorf("created.UserID = %v, want %v", created.UserID, userID)
	}
	if created.TokenHash != hashOf(refresh) {
		t.Error("created.TokenHash не соответствует sha256(refresh) — хранится не хеш возвращённого токена")
	}
	if created.FamilyID == uuid.Nil {
		t.Error("created.FamilyID пуст — новая сессия должна получить свежий family_id")
	}
	if !created.ExpiresAt.Equal(start.Add(refreshTTL)) {
		t.Errorf("created.ExpiresAt = %v, want %v", created.ExpiresAt, start.Add(refreshTTL))
	}
	if created.RevokedAt != nil {
		t.Error("created.RevokedAt должен быть nil для свежего токена")
	}
}

func TestRotate_ValidToken_IssuesNewTokenInSameFamily(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)
	userID := uuid.New()

	_, refresh1, err := svc.Issue(context.Background(), userID, "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}
	firstFamily := repo.createCalls[0].FamilyID
	firstID := repo.createCalls[0].ID

	clk.Advance(time.Hour)

	access2, refresh2, err := svc.Rotate(context.Background(), refresh1)
	if err != nil {
		t.Fatalf("Rotate() вернул ошибку: %v", err)
	}
	if access2 == "" || refresh2 == "" {
		t.Fatal("Rotate() вернул пустой access или refresh")
	}
	if refresh2 == refresh1 {
		t.Error("Rotate() вернул тот же refresh-токен — ротации не произошло")
	}

	if len(repo.createCalls) != 2 {
		t.Fatalf("Create() вызван %d раз, want 2 (issue + rotate)", len(repo.createCalls))
	}
	second := repo.createCalls[1]
	if second.FamilyID != firstFamily {
		t.Errorf("second.FamilyID = %v, want %v (та же семья)", second.FamilyID, firstFamily)
	}

	if len(repo.revokeCalls) != 1 {
		t.Fatalf("Revoke() вызван %d раз, want 1", len(repo.revokeCalls))
	}
	revoked := repo.revokeCalls[0]
	if revoked.id != firstID {
		t.Errorf("Revoke() id = %v, want %v (старый токен)", revoked.id, firstID)
	}
	if revoked.replacedBy == nil || *revoked.replacedBy != second.ID {
		t.Error("Revoke() replacedBy не указывает на новый токен")
	}

	if len(repo.revokeFamilyCalls) != 0 {
		t.Error("RevokeFamily() вызван при штатной ротации — не должен")
	}
}

func TestRotate_UnknownToken_ReturnsNotFoundWithoutSideEffects(t *testing.T) {
	clk := clock.NewFake(time.Now())
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	_, _, err := svc.Rotate(context.Background(), "неизвестный-токен")
	if !errors.Is(err, usecase.ErrRefreshTokenNotFound) {
		t.Errorf("Rotate() ошибка = %v, want ErrRefreshTokenNotFound", err)
	}
	if len(repo.createCalls) != 0 || len(repo.revokeCalls) != 0 || len(repo.revokeFamilyCalls) != 0 {
		t.Error("Rotate() неизвестного токена не должен иметь побочных эффектов")
	}
}

func TestRotate_ExpiredToken_ReturnsExpiredWithoutIssuingNew(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	_, refresh, err := svc.Issue(context.Background(), uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	clk.Advance(refreshTTL + time.Hour)

	_, _, err = svc.Rotate(context.Background(), refresh)
	if !errors.Is(err, usecase.ErrRefreshTokenExpired) {
		t.Errorf("Rotate() ошибка = %v, want ErrRefreshTokenExpired", err)
	}
	if len(repo.createCalls) != 1 {
		t.Errorf("Create() вызван %d раз после Issue+Rotate истёкшего, want 1 (только Issue)", len(repo.createCalls))
	}
	if len(repo.revokeFamilyCalls) != 0 {
		t.Error("RevokeFamily() вызван для естественно истёкшего токена — не должен, это не reuse")
	}
}

func TestRotate_ReusedToken_RevokesEntireFamily(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	_, refresh1, err := svc.Issue(context.Background(), uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}
	firstFamily := repo.createCalls[0].FamilyID

	// Легитимная ротация — refresh1 теперь отозван (replaced_by второй токен).
	_, refresh2, err := svc.Rotate(context.Background(), refresh1)
	if err != nil {
		t.Fatalf("Rotate() (легитимная) вернул ошибку: %v", err)
	}

	// Атакующий (или клиент с гонкой) предъявляет УЖЕ использованный refresh1 повторно.
	_, _, err = svc.Rotate(context.Background(), refresh1)
	if !errors.Is(err, usecase.ErrRefreshTokenReused) {
		t.Fatalf("повторный Rotate() старого токена = %v, want ErrRefreshTokenReused", err)
	}

	if len(repo.revokeFamilyCalls) != 1 || repo.revokeFamilyCalls[0] != firstFamily {
		t.Fatalf("RevokeFamily() calls = %v, want [%v]", repo.revokeFamilyCalls, firstFamily)
	}

	// Легитимный refresh2 (реально свежий, законно рождённый в этой же семье) тоже должен
	// стать бесполезен — вся семья скомпрометирована после обнаруженного reuse.
	rt2, ferr := repo.FindByHash(context.Background(), hashOf(refresh2))
	if ferr != nil {
		t.Fatalf("FindByHash(refresh2) вернул ошибку: %v", ferr)
	}
	if rt2.RevokedAt == nil {
		t.Error("refresh2 (та же семья) не отозван после reuse-detection — вся семья должна быть недействительна")
	}
}

func TestLogout_ValidToken_RevokesWithoutReplacement(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	_, refresh, err := svc.Issue(context.Background(), uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}
	issuedID := repo.createCalls[0].ID

	if err := svc.Logout(context.Background(), refresh); err != nil {
		t.Fatalf("Logout() вернул ошибку: %v", err)
	}

	if len(repo.revokeCalls) != 1 {
		t.Fatalf("Revoke() вызван %d раз, want 1", len(repo.revokeCalls))
	}
	revoked := repo.revokeCalls[0]
	if revoked.id != issuedID {
		t.Errorf("Revoke() id = %v, want %v", revoked.id, issuedID)
	}
	if revoked.replacedBy != nil {
		t.Errorf("Revoke() replacedBy = %v, want nil (logout — не ротация)", revoked.replacedBy)
	}
}

func TestLogout_UnknownToken_SilentSuccessWithoutRevoke(t *testing.T) {
	clk := clock.NewFake(time.Now())
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	if err := svc.Logout(context.Background(), "неизвестный-токен"); err != nil {
		t.Fatalf("Logout() вернул ошибку: %v, want nil", err)
	}
	if len(repo.revokeCalls) != 0 {
		t.Errorf("Revoke() вызван %d раз, want 0", len(repo.revokeCalls))
	}
}

func TestLogout_AlreadyRevokedToken_SilentSuccessWithoutRevoke(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)

	_, refresh, err := svc.Issue(context.Background(), uuid.New(), "user")
	if err != nil {
		t.Fatalf("Issue() вернул ошибку: %v", err)
	}

	if err := svc.Logout(context.Background(), refresh); err != nil {
		t.Fatalf("первый Logout() вернул ошибку: %v", err)
	}
	if len(repo.revokeCalls) != 1 {
		t.Fatalf("Revoke() после первого Logout() вызван %d раз, want 1", len(repo.revokeCalls))
	}

	if err := svc.Logout(context.Background(), refresh); err != nil {
		t.Fatalf("повторный Logout() вернул ошибку: %v, want nil", err)
	}
	if len(repo.revokeCalls) != 1 {
		t.Errorf("Revoke() после повторного Logout() вызван %d раз, want 1 (не должен вызываться снова)", len(repo.revokeCalls))
	}
}

func TestRevokeAllForUser_CallsRepoWithUserIDAndCurrentClock(t *testing.T) {
	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFake(start)
	repo := newFakeRefreshTokenRepo()
	svc := newTestService(repo, clk)
	userID := uuid.New()

	if err := svc.RevokeAllForUser(context.Background(), userID); err != nil {
		t.Fatalf("RevokeAllForUser() вернул ошибку: %v", err)
	}

	if len(repo.revokeAllForUserCalls) != 1 || repo.revokeAllForUserCalls[0] != userID {
		t.Fatalf("repo.RevokeAllForUser() calls = %v, want [%v]", repo.revokeAllForUserCalls, userID)
	}
}

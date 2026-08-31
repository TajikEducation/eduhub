package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

var (
	// ErrRefreshTokenNotFound — предъявленный refresh-токен неизвестен репозиторию.
	ErrRefreshTokenNotFound = errors.New("usecase: refresh token not found")
	// ErrRefreshTokenExpired — токен найден, не отозван, но истёк естественным образом.
	ErrRefreshTokenExpired = errors.New("usecase: refresh token expired")
	// ErrRefreshTokenReused — токен найден уже отозванным (ротацией или reuse-detection'ом
	// ранее) — предъявление означает компрометацию, вся family_id отзывается.
	ErrRefreshTokenReused = errors.New("usecase: refresh token reuse detected")
)

// SessionService выпускает access/refresh токены и ротирует refresh с reuse-detection (E2.3).
type SessionService struct {
	repo       RefreshTokenRepo
	users      UserRoleLookup
	jwtIssuer  *jwt.Issuer
	clock      clock.Clock
	refreshTTL time.Duration
}

// NewSessionService создаёт SessionService. clk — те же инжектируемые часы, что и у jwt.Issuer
// (обычно один и тот же clock.Clock на оба).
func NewSessionService(repo RefreshTokenRepo, users UserRoleLookup, jwtIssuer *jwt.Issuer, clk clock.Clock, refreshTTL time.Duration) *SessionService {
	return &SessionService{repo: repo, users: users, jwtIssuer: jwtIssuer, clock: clk, refreshTTL: refreshTTL}
}

// Issue выпускает новую сессию (новый family_id) для пользователя — вызывается при логине.
// role передаётся вызывающим (логин уже прочитал строку пользователя для проверки пароля) —
// повторный поход в БД здесь был бы лишним.
func (s *SessionService) Issue(ctx context.Context, userID uuid.UUID, role string) (accessToken, refreshToken string, err error) {
	accessToken, refreshToken, _, err = s.issueTokenPair(ctx, userID, uuid.New(), role)
	return accessToken, refreshToken, err
}

// Rotate предъявляет refresh-токен и выпускает новую пару access+refresh в той же семье.
// Роль в новом access-токене читается заново через UserRoleLookup — см. её комментарий.
//
// Reuse-detection: если предъявленный токен уже отозван (ротацией или предыдущим
// reuse-detection'ом) — это сигнал компрометации, вся family_id отзывается разом, новые
// токены не выпускаются. Естественно истёкший (но не отозванный) токен — не reuse, обычная
// ошибка истечения сессии.
func (s *SessionService) Rotate(ctx context.Context, presentedRefreshToken string) (accessToken, refreshToken string, err error) {
	hash := hashRefreshToken(presentedRefreshToken)

	existing, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return "", "", ErrRefreshTokenNotFound
		}
		return "", "", fmt.Errorf("usecase: find refresh token: %w", err)
	}

	if existing.RevokedAt != nil {
		now := s.clock.Now()
		if err := s.repo.RevokeFamily(ctx, existing.FamilyID, now); err != nil {
			return "", "", fmt.Errorf("usecase: revoke family after reuse detection: %w", err)
		}
		return "", "", ErrRefreshTokenReused
	}

	if !s.clock.Now().Before(existing.ExpiresAt) {
		return "", "", ErrRefreshTokenExpired
	}

	role, err := s.users.RoleByUserID(ctx, existing.UserID)
	if err != nil {
		return "", "", fmt.Errorf("usecase: lookup current role: %w", err)
	}

	accessToken, refreshToken, newID, err := s.issueTokenPair(ctx, existing.UserID, existing.FamilyID, role)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.Revoke(ctx, existing.ID, s.clock.Now(), &newID); err != nil {
		return "", "", fmt.Errorf("usecase: revoke rotated token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *SessionService) issueTokenPair(ctx context.Context, userID, familyID uuid.UUID, role string) (accessToken, refreshToken string, newID uuid.UUID, err error) {
	plaintext, err := generateRefreshToken()
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("usecase: generate refresh token: %w", err)
	}

	now := s.clock.Now()
	newID = uuid.New()
	rt := domain.RefreshToken{
		ID:        newID,
		UserID:    userID,
		TokenHash: hashRefreshToken(plaintext),
		FamilyID:  familyID,
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}

	if err := s.repo.Create(ctx, rt); err != nil {
		return "", "", uuid.Nil, fmt.Errorf("usecase: create refresh token: %w", err)
	}

	accessToken, err = s.jwtIssuer.Issue(userID, role)
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("usecase: issue access token: %w", err)
	}

	return accessToken, plaintext, newID, nil
}

// generateRefreshToken — криптостойкий случайный refresh-токен (256 бит энтропии).
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashRefreshToken — обычный SHA-256, не argon2id: refresh-токен уже 256 бит энтропии
// (в отличие от человеческого пароля), брутфорс по хешу вычислительно бессмыслен и без
// дорогой KDF — быстрый хеш здесь не ослабляет защиту, а argon2id только замедлил бы
// каждый Rotate() без пользы.
func hashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

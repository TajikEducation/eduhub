package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

const (
	verificationCodeTTL         = 15 * time.Minute
	verificationCodeMaxAttempts = 5
	invalidOrExpiredCodeMessage = "неверный или истёкший код"
)

// VerificationService — email-подтверждение и сброс пароля через короткоживущие коды (E2.4).
type VerificationService struct {
	users    UserRepo
	codes    VerificationCodeRepo
	sessions *SessionService
	hasher   *password.Hasher
	clock    clock.Clock
	logger   *slog.Logger
}

// NewVerificationService создаёт VerificationService.
func NewVerificationService(users UserRepo, codes VerificationCodeRepo, sessions *SessionService, hasher *password.Hasher, clk clock.Clock, logger *slog.Logger) *VerificationService {
	return &VerificationService{users: users, codes: codes, sessions: sessions, hasher: hasher, clock: clk, logger: logger}
}

// RequestEmailVerification генерирует и "отправляет" (пока — логирует, см. комментарий в конце
// функции) код подтверждения email. Тихий успех (nil без ошибки), если email не найден или
// аккаунт уже верифицирован — anti-enumeration, вызывающий transport-хендлер всегда отвечает 200.
func (s *VerificationService) RequestEmailVerification(ctx context.Context, email string) error {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("usecase: request email verification: find user: %w", err)
	}
	if u.EmailVerifiedAt != nil {
		return nil
	}
	return s.issueCode(ctx, u.ID, "register")
}

// VerifyEmail сверяет код, при успехе активирует аккаунт (MarkEmailVerified) и удаляет код.
func (s *VerificationService) VerifyEmail(ctx context.Context, email, code string) error {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return apperr.Invalid(map[string]string{"code": invalidOrExpiredCodeMessage}, invalidOrExpiredCodeMessage)
		}
		return fmt.Errorf("usecase: verify email: find user: %w", err)
	}

	vc, err := s.checkCode(ctx, u.ID, "email", "register", code)
	if err != nil {
		return err
	}

	if err := s.users.MarkEmailVerified(ctx, u.ID, s.clock.Now()); err != nil {
		return fmt.Errorf("usecase: verify email: mark verified: %w", err)
	}
	if err := s.codes.Delete(ctx, vc.ID); err != nil {
		return fmt.Errorf("usecase: verify email: delete used code: %w", err)
	}
	return nil
}

// RequestPasswordReset — та же "тихий успех" семантика, что RequestEmailVerification.
func (s *VerificationService) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("usecase: request password reset: find user: %w", err)
	}
	return s.issueCode(ctx, u.ID, "password_reset")
}

// ConfirmPasswordReset сверяет код, при успехе меняет пароль, удаляет код И отзывает ВСЕ
// активные сессии пользователя (см. RefreshTokenRepo.RevokeAllForUser — смена пароля обязана
// выкидывать из всех устройств, не только текущего).
func (s *VerificationService) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return apperr.Invalid(map[string]string{"code": invalidOrExpiredCodeMessage}, invalidOrExpiredCodeMessage)
		}
		return fmt.Errorf("usecase: confirm password reset: find user: %w", err)
	}

	vc, err := s.checkCode(ctx, u.ID, "email", "password_reset", code)
	if err != nil {
		return err
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("usecase: confirm password reset: hash password: %w", err)
	}
	if err := s.users.UpdatePasswordHash(ctx, u.ID, hash); err != nil {
		return fmt.Errorf("usecase: confirm password reset: update hash: %w", err)
	}
	if err := s.codes.Delete(ctx, vc.ID); err != nil {
		return fmt.Errorf("usecase: confirm password reset: delete used code: %w", err)
	}
	if err := s.sessions.RevokeAllForUser(ctx, u.ID); err != nil {
		return fmt.Errorf("usecase: confirm password reset: revoke sessions: %w", err)
	}
	return nil
}

// issueCode генерирует случайный 6-значный код, сохраняет его хеш, логирует САМ код на
// DEBUG-уровне как временную замену реального email-провайдера (тот же принцип, что план
// описывает для SMS/email-заглушек в вехе 5, E5.3) — НЕ логировать на уровне INFO и выше,
// код — секрет пользователя, DEBUG виден только при явно включённом LOG_LEVEL=debug в dev.
func (s *VerificationService) issueCode(ctx context.Context, userID uuid.UUID, purpose string) error {
	code, err := generateNumericCode()
	if err != nil {
		return fmt.Errorf("usecase: issue verification code: generate: %w", err)
	}

	now := s.clock.Now()
	vc := domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   "email",
		Purpose:   purpose,
		CodeHash:  hashCode(code),
		ExpiresAt: now.Add(verificationCodeTTL),
		CreatedAt: now,
	}
	if err := s.codes.Create(ctx, vc); err != nil {
		return fmt.Errorf("usecase: issue verification code: create: %w", err)
	}

	s.logger.Debug("verification code issued (заглушка вместо реальной отправки email)",
		slog.String("user_id", userID.String()), slog.String("purpose", purpose), slog.String("code", code))
	return nil
}

// checkCode сверяет предъявленный код с последним активным для (userID, channel, purpose).
// Любая причина отказа (нет активного кода, неверный код, превышены попытки) — ОДНО и то же
// apperr.Invalid с invalidOrExpiredCodeMessage — не давать атакующему сигнал, какая именно
// причина. Неверный код инкрементирует attempts_count; при достижении лимита код удаляется
// целиком (форсирует resend, не бесконечный брутфорс одного и того же кода).
func (s *VerificationService) checkCode(ctx context.Context, userID uuid.UUID, channel, purpose, presentedCode string) (domain.VerificationCode, error) {
	invalidErr := apperr.Invalid(map[string]string{"code": invalidOrExpiredCodeMessage}, invalidOrExpiredCodeMessage)

	vc, err := s.codes.FindLatestActive(ctx, userID, channel, purpose, s.clock.Now())
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return domain.VerificationCode{}, invalidErr
		}
		return domain.VerificationCode{}, fmt.Errorf("usecase: check code: find: %w", err)
	}

	if vc.AttemptsCount >= verificationCodeMaxAttempts {
		_ = s.codes.Delete(ctx, vc.ID)
		return domain.VerificationCode{}, invalidErr
	}

	if hashCode(presentedCode) != vc.CodeHash {
		if err := s.codes.IncrementAttempts(ctx, vc.ID); err != nil {
			return domain.VerificationCode{}, fmt.Errorf("usecase: check code: increment attempts: %w", err)
		}
		return domain.VerificationCode{}, invalidErr
	}

	return vc, nil
}

func generateNumericCode() (string, error) {
	const digits = "0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := make([]byte, 6)
	for i, v := range b {
		code[i] = digits[int(v)%len(digits)]
	}
	return string(code), nil
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

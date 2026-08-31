package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// invalidCredentialsMessage — единое сообщение для всех причин отказа в логине (неизвестный
// email, неверный пароль, забаненный/удалённый аккаунт) — анти-энумерация: атакующий,
// перебирающий email, не должен отличить эти случаи по ответу.
//
//nolint:gosec // G101 ложное срабатывание — это текст сообщения об ошибке, не секрет
const invalidCredentialsMessage = "неверный email или пароль"

// AccountService — регистрация, логин и чтение профиля пользователя (E2.4).
type AccountService struct {
	users          UserRepo
	hasher         *password.Hasher
	sessions       *SessionService
	clock          clock.Clock
	googleVerifier GoogleIDTokenVerifier
	oauthRepo      OAuthIdentityRepo
}

// NewAccountService создаёт AccountService.
func NewAccountService(
	users UserRepo,
	hasher *password.Hasher,
	sessions *SessionService,
	clk clock.Clock,
	googleVerifier GoogleIDTokenVerifier,
	oauthRepo OAuthIdentityRepo,
) *AccountService {
	return &AccountService{
		users: users, hasher: hasher, sessions: sessions, clock: clk,
		googleVerifier: googleVerifier, oauthRepo: oauthRepo,
	}
}

// Register создаёт нового пользователя с ролью "user" и статусом "unverified". email/password/
// consentVersion уже провалидированы транспортным слоем на синтаксис — здесь только оркестрация.
func (s *AccountService) Register(ctx context.Context, email, plainPassword, consentVersion string) (domain.User, error) {
	normalized := normalizeEmail(email)

	hash, err := s.hasher.Hash(plainPassword)
	if err != nil {
		return domain.User{}, fmt.Errorf("usecase: register: hash password: %w", err)
	}

	now := s.clock.Now()
	u := domain.User{
		Email:          normalized,
		PasswordHash:   &hash,
		Role:           "user",
		Status:         "unverified",
		ConsentAt:      now,
		ConsentVersion: consentVersion,
	}

	created, err := s.users.Create(ctx, u)
	if err != nil {
		return domain.User{}, err
	}

	return created, nil
}

// Login проверяет email+пароль и выпускает новую сессию (access+refresh). Все причины отказа
// (неизвестный email, неверный пароль, google-only аккаунт без пароля исключение — см. ниже,
// забаненный/удалённый аккаунт) — единое сообщение invalidCredentialsMessage, кроме
// google-only, у которого отдельный код для UX (подсказать войти через Google).
func (s *AccountService) Login(ctx context.Context, email, plainPassword string) (accessToken, refreshToken string, err error) {
	normalized := normalizeEmail(email)

	u, err := s.users.FindByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return "", "", apperr.Unauthorized(invalidCredentialsMessage)
		}
		return "", "", fmt.Errorf("usecase: login: find user: %w", err)
	}

	if u.PasswordHash == nil {
		return "", "", apperr.UnauthorizedCode("google_account_no_password", "аккаунт зарегистрирован через Google, войдите через Google")
	}

	ok, verifyErr := password.Verify(*u.PasswordHash, plainPassword)
	if verifyErr != nil || !ok {
		return "", "", apperr.Unauthorized(invalidCredentialsMessage)
	}

	if u.Status == "banned" || u.Status == "deleted" {
		return "", "", apperr.Unauthorized(invalidCredentialsMessage)
	}

	return s.sessions.Issue(ctx, u.ID, u.Role)
}

// Me возвращает профиль пользователя по id.
func (s *AccountService) Me(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

// googleProvider — значение поля auth.oauth_identities.provider для Google (единственный
// провайдер, поддерживаемый CHECK-констрейнтом миграции 00006).
const googleProvider = "google"

// LoginWithGoogle обменивает Google ID-токен на сессию (access+refresh). При первом входе
// через Google создаёт users+oauth_identities; при повторном — только выпускает сессию.
// consentVersion валидируется здесь (не в transport), т.к. нужен ТОЛЬКО в кейсе реального
// создания нового пользователя, что заранее не узнать до похода в БД (см. .claude/rules/go.md,
// "Размещение валидации" — это осознанное условное исключение, не нарушение слоёв).
func (s *AccountService) LoginWithGoogle(ctx context.Context, idToken, consentVersion string) (accessToken, refreshToken string, err error) {
	claims, err := s.googleVerifier.Verify(ctx, idToken)
	if err != nil {
		return "", "", apperr.Unauthorized("невалидный Google id-токен")
	}

	identity, err := s.oauthRepo.FindByProvider(ctx, googleProvider, claims.Subject)
	if err == nil {
		u, findErr := s.users.FindByID(ctx, identity.UserID)
		if findErr != nil {
			return "", "", fmt.Errorf("usecase: login with google: find linked user: %w", findErr)
		}
		return s.sessions.Issue(ctx, u.ID, u.Role)
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		return "", "", fmt.Errorf("usecase: login with google: find oauth identity: %w", err)
	}

	// Нет связки — первый вход через Google. Ищем по email: возможно пользователь уже
	// зарегистрирован через /auth/register и теперь впервые входит через Google.
	normalizedEmail := normalizeEmail(claims.Email)
	existingUser, err := s.users.FindByEmail(ctx, normalizedEmail)
	switch {
	case err == nil:
		// Пользователь с таким email уже существует — линковка, а не создание.
		if !claims.EmailVerified {
			return "", "", apperr.Unauthorized("Google не подтвердил email — привязка невозможна")
		}
		if createErr := s.oauthRepo.Create(ctx, domain.OAuthIdentity{
			UserID: existingUser.ID, Provider: googleProvider, ProviderUserID: claims.Subject,
		}); createErr != nil {
			return "", "", fmt.Errorf("usecase: login with google: link oauth identity: %w", createErr)
		}
		return s.sessions.Issue(ctx, existingUser.ID, existingUser.Role)

	case errors.Is(err, apperr.ErrNotFound):
		// Пользователя с таким email нет вообще — новая регистрация через Google.
		if !claims.EmailVerified {
			return "", "", apperr.Unauthorized("Google не подтвердил email")
		}
		if consentVersion == "" {
			return "", "", apperr.Invalid(
				map[string]string{"consent_version": "обязателен для первой регистрации через Google"},
				"некорректные данные регистрации через Google",
			)
		}

		now := s.clock.Now()
		created, createErr := s.users.Create(ctx, domain.User{
			Email:           normalizedEmail,
			PasswordHash:    nil,
			Role:            "user",
			Status:          "active",
			EmailVerifiedAt: &now,
			ConsentAt:       now,
			ConsentVersion:  consentVersion,
		})
		if createErr != nil {
			return "", "", createErr
		}

		if linkErr := s.oauthRepo.Create(ctx, domain.OAuthIdentity{
			UserID: created.ID, Provider: googleProvider, ProviderUserID: claims.Subject,
		}); linkErr != nil {
			return "", "", fmt.Errorf("usecase: login with google: link new oauth identity: %w", linkErr)
		}

		return s.sessions.Issue(ctx, created.ID, created.Role)

	default:
		return "", "", fmt.Errorf("usecase: login with google: find user by email: %w", err)
	}
}

// UpdateConsent фиксирует переподтверждение согласия (например после обновления политики) —
// отдельно от начального consent при регистрации.
func (s *AccountService) UpdateConsent(ctx context.Context, userID uuid.UUID, consentVersion string) error {
	return s.users.UpdateConsent(ctx, userID, consentVersion, s.clock.Now())
}

// DeleteMe — право на удаление аккаунта (закон РТ №1537): soft-delete + анонимизация + отзыв
// ВСЕХ активных сессий. Email анонимизируется детерминированно по id — не оставляет исходный
// email читаемым, но не требует отдельного хранилища "какой email был".
func (s *AccountService) DeleteMe(ctx context.Context, userID uuid.UUID) error {
	anonymizedEmail := fmt.Sprintf("deleted-%s@eduhub.local", userID.String())
	if err := s.users.SoftDelete(ctx, userID, anonymizedEmail, s.clock.Now()); err != nil {
		return fmt.Errorf("usecase: delete me: soft delete: %w", err)
	}
	return s.sessions.RevokeAllForUser(ctx, userID)
}

// normalizeEmail — lowercase+trim, единый канонический вид email для сравнения/хранения.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

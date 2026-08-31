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
	users    UserRepo
	hasher   *password.Hasher
	sessions *SessionService
	clock    clock.Clock
}

// NewAccountService создаёт AccountService.
func NewAccountService(users UserRepo, hasher *password.Hasher, sessions *SessionService, clk clock.Clock) *AccountService {
	return &AccountService{users: users, hasher: hasher, sessions: sessions, clock: clk}
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

// normalizeEmail — lowercase+trim, единый канонический вид email для сравнения/хранения.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

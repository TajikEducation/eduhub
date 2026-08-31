package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/security"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// invalidCredentials — единое сообщение для «нет юзера» и «неверный пароль» (anti-enumeration,
// см. SRS Веха 2, E2.4): по тексту ошибки нельзя понять, зарегистрирован ли email.
const invalidCredentials = "неверный email или пароль"

// TokenPair — пара токенов, возвращаемая Register/Login/Refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // секунды, для клиента
}

// Service — usecase-слой auth: регистрация, вход, refresh, чтение профиля.
type Service struct {
	users         UserRepo
	refreshTokens RefreshTokenRepo
	clock         clock.Clock

	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// Config — параметры, которыми Service параметризован (секрет и TTL — из конфигурации процесса,
// не хардкод, чтобы их можно было менять/тестировать).
type Config struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// New создаёт Service. clk — инжектируемые часы (clock.New() в проде, clock.Fake в тестах).
func New(users UserRepo, refreshTokens RefreshTokenRepo, clk clock.Clock, cfg Config) *Service {
	return &Service{
		users:         users,
		refreshTokens: refreshTokens,
		clock:         clk,
		jwtSecret:     cfg.JWTSecret,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

// Register создаёт нового пользователя и сразу выдаёт пару токенов (email-verification
// ещё не реализован — см. миграцию 00006, статус по умолчанию 'active').
func (s *Service) Register(ctx context.Context, in domain.RegisterInput) (domain.User, TokenPair, error) {
	if err := in.Validate(); err != nil {
		return domain.User{}, TokenPair{}, err
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return domain.User{}, TokenPair{}, apperr.Internal(fmt.Errorf("usecase: hash password: %w", err))
	}

	user := domain.User{
		Email:        in.NormalizedEmail(),
		PasswordHash: hash,
		Role:         domain.RoleUser,
		Status:       domain.StatusActive,
		DisplayName:  in.DisplayName,
	}

	created, err := s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, TokenPair{}, fmt.Errorf("usecase: create user: %w", err)
	}

	pair, err := s.issueTokenPair(ctx, created)
	if err != nil {
		return domain.User{}, TokenPair{}, err
	}
	return created, pair, nil
}

// Login проверяет email+пароль, выдаёт пару токенов. Ошибка формулировки всегда одна и та же
// (invalidCredentials) независимо от причины — не должно быть возможности узнать по ответу,
// существует ли аккаунт с этим email.
func (s *Service) Login(ctx context.Context, email, password string) (domain.User, TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return domain.User{}, TokenPair{}, apperr.Unauthorized(invalidCredentials)
		}
		return domain.User{}, TokenPair{}, fmt.Errorf("usecase: get user by email: %w", err)
	}

	if user.PasswordHash == "" {
		// Google-only аккаунт (пока не реализовано отдельно) — тот же общий текст ошибки,
		// не раскрываем механизм.
		return domain.User{}, TokenPair{}, apperr.Unauthorized(invalidCredentials)
	}

	ok, err := security.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return domain.User{}, TokenPair{}, apperr.Internal(fmt.Errorf("usecase: verify password: %w", err))
	}
	if !ok {
		return domain.User{}, TokenPair{}, apperr.Unauthorized(invalidCredentials)
	}

	if user.Status != domain.StatusActive {
		return domain.User{}, TokenPair{}, apperr.Forbidden("аккаунт заблокирован или удалён")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return domain.User{}, TokenPair{}, err
	}
	return user, pair, nil
}

// Refresh ротирует refresh-токен: старый отзывается, новый выдаётся вместе со свежим access.
// Reuse-detection (отзыв всей family_id при повторном предъявлении уже использованного токена)
// сознательно не реализован в этой версии — см. миграцию 00006, family_id только записывается.
func (s *Service) Refresh(ctx context.Context, rawToken string) (TokenPair, error) {
	hash := security.HashRefreshToken(rawToken)

	record, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return TokenPair{}, fmt.Errorf("usecase: get refresh token: %w", err)
	}

	now := s.clock.Now()
	if record.RevokedAt != nil || now.After(record.ExpiresAt) {
		return TokenPair{}, apperr.Unauthorized("refresh token invalid or expired")
	}

	user, err := s.users.GetByID(ctx, record.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("usecase: get user by id: %w", err)
	}

	newRawRefresh, newHash, err := security.NewRefreshToken()
	if err != nil {
		return TokenPair{}, apperr.Internal(fmt.Errorf("usecase: generate refresh token: %w", err))
	}
	newID := uuid.New()

	if err := s.refreshTokens.Store(ctx, RefreshTokenRecord{
		ID:        newID,
		UserID:    user.ID,
		TokenHash: newHash,
		FamilyID:  record.FamilyID,
		ExpiresAt: now.Add(s.refreshTTL),
	}); err != nil {
		return TokenPair{}, fmt.Errorf("usecase: store rotated refresh token: %w", err)
	}

	if err := s.refreshTokens.Revoke(ctx, record.ID, newID); err != nil {
		return TokenPair{}, fmt.Errorf("usecase: revoke old refresh token: %w", err)
	}

	access, err := security.IssueAccessToken(s.jwtSecret, user.ID, string(user.Role), s.accessTTL)
	if err != nil {
		return TokenPair{}, apperr.Internal(fmt.Errorf("usecase: issue access token: %w", err))
	}

	return TokenPair{AccessToken: access, RefreshToken: newRawRefresh, ExpiresIn: int(s.accessTTL.Seconds())}, nil
}

// Me возвращает профиль по id (для GET /auth/me).
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("usecase: get user: %w", err)
	}
	return user, nil
}

// issueTokenPair выпускает access-токен и создаёт новую refresh-семью (family_id) для user.
func (s *Service) issueTokenPair(ctx context.Context, user domain.User) (TokenPair, error) {
	access, err := security.IssueAccessToken(s.jwtSecret, user.ID, string(user.Role), s.accessTTL)
	if err != nil {
		return TokenPair{}, apperr.Internal(fmt.Errorf("usecase: issue access token: %w", err))
	}

	rawRefresh, hash, err := security.NewRefreshToken()
	if err != nil {
		return TokenPair{}, apperr.Internal(fmt.Errorf("usecase: generate refresh token: %w", err))
	}

	now := s.clock.Now()
	if err := s.refreshTokens.Store(ctx, RefreshTokenRecord{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		FamilyID:  uuid.New(),
		ExpiresAt: now.Add(s.refreshTTL),
	}); err != nil {
		return TokenPair{}, fmt.Errorf("usecase: store refresh token: %w", err)
	}

	return TokenPair{AccessToken: access, RefreshToken: rawRefresh, ExpiresIn: int(s.accessTTL.Seconds())}, nil
}

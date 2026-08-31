package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// accessTokenTTLSeconds — TTL access-токена в секундах для поля expires_in ответа.
// Значение синхронизировано с cmd/api.accessTokenTTL (15 минут) — единственное место
// конфигурации TTL остаётся в cmd/api, здесь только константа для сериализации ответа.
const accessTokenTTLSeconds = 15 * 60

// minPasswordLength — минимальная длина пароля при регистрации (синтаксическая валидация).
const minPasswordLength = 8

// accountService — то, что нужно транспорту от usecase-слоя аккаунтов.
type accountService interface {
	Register(ctx context.Context, email, password, consentVersion string) (domain.User, error)
	Login(ctx context.Context, email, password string) (access, refresh string, err error)
	Me(ctx context.Context, userID uuid.UUID) (domain.User, error)
	LoginWithGoogle(ctx context.Context, idToken, consentVersion string) (access, refresh string, err error)
}

// sessionService — то, что нужно транспорту от usecase-слоя сессий.
type sessionService interface {
	Rotate(ctx context.Context, refreshToken string) (access, refresh string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

// RegisterHandler — POST /auth/register.
func RegisterHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		fields := map[string]string{}
		if req.Email == "" || !strings.Contains(req.Email, "@") {
			fields["email"] = "обязателен и должен содержать @"
		}
		if len(req.Password) < minPasswordLength {
			fields["password"] = "минимум 8 символов"
		}
		if req.ConsentVersion == "" {
			fields["consent_version"] = "обязателен"
		}
		if len(fields) > 0 {
			httpx.WriteError(w, r, logger, apperr.Invalid(fields, "некорректные данные регистрации"))
			return
		}

		u, err := svc.Register(r.Context(), req.Email, req.Password, req.ConsentVersion)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusCreated, registerResponse{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		})
	}
}

// LoginHandler — POST /auth/login.
func LoginHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		fields := map[string]string{}
		if req.Email == "" {
			fields["email"] = "обязателен"
		}
		if req.Password == "" {
			fields["password"] = "обязателен"
		}
		if len(fields) > 0 {
			httpx.WriteError(w, r, logger, apperr.Invalid(fields, "некорректные данные входа"))
			return
		}

		access, refresh, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, tokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			ExpiresIn:    accessTokenTTLSeconds,
		})
	}
}

// RefreshHandler — POST /auth/refresh.
func RefreshHandler(svc sessionService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if req.RefreshToken == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"refresh_token": "обязателен"}, "некорректные данные"))
			return
		}

		access, refresh, err := svc.Rotate(r.Context(), req.RefreshToken)
		if err != nil {
			httpx.WriteError(w, r, logger, mapRotateError(err))
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, tokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			ExpiresIn:    accessTokenTTLSeconds,
		})
	}
}

// LogoutHandler — POST /auth/logout.
func LogoutHandler(svc sessionService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req logoutRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if req.RefreshToken == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"refresh_token": "обязателен"}, "некорректные данные"))
			return
		}

		if err := svc.Logout(r.Context(), req.RefreshToken); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, logoutResponse{Status: "ok"})
	}
}

// OAuthGoogleHandler — POST /auth/oauth/google. consent_version НЕ валидируется здесь (400) —
// он нужен только в кейсе реального создания нового пользователя, что в transport-слое заранее
// не узнать без похода в БД; условная валидация оставлена в usecase (см. AccountService.LoginWithGoogle).
func OAuthGoogleHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req oauthGoogleRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		if req.IDToken == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"id_token": "обязателен"}, "некорректные данные"))
			return
		}

		access, refresh, err := svc.LoginWithGoogle(r.Context(), req.IDToken, req.ConsentVersion)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, tokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			ExpiresIn:    accessTokenTTLSeconds,
		})
	}
}

// MeHandler — GET /auth/me. Требует Principal из контекста (RequireAuth уже отработал).
func MeHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			// Не должно случиться в норме — RequireAuth уже гарантирует Principal в контексте.
			// Если случилось — баг wiring, не ошибка вызывающего.
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}

		u, err := svc.Me(r.Context(), principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, meResponse{
			ID:              u.ID,
			Email:           u.Email,
			DisplayName:     u.DisplayName,
			Role:            u.Role,
			Status:          u.Status,
			EmailVerifiedAt: u.EmailVerifiedAt,
			CreatedAt:       u.CreatedAt,
		})
	}
}

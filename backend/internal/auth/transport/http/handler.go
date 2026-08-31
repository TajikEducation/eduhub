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
	UpdateConsent(ctx context.Context, userID uuid.UUID, consentVersion string) error
	DeleteMe(ctx context.Context, userID uuid.UUID) error
}

// sessionService — то, что нужно транспорту от usecase-слоя сессий.
type sessionService interface {
	Rotate(ctx context.Context, refreshToken string) (access, refresh string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

// verificationService — то, что нужно транспорту от usecase-слоя верификации/password-reset.
type verificationService interface {
	RequestEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, code string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error
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

// VerifyEmailHandler — POST /auth/verify.
func VerifyEmailHandler(svc verificationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req verifyEmailRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		fields := map[string]string{}
		if req.Email == "" {
			fields["email"] = "обязателен"
		}
		if req.Code == "" {
			fields["code"] = "обязателен"
		}
		if len(fields) > 0 {
			httpx.WriteError(w, r, logger, apperr.Invalid(fields, "некорректные данные"))
			return
		}
		if err := svc.VerifyEmail(r.Context(), req.Email, req.Code); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// ResendVerificationHandler — POST /auth/verify/resend. Всегда 200 (anti-enumeration) —
// см. комментарий VerificationService.RequestEmailVerification.
func ResendVerificationHandler(svc verificationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req resendVerificationRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		if req.Email == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"email": "обязателен"}, "некорректные данные"))
			return
		}
		if err := svc.RequestEmailVerification(r.Context(), req.Email); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// PasswordResetRequestHandler — POST /auth/password/reset-request. Всегда 200 (anti-enumeration).
func PasswordResetRequestHandler(svc verificationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req passwordResetRequestRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		if req.Email == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"email": "обязателен"}, "некорректные данные"))
			return
		}
		if err := svc.RequestPasswordReset(r.Context(), req.Email); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// PasswordResetConfirmHandler — POST /auth/password/reset-confirm.
func PasswordResetConfirmHandler(svc verificationService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req passwordResetConfirmRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		fields := map[string]string{}
		if req.Email == "" {
			fields["email"] = "обязателен"
		}
		if req.Code == "" {
			fields["code"] = "обязателен"
		}
		if len(req.NewPassword) < minPasswordLength {
			fields["new_password"] = "минимум 8 символов"
		}
		if len(fields) > 0 {
			httpx.WriteError(w, r, logger, apperr.Invalid(fields, "некорректные данные"))
			return
		}
		if err := svc.ConfirmPasswordReset(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// ConsentHandler — POST /auth/consent, защищён RequireAuth.
func ConsentHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}
		var req consentRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		if req.ConsentVersion == "" {
			httpx.WriteError(w, r, logger, apperr.Invalid(map[string]string{"consent_version": "обязателен"}, "некорректные данные"))
			return
		}
		if err := svc.UpdateConsent(r.Context(), principal.UserID, req.ConsentVersion); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		_ = httpx.WriteJSON(w, logger, http.StatusOK, statusResponse{Status: "ok"})
	}
}

// DeleteMeHandler — DELETE /auth/me, защищён RequireAuth.
func DeleteMeHandler(svc accountService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Internal(errNoPrincipalInContext))
			return
		}
		if err := svc.DeleteMe(r.Context(), principal.UserID); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

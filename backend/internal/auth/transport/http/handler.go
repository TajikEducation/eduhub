// Package http — транспортный слой auth поверх net/http: register/login/refresh/me.
package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// authService — то, что нужно транспорту от usecase-слоя (порт, определённый потребителем).
type authService interface {
	Register(ctx context.Context, in domain.RegisterInput) (domain.User, usecase.TokenPair, error)
	Login(ctx context.Context, email, password string) (domain.User, usecase.TokenPair, error)
	Refresh(ctx context.Context, rawToken string) (usecase.TokenPair, error)
	Me(ctx context.Context, userID uuid.UUID) (domain.User, error)
}

// RegisterHandler — POST /auth/register.
func RegisterHandler(svc authService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		user, tokens, err := svc.Register(r.Context(), domain.RegisterInput{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		})
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusCreated, authResponse{User: toUserDTO(user), Tokens: toTokenPairDTO(tokens)})
	}
}

// LoginHandler — POST /auth/login.
func LoginHandler(svc authService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		user, tokens, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, authResponse{User: toUserDTO(user), Tokens: toTokenPairDTO(tokens)})
	}
}

// RefreshHandler — POST /auth/refresh.
func RefreshHandler(svc authService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		tokens, err := svc.Refresh(r.Context(), req.RefreshToken)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toTokenPairDTO(tokens))
	}
}

// MeHandler — GET /auth/me. Должен идти за rbac.RequireAuth — читает Principal из контекста.
func MeHandler(svc authService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := rbac.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
			return
		}

		user, err := svc.Me(r.Context(), principal.UserID)
		if err != nil {
			httpx.WriteError(w, r, logger, err)
			return
		}

		_ = httpx.WriteJSON(w, logger, http.StatusOK, toUserDTO(user))
	}
}

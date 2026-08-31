// Package http — транспортный слой ядра auth-эндпоинтов (E2.4, веха 2):
// register/login/refresh/logout/me.
package http

import (
	"time"

	"github.com/google/uuid"
)

type registerRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	ConsentVersion string `json:"consent_version"`
}

type registerResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutResponse struct {
	Status string `json:"status"`
}

type oauthGoogleRequest struct {
	IDToken        string `json:"id_token"`
	ConsentVersion string `json:"consent_version"`
}

type meResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	DisplayName     *string    `json:"display_name,omitempty"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

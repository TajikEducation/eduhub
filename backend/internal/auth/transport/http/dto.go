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

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type passwordResetRequestRequest struct {
	Email string `json:"email"`
}

type passwordResetConfirmRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type consentRequest struct {
	ConsentVersion string `json:"consent_version"`
}

type statusResponse struct {
	Status string `json:"status"`
}

type createChildRequest struct {
	InstitutionID uuid.UUID `json:"institution_id"`
	AgeGroup      string    `json:"age_group"`
	Status        string    `json:"status"`
}

type rejectChildRequest struct {
	ReasonCode string  `json:"reason_code"`
	ReasonText *string `json:"reason_text,omitempty"`
}

type childResponse struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	InstitutionID      uuid.UUID  `json:"institution_id"`
	AgeGroup           string     `json:"age_group"`
	Status             string     `json:"status"`
	ConfirmationStatus string     `json:"confirmation_status"`
	ConfirmedBy        *uuid.UUID `json:"confirmed_by,omitempty"`
	ConfirmedAt        *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

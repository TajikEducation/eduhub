package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
)

// registerRequest — тело POST /auth/register.
type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// loginRequest — тело POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// refreshRequest — тело POST /auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// userDTO — публичный профиль пользователя (без password_hash).
type userDTO struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	DisplayName string    `json:"display_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toUserDTO(u domain.User) userDTO {
	return userDTO{
		ID:          u.ID,
		Email:       u.Email,
		Role:        string(u.Role),
		Status:      string(u.Status),
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}

// tokenPairDTO — тело ответа с парой токенов.
type tokenPairDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func toTokenPairDTO(p usecase.TokenPair) tokenPairDTO {
	return tokenPairDTO{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresIn:    p.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// authResponse — тело ответа register/login: профиль + токены одним объектом.
type authResponse struct {
	User   userDTO      `json:"user"`
	Tokens tokenPairDTO `json:"tokens"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthIdentity — строка auth.oauth_identities: связка пользователя с внешним провайдером
// (сейчас только "google") по стабильному provider_user_id (sub).
type OAuthIdentity struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	CreatedAt      time.Time
}

package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AccessClaims — полезная нагрузка access-токена. HS256 достаточен, пока валидатор один
// (монолит) — см. SRS Веха 2, E2.3.
type AccessClaims struct {
	UserID uuid.UUID `json:"sub"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// IssueAccessToken выпускает подписанный HS256 access-токен с TTL.
func IssueAccessToken(secret string, userID uuid.UUID, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("security: sign access token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken проверяет подпись и срок действия, возвращает claims. Любая ошибка (истёк,
// неверная подпись, неверный алгоритм) сворачивается в один общий error — вызывающий код
// (RBAC middleware) не обязан различать причины отказа, для клиента это всегда 401.
func ParseAccessToken(secret, raw string) (AccessClaims, error) {
	var claims AccessClaims
	token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return AccessClaims{}, fmt.Errorf("security: invalid access token: %w", err)
	}
	return claims, nil
}

// refreshTokenBytes — энтропия сырого refresh-токена перед хешированием.
const refreshTokenBytes = 32

// NewRefreshToken генерирует случайный refresh-токен (возвращается клиенту один раз) и его
// хеш (то, что хранится в БД — см. SRS: «хранится только хеш»).
func NewRefreshToken() (raw string, hash string, err error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("security: generate refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken — детерминированный хеш сырого refresh-токена для поиска/сравнения в БД
// (SHA-256 достаточен: токен уже высокоэнтропийный случайный секрет, не пароль пользователя).
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

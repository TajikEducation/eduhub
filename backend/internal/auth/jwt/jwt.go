// Package jwt — выпуск и проверка access-токенов (E2.3, веха 2). HS256: достаточно, пока
// валидатор токенов один и тот же процесс, что и выпускающий (монолит) — переход на
// RS256/JWKS понадобится только если появится второй независимый потребитель токенов.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// ErrInvalidToken — токен не прошёл проверку: битая подпись, истёк, неверный формат.
// Единый sentinel — вызывающему на этом уровне не нужно различать причину отказа.
var ErrInvalidToken = errors.New("jwt: invalid or expired token")

// Claims — полезная нагрузка access-токена. Role — кастомное поле, остальное — RFC 7519.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// UserID парсит Subject (строка) обратно в UUID.
func (c Claims) UserID() (uuid.UUID, error) {
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt: claims.sub is not a valid UUID: %w", err)
	}
	return id, nil
}

// Issuer выпускает и проверяет access-токены с фиксированным секретом и TTL.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	clock  clock.Clock
}

// NewIssuer создаёт Issuer. clk — инжектируемые часы (реальные в проде, Fake в тестах —
// без time.Sleep, тот же паттерн, что internal/platform/httpx rate-limit, задача 15).
func NewIssuer(secret []byte, ttl time.Duration, clk clock.Clock) *Issuer {
	return &Issuer{secret: secret, ttl: ttl, clock: clk}
}

// Issue выпускает подписанный access-токен для пользователя userID с ролью role.
func (i *Issuer) Issue(userID uuid.UUID, role string) (string, error) {
	now := i.clock.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: sign token: %w", err)
	}
	return signed, nil
}

// Parse проверяет подпись и срок действия токена, возвращает разобранные Claims.
// Любая проблема (подпись, срок, формат) — единый ErrInvalidToken, детали не пробрасываются:
// вызывающему на этом уровне не нужно различать причину, только факт «токен недействителен».
//
// jwt/v4 умеет сверять exp только через ПАКЕТНУЮ переменную jwt.TimeFunc (per-parser опции
// инъекции времени в v4 нет, появилась только в v5) — мутировать глобальную переменную под
// конкурентными HTTP-запросами было бы гонкой. Поэтому здесь: WithoutClaimsValidation()
// отключает встроенную проверку exp/nbf/iat (подпись по-прежнему проверяется), срок действия
// сверяется вручную через инжектированные часы Issuer.clock — тот же принцип, что и у
// rate-limiter'а (platform/clock), тестируемо без time.Sleep.
func (i *Issuer) Parse(tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		// Защита от alg-confusion: принимаем только HMAC, каким бы ни был alg в заголовке
		// присланного токена.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method %v", t.Header["alg"])
		}
		return i.secret, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.ExpiresAt == nil || !i.clock.Now().Before(claims.ExpiresAt.Time) {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

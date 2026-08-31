package http

import (
	"errors"

	"github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// invalidRefreshTokenMessage — единое сообщение для всех причин отказа предъявленного
// refresh-токена (неизвестен/истёк/reuse обнаружен) — вызывающему клиенту не нужно различать
// причину, только факт «нужно перелогиниться».
//
//nolint:gosec // G101 ложное срабатывание — это текст сообщения об ошибке, не секрет
const invalidRefreshTokenMessage = "невалидный или истёкший refresh-токен"

// errNoPrincipalInContext — используется только для apperr.Internal при отсутствии Principal
// в контексте MeHandler — сигнал бага wiring (RequireAuth не отработал), не бизнес-ошибка.
var errNoPrincipalInContext = errors.New("auth/transport/http: no Principal in request context")

// mapRotateError переводит sentinel-ошибки usecase.SessionService.Rotate в apperr.Unauthorized —
// иначе httpx.WriteError классифицировал бы их как неизвестные и вернул 500.
func mapRotateError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrRefreshTokenNotFound),
		errors.Is(err, usecase.ErrRefreshTokenExpired),
		errors.Is(err, usecase.ErrRefreshTokenReused):
		return apperr.Unauthorized(invalidRefreshTokenMessage)
	default:
		return err
	}
}

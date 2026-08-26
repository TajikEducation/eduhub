package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// maxBodyBytes — верхняя граница размера тела запроса, защита от DoS через огромные payload'ы.
const maxBodyBytes = 1 << 20 // 1 MiB

// DecodeJSON декодирует тело запроса в dst с лимитом размера и запретом неизвестных полей;
// любая ошибка декодирования классифицируется в apperr.Invalid — вызывающий код не разбирает
// *json.SyntaxError и т.п. сам.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}

	return nil
}

// decodeError классифицирует сырую ошибку decoder.Decode в apperr.Invalid по тексту:
// стандартная библиотека не даёт типизированной ошибки для "тело слишком большое".
func decodeError(err error) error {
	switch {
	case strings.Contains(err.Error(), "http: request body too large"):
		return apperr.Invalid(map[string]string{"body": "body_too_large"}, "request body exceeds 1MB limit")
	case strings.Contains(err.Error(), "unknown field"):
		return apperr.Invalid(map[string]string{"body": "unknown_field"}, err.Error())
	default:
		return apperr.Invalid(map[string]string{"body": "malformed_json"}, "malformed request body")
	}
}

// WriteJSON сериализует v в ответ через потоковый Encoder (без промежуточного буфера от Marshal),
// выставляя Content-Type до WriteHeader; ошибка Encode логируется (тело уже частично отправлено —
// поздно паниковать) и возвращается вызывающему, который сам решает, как завершить обработку запроса.
func WriteJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn("failed to write json response body", slog.Any("error", err))
		return err
	}

	return nil
}

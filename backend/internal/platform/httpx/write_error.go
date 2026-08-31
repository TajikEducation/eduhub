package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// internalMessage — фиксированный текст для клиента при 500: cause никогда не уходит наружу.
const internalMessage = "internal server error"

// rateLimitRetryAfterSeconds — временный дефолт, пока нет реального rate-limiter'а (задача 15).
const rateLimitRetryAfterSeconds = 60

// errorResponse — контракт тела ответа, см. docs/EduHub_Backend_Architecture.md, раздел 8.
type errorResponse struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id"`
}

type errorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteError — единственное место, где apperr.Error превращается в HTTP-ответ.
// Internal и незнакомые ошибки логируются на Error (это инциденты), остальные категории — нет
// (ожидаемые бизнес-ошибки).
func WriteError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, code, message, fields, retryAfter := classify(err)

	if status == http.StatusInternalServerError {
		logger.Error("request failed",
			slog.String("error", err.Error()),
			slog.String("request_id", RequestID(r.Context())),
		)
	}

	body := errorResponse{
		Error: errorPayload{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		RequestID: RequestID(r.Context()),
	}

	payload, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		// Маршалинг фиксированной структуры не должен падать — но если случилось,
		// это тоже инцидент, а не бизнес-ошибка.
		logger.Error("failed to marshal error response", slog.Any("error", marshalErr))
		payload = []byte(`{"error":{"code":"internal","message":"internal server error"}}`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	}
	w.WriteHeader(status)

	if _, writeErr := w.Write(payload); writeErr != nil {
		logger.Warn("failed to write error response body", slog.Any("error", writeErr))
	}
}

// classify маппит apperr-категорию (или произвольную ошибку) на HTTP-статус и тело ответа.
func classify(err error) (status int, code string, message string, fields map[string]string, retryAfterSeconds int) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, errCode(err, "not_found"), errMessage(err), nil, 0
	case errors.Is(err, apperr.ErrInvalid):
		return http.StatusBadRequest, errCode(err, "invalid_input"), errMessage(err), errFields(err), 0
	case errors.Is(err, apperr.ErrUnauthorized):
		return http.StatusUnauthorized, errCode(err, "unauthorized"), errMessage(err), nil, 0
	case errors.Is(err, apperr.ErrForbidden):
		return http.StatusForbidden, errCode(err, "forbidden"), errMessage(err), nil, 0
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, errCode(err, "conflict"), errMessage(err), nil, 0
	case errors.Is(err, apperr.ErrRateLimited):
		return http.StatusTooManyRequests, errCode(err, "rate_limited"), errMessage(err), nil, rateLimitRetryAfterSeconds
	default:
		// apperr.ErrInternal и любая произвольная ошибка (не *apperr.Error) — одна ветка:
		// клиенту всегда фиксированный текст, cause никогда не уходит наружу.
		return http.StatusInternalServerError, "internal", internalMessage, nil, 0
	}
}

// errMessage достаёт человекочитаемое сообщение через apperr.Error.Message(), если err им является.
func errMessage(err error) string {
	var target *apperr.Error
	if errors.As(err, &target) {
		return target.Message()
	}
	return err.Error()
}

// errCode достаёт переопределённый code через apperr.Error.Code(), если он непустой; иначе
// возвращает fallback (дефолтный код категории).
func errCode(err error, fallback string) string {
	var target *apperr.Error
	if errors.As(err, &target) {
		if code := target.Code(); code != "" {
			return code
		}
	}
	return fallback
}

// errFields достаёт map полей для invalid_input, если err — *apperr.Error.
func errFields(err error) map[string]string {
	var target *apperr.Error
	if errors.As(err, &target) {
		return target.Fields
	}
	return nil
}

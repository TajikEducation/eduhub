// Package apperr — единая таксономия доменных ошибок бэкенда EduHub, маппится в HTTP одним местом (httpx.WriteError).
package apperr

// Sentinel-ошибки категорий, сравниваются через errors.Is (работает и через fmt.Errorf("...: %w", err)).
var (
	ErrNotFound     = newSentinel("not found")      // HTTP 404
	ErrInvalid      = newSentinel("invalid input")  // HTTP 400, несёт Fields
	ErrUnauthorized = newSentinel("unauthorized")   // HTTP 401
	ErrForbidden    = newSentinel("forbidden")      // HTTP 403
	ErrConflict     = newSentinel("conflict")       // HTTP 409
	ErrRateLimited  = newSentinel("rate limited")   // HTTP 429
	ErrInternal     = newSentinel("internal error") // HTTP 500, cause только для логов
)

// sentinel — тип ошибки-категории, отдельный от errors.errorString.
type sentinel struct {
	text string
}

func (s *sentinel) Error() string { return s.text }

func newSentinel(text string) error {
	return &sentinel{text: text}
}

// Error — доменная ошибка EduHub: категория + опциональный контекст.
type Error struct {
	category error
	message  string
	cause    error

	// Fields — какое поле что не так, заполняется только для Invalid.
	Fields map[string]string
}

func (e *Error) Error() string {
	if e.message == "" {
		return e.category.Error()
	}
	return e.category.Error() + ": " + e.message
}

// Is матчит *Error с sentinel-категорией для errors.Is.
func (e *Error) Is(target error) bool {
	return e.category == target
}

// Unwrap открывает cause — например, чтобы errors.Is(err, sql.ErrNoRows) сработал после apperr.Internal(sql.ErrNoRows).
func (e *Error) Unwrap() error {
	return e.cause
}

// NotFound: resource — что искали, id — по какому идентификатору не нашли.
func NotFound(resource string, id string) error {
	return &Error{
		category: ErrNotFound,
		message:  resource + " not found: id=" + id,
	}
}

// Invalid: fields — поле→причина, доступны через errors.As(err, &target) как target.Fields.
func Invalid(fields map[string]string, message string) error {
	return &Error{
		category: ErrInvalid,
		message:  message,
		Fields:   fields,
	}
}

func Unauthorized(message string) error {
	return &Error{category: ErrUnauthorized, message: message}
}

func Forbidden(message string) error {
	return &Error{category: ErrForbidden, message: message}
}

func Conflict(message string) error {
	return &Error{category: ErrConflict, message: message}
}

func RateLimited(message string) error {
	return &Error{category: ErrRateLimited, message: message}
}

// Internal оборачивает техническую ошибку err — не показывать пользователю текстом, только логировать.
func Internal(err error) error {
	return &Error{
		category: ErrInternal,
		message:  "internal error",
		cause:    err,
	}
}

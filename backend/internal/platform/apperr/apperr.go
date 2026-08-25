// Package apperr определяет единую таксономию доменных ошибок бэкенда
// EduHub: NotFound, Invalid, Unauthorized, Forbidden, Conflict, RateLimited,
// Internal. Доменные пакеты (usecase, domain) возвращают ошибки этого
// пакета вместо голых fmt.Errorf, а транспортный слой (httpx.WriteError,
// отдельная задача) сопоставляет их с HTTP-кодом ровно в одном месте
// кодовой базы. Сравнение — через errors.Is/errors.As, не строковый парсинг
// текста ошибки. Не зависит от транспорта (net/http) или драйверов БД
// (pgx) — чистый платформенный строительный блок.
package apperr

// Sentinel-ошибки категорий — для сравнения через errors.Is. Конструкторы
// ниже возвращают *Error, который реализует Is(target error) bool и
// матчится с соответствующим sentinel — в том числе через цепочку
// fmt.Errorf("...: %w", err), так как errors.Is разворачивает обёртки и
// вызывает Is на каждом уровне.
var (
	// ErrNotFound — запрошенный ресурс не существует. HTTP 404.
	ErrNotFound = newSentinel("not found")
	// ErrInvalid — вход не прошёл валидацию. Несёт Fields через *Error.
	// HTTP 400.
	ErrInvalid = newSentinel("invalid input")
	// ErrUnauthorized — запрос не аутентифицирован (нет/невалиден токен).
	// HTTP 401.
	ErrUnauthorized = newSentinel("unauthorized")
	// ErrForbidden — запрос аутентифицирован, но не авторизован на
	// операцию. HTTP 403.
	ErrForbidden = newSentinel("forbidden")
	// ErrConflict — операция конфликтует с текущим состоянием ресурса
	// (дубликат, гонка версий). HTTP 409.
	ErrConflict = newSentinel("conflict")
	// ErrRateLimited — превышен лимит запросов. HTTP 429.
	ErrRateLimited = newSentinel("rate limited")
	// ErrInternal — непредвиденная техническая ошибка. Пользователю
	// показывается только обобщённое сообщение, исходная ошибка (cause)
	// предназначена только для логирования. HTTP 500.
	ErrInternal = newSentinel("internal error")
)

// sentinel — минимальный тип ошибки-категории, сравнимый через == (как
// errors.New), но объявленный отдельным типом, чтобы не полагаться на
// конкретную реализацию errors.errorString.
type sentinel struct {
	text string
}

func (s *sentinel) Error() string { return s.text }

// newSentinel создаёт новую sentinel-ошибку категории. Неэкспортирована —
// используется только для объявления Err* переменных этого пакета, извне
// пакета sentinel-категории не создаются.
func newSentinel(text string) error {
	return &sentinel{text: text}
}

// Error — типизированная доменная ошибка EduHub. Оборачивает sentinel
// категории и опциональный контекст: человекочитаемое сообщение, набор
// невалидных полей (только для Invalid) и исходную техническую ошибку
// (только для Internal, для логирования).
type Error struct {
	category error
	message  string
	cause    error

	// Fields — какое поле что не так, заполняется только для Invalid.
	Fields map[string]string
}

// Error реализует интерфейс error.
func (e *Error) Error() string {
	if e.message == "" {
		return e.category.Error()
	}
	return e.category.Error() + ": " + e.message
}

// Is сообщает errors.Is, что *Error относится к категории target — вызов
// errors.Is(err, apperr.ErrNotFound) матчит любую ошибку, созданную через
// apperr.NotFound(...), в том числе после fmt.Errorf("...: %w", err),
// потому что errors.Is разворачивает обёртки и на каждом уровне
// цепочки вызывает Is у текущей ошибки.
func (e *Error) Is(target error) bool {
	return e.category == target
}

// Unwrap открывает доступ к исходной технической ошибке (cause),
// обёрнутой через Internal — например, чтобы errors.Is(err, sql.ErrNoRows)
// сработал даже после apperr.Internal(sql.ErrNoRows). Для остальных
// категорий cause не задан, Unwrap возвращает nil.
func (e *Error) Unwrap() error {
	return e.cause
}

// NotFound создаёт ошибку категории ErrNotFound: resource — что искали
// (например, "institution"), id — по какому идентификатору не нашли.
func NotFound(resource string, id string) error {
	return &Error{
		category: ErrNotFound,
		message:  resource + " not found: id=" + id,
	}
}

// Invalid создаёт ошибку категории ErrInvalid. fields — какое поле что не
// так (ключ — имя поля, значение — причина), message — общее описание
// проблемы. fields доступны типизированно через errors.As(err, &target)
// как target.Fields.
func Invalid(fields map[string]string, message string) error {
	return &Error{
		category: ErrInvalid,
		message:  message,
		Fields:   fields,
	}
}

// Unauthorized создаёт ошибку категории ErrUnauthorized (запрос не
// аутентифицирован).
func Unauthorized(message string) error {
	return &Error{category: ErrUnauthorized, message: message}
}

// Forbidden создаёт ошибку категории ErrForbidden (запрос аутентифицирован,
// но не авторизован на операцию).
func Forbidden(message string) error {
	return &Error{category: ErrForbidden, message: message}
}

// Conflict создаёт ошибку категории ErrConflict (операция конфликтует с
// текущим состоянием ресурса).
func Conflict(message string) error {
	return &Error{category: ErrConflict, message: message}
}

// RateLimited создаёт ошибку категории ErrRateLimited (превышен лимит
// запросов).
func RateLimited(message string) error {
	return &Error{category: ErrRateLimited, message: message}
}

// Internal оборачивает исходную техническую ошибку err в категорию
// ErrInternal. err не должен показываться пользователю текстом — только
// логироваться (доступен через errors.Unwrap/errors.Is); HTTP-ответ на
// Internal всегда обобщённое сообщение (маппинг — в httpx.WriteError,
// отдельная задача).
func Internal(err error) error {
	return &Error{
		category: ErrInternal,
		message:  "internal error",
		cause:    err,
	}
}

// Package logger — тонкая обёртка над slog для бэкенда EduHub.
// Не зависит от транспорта (net/http) или драйверов БД (pgx) — чистый
// платформенный строительный блок, доступный любому слою.
package logger

import (
	"io"
	"log/slog"
)

// New создаёт структурированный JSON-логгер поверх w. level — одно из
// "debug"/"info"/"warn"/"error" (те же значения, что валидирует
// config.LogLevel); неизвестное значение трактуется как "info", чтобы New
// никогда не паниковала на некорректном конфиге. Каждая запись содержит
// поле service — имя сервиса, чтобы отличать источник в общих логах.
func New(level string, service string, w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(level),
	})

	return slog.New(handler).With(slog.String("service", service))
}

// parseLevel маппит строковый уровень логирования в slog.Level.
// Неизвестное значение приводится к Info — безопасный дефолт.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// PtrOrNil — единственный разрешённый способ положить значение опционального
// поля (*T) в лог-запись: возвращает nil для nil-указателя вместо паники или
// разыменования непроверенного указателя. Используется вместо *ptr везде,
// где значение может быть отсутствующим (см. .claude/rules/go.md —
// «не разыменовывать *T без nil-проверки»).
func PtrOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

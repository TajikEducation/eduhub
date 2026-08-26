// Package logger — тонкая обёртка над slog для бэкенда EduHub.
package logger

import (
	"io"
	"log/slog"
)

// New создаёт JSON-логгер; неизвестный level трактуется как "info". Каждая запись содержит service.
func New(level string, service string, w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(level),
	})

	return slog.New(handler).With(slog.String("service", service))
}

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

// PtrOrNil — безопасный способ положить *T в лог-запись без разыменования nil.
func PtrOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

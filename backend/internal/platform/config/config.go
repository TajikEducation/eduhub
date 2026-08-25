// Package config загружает конфигурацию бэкенда EduHub из переменных
// окружения. Не зависит от транспорта (net/http) или драйверов БД (pgx) —
// чистый платформенный строительный блок.
package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultHTTPAddr        = ":8080"
	defaultShutdownTimeout = 10 * time.Second
)

// Config хранит конфигурацию бэкенда EduHub, читается из переменных
// окружения (см. backend/.env.example).
type Config struct {
	AppEnv          string
	HTTPAddr        string
	DatabaseURL     string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load читает конфигурацию из переменных окружения. Возвращает ошибку,
// если обязательная переменная отсутствует или не удалось её распарсить.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      os.Getenv("APP_ENV"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}

	shutdownTimeoutRaw := os.Getenv("SHUTDOWN_TIMEOUT")
	if shutdownTimeoutRaw == "" {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	} else {
		d, err := time.ParseDuration(shutdownTimeoutRaw)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid SHUTDOWN_TIMEOUT %q: %w", shutdownTimeoutRaw, err)
		}
		cfg.ShutdownTimeout = d
	}

	return cfg, nil
}

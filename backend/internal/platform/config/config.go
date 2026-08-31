// Package config загружает конфигурацию бэкенда EduHub из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr        = ":8080"
	defaultShutdownTimeout = 10 * time.Second
)

// Config — конфигурация бэкенда.
type Config struct {
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	RedisAddr          string
	LogLevel           string
	ShutdownTimeout    time.Duration
	CORSAllowedOrigins []string

	// Параметры стоимости argon2id (E2.2) — 0 означает «не задано», composition root
	// (cmd/api/main.go) в этом случае берёт значение из password.DefaultParams. Config
	// намеренно не знает про internal/auth/password (platform не знает про домены).
	ArgonMemoryKiB   uint32
	ArgonIterations  uint32
	ArgonParallelism uint8

	// JWTSecret — ключ подписи access-токенов HS256 (E2.3). Обязателен: у секрета подписи
	// нет безопасного дефолта, в отличие от параметров стоимости хеширования.
	JWTSecret string

	// GoogleClientID — client_id Google OAuth-приложения (аудитория Google ID-токена, E2.4).
	// НЕОБЯЗАТЕЛЬНОЕ поле: пустая строка — осмысленный и безопасный дефолт "фича Google-входа
	// отключена" (эндпоинт POST /auth/oauth/google при этом технически существует, но реально
	// нерабочий — все токены отклоняются верификатором), в отличие от JWTSecret/RedisAddr.
	GoogleClientID string
}

// Load читает конфигурацию из ENV, возвращает ошибку при отсутствии обязательной переменной.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:             os.Getenv("APP_ENV"),
		HTTPAddr:           os.Getenv("HTTP_ADDR"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisAddr:          os.Getenv("REDIS_ADDR"),
		LogLevel:           os.Getenv("LOG_LEVEL"),
		CORSAllowedOrigins: parseCORSAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}

	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("config: REDIS_ADDR is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("config: JWT_SECRET is required")
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

	argonMemoryKiB, err := parseOptionalUint32("ARGON_MEMORY_KIB")
	if err != nil {
		return Config{}, err
	}
	cfg.ArgonMemoryKiB = argonMemoryKiB

	argonIterations, err := parseOptionalUint32("ARGON_ITERATIONS")
	if err != nil {
		return Config{}, err
	}
	cfg.ArgonIterations = argonIterations

	argonParallelism, err := parseOptionalUint8("ARGON_PARALLELISM")
	if err != nil {
		return Config{}, err
	}
	cfg.ArgonParallelism = argonParallelism

	return cfg, nil
}

// parseOptionalUint32 — пустая переменная окружения → 0 («не задано», не ошибка),
// непустая нечисловая → ошибка формата.
func parseOptionalUint32(name string) (uint32, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s %q: %w", name, raw, err)
	}
	return uint32(v), nil
}

// parseOptionalUint8 — та же семантика, что parseOptionalUint32, для 8-битных значений.
func parseOptionalUint8(name string) (uint8, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s %q: %w", name, raw, err)
	}
	return uint8(v), nil
}

// parseCORSAllowedOrigins разбирает CORS_ALLOWED_ORIGINS (список через запятую) в slice origin'ов.
// Пустая переменная → пустой slice: fail-closed, CORS ничего не разрешает по умолчанию.
func parseCORSAllowedOrigins(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return origins
}

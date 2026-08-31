// Package config загружает конфигурацию бэкенда EduHub из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultHTTPAddr        = ":8080"
	defaultShutdownTimeout = 10 * time.Second
	// defaultAccessTTL/defaultRefreshTTL — см. SRS Веха 2, E2.3 (access 15 мин, refresh 30 дней).
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 30 * 24 * time.Hour
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
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
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
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}

	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("config: REDIS_ADDR is required")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("config: JWT_SECRET is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}

	cfg.AccessTokenTTL = defaultAccessTTL
	if raw := os.Getenv("ACCESS_TOKEN_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid ACCESS_TOKEN_TTL %q: %w", raw, err)
		}
		cfg.AccessTokenTTL = d
	}

	cfg.RefreshTokenTTL = defaultRefreshTTL
	if raw := os.Getenv("REFRESH_TOKEN_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid REFRESH_TOKEN_TTL %q: %w", raw, err)
		}
		cfg.RefreshTokenTTL = d
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

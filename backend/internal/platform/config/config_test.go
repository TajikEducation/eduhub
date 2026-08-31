package config_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
)

func TestLoad_AllVarsSet(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.AppEnv != "dev" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "dev")
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}
	if cfg.DatabaseURL != "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable" { //nolint:gosec // тестовый DSN локального dev-контейнера, не секрет
		t.Errorf("DatabaseURL = %q, unexpected", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 15*time.Second)
	}
	if cfg.RedisAddr != "localhost:6380" {
		t.Errorf("RedisAddr = %q, want %q", cfg.RedisAddr, "localhost:6380")
	}
	if cfg.JWTSecret != "test-secret" { //nolint:gosec // тестовый секрет, не реальный
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "test-secret")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing DATABASE_URL, got nil")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error %q does not mention DATABASE_URL", err.Error())
	}
}

func TestLoad_MissingRedisAddr(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing REDIS_ADDR, got nil")
	}
	if !strings.Contains(err.Error(), "REDIS_ADDR") {
		t.Errorf("error %q does not mention REDIS_ADDR", err.Error())
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing JWT_SECRET, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("error %q does not mention JWT_SECRET", err.Error())
	}
}

func TestLoad_GoogleClientIDUnset_IsEmptyAndNotAnError(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.GoogleClientID != "" {
		t.Errorf("GoogleClientID = %q, want пустую строку", cfg.GoogleClientID)
	}
}

func TestLoad_DefaultHTTPAddr(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want default %q", cfg.HTTPAddr, ":8080")
	}
}

func TestLoad_CORSAllowedOriginsParsesCommaSeparated(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://eduhub.tj")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := []string{"http://localhost:3000", "https://eduhub.tj"}
	if !slices.Equal(cfg.CORSAllowedOrigins, want) {
		t.Errorf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
}

func TestLoad_CORSAllowedOriginsUnsetIsEmpty(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Errorf("CORSAllowedOrigins = %v, want empty", cfg.CORSAllowedOrigins)
	}
}

func TestLoad_ArgonParamsUnsetAreZero(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ARGON_MEMORY_KIB", "")
	t.Setenv("ARGON_ITERATIONS", "")
	t.Setenv("ARGON_PARALLELISM", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ArgonMemoryKiB != 0 || cfg.ArgonIterations != 0 || cfg.ArgonParallelism != 0 {
		t.Errorf("Argon* = (%d,%d,%d), want (0,0,0) когда переменные не заданы",
			cfg.ArgonMemoryKiB, cfg.ArgonIterations, cfg.ArgonParallelism)
	}
}

func TestLoad_ArgonParamsParsesSetValues(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ARGON_MEMORY_KIB", "65536")
	t.Setenv("ARGON_ITERATIONS", "3")
	t.Setenv("ARGON_PARALLELISM", "4")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ArgonMemoryKiB != 65536 {
		t.Errorf("ArgonMemoryKiB = %d, want 65536", cfg.ArgonMemoryKiB)
	}
	if cfg.ArgonIterations != 3 {
		t.Errorf("ArgonIterations = %d, want 3", cfg.ArgonIterations)
	}
	if cfg.ArgonParallelism != 4 {
		t.Errorf("ArgonParallelism = %d, want 4", cfg.ArgonParallelism)
	}
}

func TestLoad_ArgonMemoryKiBInvalid_ReturnsError(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ARGON_MEMORY_KIB", "не-число")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() с невалидным ARGON_MEMORY_KIB вернул nil-ошибку")
	}
}

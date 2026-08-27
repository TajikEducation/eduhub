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
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LOG_LEVEL", "info")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing DATABASE_URL, got nil")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error %q does not mention DATABASE_URL", err.Error())
	}
}

func TestLoad_DefaultHTTPAddr(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "postgres://eduhub:eduhub@localhost:5433/eduhub?sslmode=disable")
	t.Setenv("LOG_LEVEL", "info")

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
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Errorf("CORSAllowedOrigins = %v, want empty", cfg.CORSAllowedOrigins)
	}
}

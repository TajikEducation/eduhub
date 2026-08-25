package logger_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// TestNew_WritesJSONWithExpectedKeys проверяет кейс (а): New с уровнем
// "info" пишет структурированный JSON с обязательными ключами
// time/level/msg/service.
func TestNew_WritesJSONWithExpectedKeys(t *testing.T) {
	var buf bytes.Buffer

	log := logger.New("info", "eduhub-test", &buf)
	log.Info("hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, log output = %q", err, buf.String())
	}

	for _, key := range []string{"time", "level", "msg", "service"} {
		if _, ok := entry[key]; !ok {
			t.Errorf("log entry missing key %q, entry = %v", key, entry)
		}
	}

	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want %q", entry["msg"], "hello")
	}
	if entry["level"] != "INFO" {
		t.Errorf("level = %v, want %q", entry["level"], "INFO")
	}
	if entry["service"] != "eduhub-test" {
		t.Errorf("service = %v, want %q", entry["service"], "eduhub-test")
	}
}

// TestNew_DebugSuppressedAtInfoLevel проверяет кейс (б): при уровне "info"
// вызов Debug ничего не пишет в вывод.
func TestNew_DebugSuppressedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer

	log := logger.New("info", "eduhub-test", &buf)
	log.Debug("should not appear")

	if buf.Len() != 0 {
		t.Errorf("buf.Len() = %d, want 0; output = %q", buf.Len(), buf.String())
	}
}

// TestPtrOrNil проверяет кейс (в): PtrOrNil возвращает nil для nil-указателя
// и значение — для непустого, без паники разыменования.
func TestPtrOrNil(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var p *string
		got := logger.PtrOrNil(p)
		if got != nil {
			t.Errorf("PtrOrNil(nil) = %v, want nil", got)
		}
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		v := "phone-hidden"
		got := logger.PtrOrNil(&v)
		if got != "phone-hidden" {
			t.Errorf("PtrOrNil(&v) = %v, want %q", got, "phone-hidden")
		}
	})

	t.Run("non-nil pointer to int", func(t *testing.T) {
		v := 42
		got := logger.PtrOrNil(&v)
		if got != 42 {
			t.Errorf("PtrOrNil(&v) = %v, want %d", got, 42)
		}
	})
}

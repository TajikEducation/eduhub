package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// TestNotFound_MatchesSentinel — RED-кейс (а): apperr.NotFound(...) должен
// матчиться через errors.Is(err, apperr.ErrNotFound).
func TestNotFound_MatchesSentinel(t *testing.T) {
	err := apperr.NotFound("institution", "42")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(err, apperr.ErrNotFound) = false, want true; err=%v", err)
	}
}

// TestNotFound_MatchesSentinel_ThroughWrap — RED-кейс (б): обёрнутая через
// fmt.Errorf("repo: %w", err) ошибка тоже должна матчиться через errors.Is —
// не только прямое сравнение.
func TestNotFound_MatchesSentinel_ThroughWrap(t *testing.T) {
	inner := apperr.NotFound("institution", "42")
	wrapped := fmt.Errorf("repo: %w", inner)

	if !errors.Is(wrapped, apperr.ErrNotFound) {
		t.Fatalf("errors.Is(wrapped, apperr.ErrNotFound) = false, want true; wrapped=%v", wrapped)
	}
}

// TestInvalid_FieldsThroughAs — RED-кейс (в): apperr.Invalid несёт
// map[string]string полей, извлекаемый типизированно через errors.As.
func TestInvalid_FieldsThroughAs(t *testing.T) {
	fields := map[string]string{"min_price": "must be non-negative"}
	err := apperr.Invalid(fields, "validation failed")

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err=%v", err)
	}
	if target.Fields["min_price"] != "must be non-negative" {
		t.Errorf("target.Fields[%q] = %q, want %q", "min_price", target.Fields["min_price"], "must be non-negative")
	}
}

// TestInvalid_FieldsThroughAs_ThroughWrap — errors.As тоже должен работать
// сквозь обёртку fmt.Errorf("...: %w", err).
func TestInvalid_FieldsThroughAs_ThroughWrap(t *testing.T) {
	fields := map[string]string{"email": "required"}
	inner := apperr.Invalid(fields, "validation failed")
	wrapped := fmt.Errorf("handler: %w", inner)

	var target *apperr.Error
	if !errors.As(wrapped, &target) {
		t.Fatalf("errors.As(wrapped, &target) = false, want true; wrapped=%v", wrapped)
	}
	if target.Fields["email"] != "required" {
		t.Errorf("target.Fields[%q] = %q, want %q", "email", target.Fields["email"], "required")
	}
	if !errors.Is(wrapped, apperr.ErrInvalid) {
		t.Errorf("errors.Is(wrapped, apperr.ErrInvalid) = false, want true")
	}
}

// TestCategories_MatchSentinels — по одному тесту на каждую из оставшихся
// пяти категорий (Unauthorized/Forbidden/Conflict/RateLimited/Internal):
// сравнение через errors.Is с соответствующим sentinel, в том числе через
// обёртку fmt.Errorf("...: %w", err).
func TestCategories_MatchSentinels(t *testing.T) {
	cause := errors.New("connection refused")

	tests := []struct {
		name     string
		err      error
		sentinel error
	}{
		{"Unauthorized", apperr.Unauthorized("токен истёк"), apperr.ErrUnauthorized},
		{"Forbidden", apperr.Forbidden("недостаточно прав"), apperr.ErrForbidden},
		{"Conflict", apperr.Conflict("отзыв уже существует"), apperr.ErrConflict},
		{"RateLimited", apperr.RateLimited("слишком много попыток"), apperr.ErrRateLimited},
		{"Internal", apperr.Internal(cause), apperr.ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.sentinel) {
				t.Errorf("errors.Is(err, sentinel) = false, want true; err=%v", tt.err)
			}

			wrapped := fmt.Errorf("usecase: %w", tt.err)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("errors.Is(wrapped, sentinel) = false, want true; wrapped=%v", wrapped)
			}

			// Разные категории не должны матчиться друг с другом.
			if tt.sentinel != apperr.ErrNotFound && errors.Is(tt.err, apperr.ErrNotFound) {
				t.Errorf("errors.Is(err, apperr.ErrNotFound) = true, want false для категории %s", tt.name)
			}
		})
	}
}

// TestInternal_UnwrapsCause — Internal должен оборачивать исходную
// техническую ошибку так, чтобы её тоже можно было достать через
// errors.Is/errors.Unwrap — для логирования (не для показа пользователю).
func TestInternal_UnwrapsCause(t *testing.T) {
	cause := errors.New("dial tcp: connection refused")
	err := apperr.Internal(cause)

	if !errors.Is(err, cause) {
		t.Errorf("errors.Is(err, cause) = false, want true; err=%v", err)
	}
	if !errors.Is(err, apperr.ErrInternal) {
		t.Errorf("errors.Is(err, apperr.ErrInternal) = false, want true; err=%v", err)
	}
}

// TestError_MessageNotEmpty — базовая проверка, что Error() возвращает
// непустую строку с полезным контекстом (не голое имя категории).
func TestError_MessageNotEmpty(t *testing.T) {
	err := apperr.NotFound("institution", "42")
	if err.Error() == "" {
		t.Fatal("Error() returned empty string")
	}
}

// TestMessage_ReturnsMessageWithoutCategoryPrefix — RED-кейс: Message()
// возвращает только текст сообщения, без дублирования префикса категории,
// который Error() добавляет для errors.Is-совместимого текста.
func TestMessage_ReturnsMessageWithoutCategoryPrefix(t *testing.T) {
	err := apperr.NotFound("institution", "42")

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err=%v", err)
	}

	want := "institution not found: id=42"
	if got := target.Message(); got != want {
		t.Errorf("Message() = %q, want %q", got, want)
	}
}

// TestMessage_FallsBackToCategoryWhenEmpty — RED-кейс: если message
// пустой (например, Unauthorized("") — вызывающий не указал деталей),
// Message() возвращает текст категории, а не пустую строку.
func TestMessage_FallsBackToCategoryWhenEmpty(t *testing.T) {
	err := apperr.Unauthorized("")

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true; err=%v", err)
	}

	want := apperr.ErrUnauthorized.Error()
	if got := target.Message(); got != want {
		t.Errorf("Message() = %q, want %q", got, want)
	}
}

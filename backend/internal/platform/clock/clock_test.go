package clock_test

import (
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
)

// TestNew_ReturnsRealTime проверяет, что New().Now() близок к time.Now() (реальные часы, без подмены).
func TestNew_ReturnsRealTime(t *testing.T) {
	got := clock.New().Now()

	if diff := time.Since(got); diff < 0 || diff > time.Second {
		t.Errorf("clock.New().Now() = %v, differs from time.Now() by %v, want < 1s", got, diff)
	}
}

// TestFake_NowReturnsStart проверяет, что свежесозданный Fake отдаёт стартовое время как есть.
func TestFake_NowReturnsStart(t *testing.T) {
	start := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	f := clock.NewFake(start)

	if got := f.Now(); !got.Equal(start) {
		t.Errorf("Now() = %v, want %v", got, start)
	}
}

// TestFake_AdvanceMovesNow проверяет, что Advance сдвигает время, возвращаемое Now(), на заданную длительность.
func TestFake_AdvanceMovesNow(t *testing.T) {
	start := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	f := clock.NewFake(start)

	f.Advance(time.Minute)

	want := start.Add(time.Minute)
	if got := f.Now(); !got.Equal(want) {
		t.Errorf("Now() after Advance(1m) = %v, want %v", got, want)
	}
}

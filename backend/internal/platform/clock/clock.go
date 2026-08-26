// Package clock — инжектируемые часы: реальные для прода, управляемые вручную для тестов
// (rate-limiter должен двигать время без time.Sleep).
package clock

import (
	"sync"
	"time"
)

// Clock абстрагирует источник текущего времени.
type Clock interface {
	Now() time.Time
}

// realClock — Clock поверх time.Now().
type realClock struct{}

// New создаёт Clock, возвращающий реальное текущее время.
func New() Clock {
	return realClock{}
}

func (realClock) Now() time.Time {
	return time.Now()
}

// Fake — тестовый Clock с ручным продвижением времени; безопасен для конкурентного использования
// (rate-limiter читает Now() из параллельно обслуживаемых HTTP-хендлеров).
type Fake struct {
	mu  sync.Mutex
	now time.Time
}

// NewFake создаёт Fake, зафиксированный на start.
func NewFake(start time.Time) *Fake {
	return &Fake{now: start}
}

// Now возвращает текущее (зафиксированное) время часов.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Advance сдвигает часы вперёд на d.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

package usage

import (
	"sync"
	"time"
)

type Limiter interface {
	Allow(subject string, now time.Time) bool
}

type FixedWindow struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]entry
}
type entry struct {
	started time.Time
	count   int
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{limit: limit, window: window, entries: make(map[string]entry)}
}
func (l *FixedWindow) Allow(subject string, now time.Time) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	current := l.entries[subject]
	if current.started.IsZero() || now.Sub(current.started) >= l.window {
		current = entry{started: now}
	}
	if current.count >= l.limit {
		l.entries[subject] = current
		return false
	}
	current.count++
	l.entries[subject] = current
	return true
}

package usage

import (
	"testing"
	"time"
)

func TestFixedWindow(t *testing.T) {
	l := NewFixedWindow(2, time.Minute)
	now := time.Now()
	if !l.Allow("key-1", now) || !l.Allow("key-1", now) || l.Allow("key-1", now) {
		t.Fatal("window limit was not enforced")
	}
	if !l.Allow("key-1", now.Add(time.Minute)) {
		t.Fatal("new window was not allowed")
	}
}

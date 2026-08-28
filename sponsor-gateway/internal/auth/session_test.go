package auth

import (
	"testing"
	"time"
)

func TestSessionCannotBeTamperedOrReusedAfterExpiry(t *testing.T) {
	sessions, err := NewSessions("01234567890123456789012345678901", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	token, err := sessions.Issue("user-1", now)
	if err != nil {
		t.Fatal(err)
	}
	id, err := sessions.Validate(token, now)
	if err != nil || id != "user-1" {
		t.Fatal("valid session rejected")
	}
	if _, err = sessions.Validate(token+"x", now); err == nil {
		t.Fatal("tampered session accepted")
	}
	if _, err = sessions.Validate(token, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expired session accepted")
	}
}

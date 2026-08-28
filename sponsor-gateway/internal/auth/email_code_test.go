package auth

import (
	"testing"
	"time"
)

func TestEmailCodeSingleUseAndAttempts(t *testing.T) {
	codes, err := NewEmailCodes("01234567890123456789012345678901", time.Minute, 2)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	code, err := codes.Issue("user@example.com", now)
	if err != nil {
		t.Fatal(err)
	}
	if codes.Verify("user@example.com", "000000", now) {
		t.Fatal("wrong code accepted")
	}
	if !codes.Verify("user@example.com", code, now) {
		t.Fatal("correct code rejected")
	}
	if codes.Verify("user@example.com", code, now) {
		t.Fatal("code was reused")
	}
}

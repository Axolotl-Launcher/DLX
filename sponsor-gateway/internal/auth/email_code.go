package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EmailCodes struct {
	mu          sync.Mutex
	pepper      []byte
	ttl         time.Duration
	maxAttempts int
	records     map[string]emailCode
}
type emailCode struct {
	digest    string
	expiresAt time.Time
	attempts  int
}

func NewEmailCodes(pepper string, ttl time.Duration, maxAttempts int) (*EmailCodes, error) {
	if len(pepper) < 32 {
		return nil, errors.New("email code pepper must be at least 32 bytes")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return &EmailCodes{pepper: []byte(pepper), ttl: ttl, maxAttempts: maxAttempts, records: map[string]emailCode{}}, nil
}
func (c *EmailCodes) Issue(email string, now time.Time) (string, error) {
	email = normalizeEmail(email)
	number := make([]byte, 4)
	if _, err := rand.Read(number); err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", (uint32(number[0])<<24|uint32(number[1])<<16|uint32(number[2])<<8|uint32(number[3]))%1000000)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records[email] = emailCode{digest: c.digest(email, code), expiresAt: now.Add(c.ttl)}
	return code, nil
}
func (c *EmailCodes) Verify(email, code string, now time.Time) bool {
	email = normalizeEmail(email)
	c.mu.Lock()
	defer c.mu.Unlock()
	record, ok := c.records[email]
	if !ok || now.After(record.expiresAt) || record.attempts >= c.maxAttempts {
		delete(c.records, email)
		return false
	}
	record.attempts++
	if hmac.Equal([]byte(record.digest), []byte(c.digest(email, code))) {
		delete(c.records, email)
		return true
	}
	c.records[email] = record
	return false
}
func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (c *EmailCodes) digest(email, code string) string {
	mac := hmac.New(sha256.New, c.pepper)
	_, _ = mac.Write([]byte(email + "\x00" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

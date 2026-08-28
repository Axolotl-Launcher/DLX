package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Sessions struct {
	secret []byte
	ttl    time.Duration
}
type sessionClaims struct {
	Subject   string `json:"sub"`
	ExpiresAt int64  `json:"exp"`
}

func NewSessions(secret string, ttl time.Duration) (*Sessions, error) {
	if len(secret) < 32 {
		return nil, errors.New("session secret must be at least 32 bytes")
	}
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	return &Sessions{secret: []byte(secret), ttl: ttl}, nil
}
func (s *Sessions) Issue(userID string, now time.Time) (string, error) {
	if userID == "" {
		return "", errors.New("user id is required")
	}
	payload, err := json.Marshal(sessionClaims{Subject: userID, ExpiresAt: now.Add(s.ttl).Unix()})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + s.sign(encoded), nil
}
func (s *Sessions) Validate(token string, now time.Time) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || !hmac.Equal([]byte(s.sign(parts[0])), []byte(parts[1])) {
		return "", errors.New("invalid session")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("invalid session")
	}
	var claims sessionClaims
	if json.Unmarshal(payload, &claims) != nil || claims.Subject == "" || now.Unix() > claims.ExpiresAt {
		return "", errors.New("expired session")
	}
	return claims.Subject, nil
}
func (s *Sessions) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

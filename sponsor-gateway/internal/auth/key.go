package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const keyPrefix = "axl_live_"

type Hasher struct{ pepper []byte }

func NewHasher(pepper string) (*Hasher, error) {
	if len(pepper) < 32 {
		return nil, errors.New("API key pepper must be at least 32 bytes")
	}
	return &Hasher{pepper: []byte(pepper)}, nil
}
func (h *Hasher) Hash(key string) string {
	mac := hmac.New(sha256.New, h.pepper)
	_, _ = mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}
func (h *Hasher) Equal(stored, key string) bool {
	return hmac.Equal([]byte(stored), []byte(h.Hash(key)))
}
func NewAPIKey(id string, hasher *Hasher) (string, string, error) {
	if hasher == nil {
		return "", "", errors.New("key hasher is required")
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", "", err
	}
	plaintext := fmt.Sprintf("%s%s_%s", keyPrefix, id, hex.EncodeToString(secret))
	return plaintext, hasher.Hash(plaintext), nil
}
func ParseBearer(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !strings.HasPrefix(parts[1], keyPrefix) || len(parts[1]) < 50 {
		return "", errors.New("invalid bearer token")
	}
	return parts[1], nil
}

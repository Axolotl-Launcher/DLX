package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
)

const recoveryCodePrefix = "axl_login_"

// NewRecoveryCode creates the only long-term user credential in the simplified
// flow. The plaintext is returned once; callers must persist only its HMAC.
func NewRecoveryCode(hasher *Hasher) (string, string, error) {
	if hasher == nil {
		return "", "", errors.New("code hasher is required")
	}
	secret := make([]byte, 24)
	if _, err := rand.Read(secret); err != nil {
		return "", "", err
	}
	code := recoveryCodePrefix + base64.RawURLEncoding.EncodeToString(secret)
	return code, hasher.Hash("recovery:" + code), nil
}
func ParseRecoveryCode(code string) (string, error) {
	code = strings.TrimSpace(code)
	if !strings.HasPrefix(code, recoveryCodePrefix) || len(code) < len(recoveryCodePrefix)+32 {
		return "", errors.New("invalid login code")
	}
	if _, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(code, recoveryCodePrefix)); err != nil {
		return "", errors.New("invalid login code")
	}
	return code, nil
}
func VerifyRecoveryCode(hasher *Hasher, stored, code string) bool {
	return hasher != nil && hasher.Equal(stored, "recovery:"+code)
}

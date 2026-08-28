package auth

import "testing"

func TestAPIKeyHashAndBearer(t *testing.T) {
	hasher, err := NewHasher("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	key, digest, err := NewAPIKey("key-1", hasher)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseBearer("Bearer " + key)
	if err != nil || parsed != key {
		t.Fatalf("parse bearer: %v", err)
	}
	if !hasher.Equal(digest, key) {
		t.Fatal("HMAC digest did not verify")
	}
	if _, err := ParseBearer("Bearer " + key + " extra"); err == nil {
		t.Fatal("accepted malformed bearer")
	}
}

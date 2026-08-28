package auth

import "testing"

func TestRecoveryCodeIsOneTimeDisplayMaterial(t *testing.T) {
	h, err := NewHasher("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	code, digest, err := NewRecoveryCode(h)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseRecoveryCode(code)
	if err != nil || parsed != code {
		t.Fatal("valid code did not parse")
	}
	if !VerifyRecoveryCode(h, digest, code) {
		t.Fatal("stored digest did not verify")
	}
	if VerifyRecoveryCode(h, digest, code+"x") {
		t.Fatal("altered recovery code verified")
	}
}

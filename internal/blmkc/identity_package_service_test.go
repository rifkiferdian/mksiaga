package blmkc

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestEncryptNIKUsesRandomCiphertextAndStableHash(t *testing.T) {
	const nik = "3401010101010001"
	const key = "test-key-with-at-least-thirty-two-characters"
	first, firstHash, err := encryptNIK(nik, key)
	if err != nil {
		t.Fatal(err)
	}
	second, secondHash, err := encryptNIK(nik, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) == nik || bytes.Equal(first, second) {
		t.Fatal("NIK ciphertext must not contain plaintext and must use a random nonce")
	}
	if !bytes.Equal(firstHash, secondHash) || len(firstHash) != 32 {
		t.Fatal("NIK hash must be stable SHA-256 output")
	}
	if _, err := base64.StdEncoding.DecodeString(string(first)); err != nil {
		t.Fatalf("ciphertext must be database-safe base64: %v", err)
	}
}

func TestIdentityPackageReturnToOnlyAllowsInternalDetailPages(t *testing.T) {
	tests := map[string]string{
		"/blmkc/identity/8":           "/blmkc/identity/8",
		"/blmkc/cases/12":             "/blmkc/cases/12",
		"https://example.com":         "/blmkc/cases/7",
		"//example.com":               "/blmkc/cases/7",
		"/blmkc/cases/12/../../admin": "/blmkc/cases/7",
	}
	for input, want := range tests {
		if got := identityPackageReturnTo(input, 7); got != want {
			t.Errorf("identityPackageReturnTo(%q) = %q, want %q", input, got, want)
		}
	}
}

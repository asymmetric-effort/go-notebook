// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	passphrase := []byte("test-passphrase-long-enough")
	masterKey := DeriveKey(passphrase, salt[:])
	payloadKey, err := DerivePayloadKey(masterKey, "test-payload")
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("secret data to encrypt")
	ciphertext, err := Encrypt(payloadKey, plaintext, "test-payload")
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := Decrypt(payloadKey, ciphertext, "test-payload")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestWrongPassphrase(t *testing.T) {
	salt, _ := GenerateSalt()
	key1 := DeriveKey([]byte("correct-passphrase!!"), salt[:])
	key2 := DeriveKey([]byte("wrong-passphrase!!!!"), salt[:])

	payloadKey1, _ := DerivePayloadKey(key1, "test")
	payloadKey2, _ := DerivePayloadKey(key2, "test")

	ciphertext, err := Encrypt(payloadKey1, []byte("secret"), "test")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(payloadKey2, ciphertext, "test")
	if err == nil {
		t.Error("expected error with wrong passphrase")
	}
}

func TestWrongPayloadName(t *testing.T) {
	salt, _ := GenerateSalt()
	key := DeriveKey([]byte("test-passphrase-long-enough"), salt[:])
	payloadKey, _ := DerivePayloadKey(key, "payload-a")

	ciphertext, err := Encrypt(payloadKey, []byte("secret"), "payload-a")
	if err != nil {
		t.Fatal(err)
	}

	// Trying to decrypt with wrong payload name should fail (AAD mismatch).
	_, err = Decrypt(payloadKey, ciphertext, "payload-b")
	if err == nil {
		t.Error("expected error with wrong payload name (swapped payloads)")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := make([]byte, 32)
	_, err := Decrypt(key, []byte("short"), "test")
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}

func TestDecryptTampered(t *testing.T) {
	salt, _ := GenerateSalt()
	key := DeriveKey([]byte("test-passphrase-long-enough"), salt[:])
	payloadKey, _ := DerivePayloadKey(key, "test")

	ciphertext, _ := Encrypt(payloadKey, []byte("secret"), "test")

	// Flip a bit.
	ciphertext[len(ciphertext)-1] ^= 0x01

	_, err := Decrypt(payloadKey, ciphertext, "test")
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if salt1 == salt2 {
		t.Error("two salts should not be equal")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := [SaltLen]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	key1 := DeriveKey([]byte("passphrase"), salt[:])
	key2 := DeriveKey([]byte("passphrase"), salt[:])
	if !bytes.Equal(key1, key2) {
		t.Error("same passphrase+salt should produce same key")
	}
}

func TestDerivePayloadKeyDifferentNames(t *testing.T) {
	masterKey := make([]byte, 32)
	key1, _ := DerivePayloadKey(masterKey, "payload-a")
	key2, _ := DerivePayloadKey(masterKey, "payload-b")
	if bytes.Equal(key1, key2) {
		t.Error("different payload names should produce different keys")
	}
}

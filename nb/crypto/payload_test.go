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

func TestHeaderMarshalRoundTrip(t *testing.T) {
	salt := [SaltLen]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	h := Header{
		Version:       FormatVersion,
		Argon2Time:    Argon2Time,
		Argon2Memory:  Argon2Memory,
		Argon2Threads: Argon2Threads,
		Salt:          salt,
	}
	data := h.MarshalBinary()
	if len(data) != HeaderSize {
		t.Fatalf("header size = %d, want %d", len(data), HeaderSize)
	}

	h2, err := UnmarshalHeader(data)
	if err != nil {
		t.Fatal(err)
	}
	if h2.Version != h.Version {
		t.Errorf("version = %d, want %d", h2.Version, h.Version)
	}
	if h2.Argon2Time != h.Argon2Time {
		t.Errorf("argon2 time = %d, want %d", h2.Argon2Time, h.Argon2Time)
	}
	if h2.Argon2Memory != h.Argon2Memory {
		t.Errorf("argon2 memory = %d, want %d", h2.Argon2Memory, h.Argon2Memory)
	}
	if h2.Argon2Threads != h.Argon2Threads {
		t.Errorf("argon2 threads = %d, want %d", h2.Argon2Threads, h.Argon2Threads)
	}
	if h2.Salt != h.Salt {
		t.Error("salt mismatch")
	}
}

func TestUnmarshalHeaderTooShort(t *testing.T) {
	_, err := UnmarshalHeader([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for short header")
	}
}

func TestEncryptDecryptPayload(t *testing.T) {
	passphrase := []byte("a-long-enough-passphrase")
	plaintext := []byte("secret payload data")

	encrypted, err := EncryptPayload(passphrase, plaintext, "test-payload")
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := DecryptPayload(passphrase, encrypted, "test-payload")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptPayloadWrongPassphrase(t *testing.T) {
	encrypted, err := EncryptPayload([]byte("correct-passphrase!!"), []byte("secret"), "test")
	if err != nil {
		t.Fatal(err)
	}

	_, err = DecryptPayload([]byte("wrong-passphrase!!!!"), encrypted, "test")
	if err == nil {
		t.Error("expected error with wrong passphrase")
	}
}

func TestDecryptPayloadWrongName(t *testing.T) {
	encrypted, err := EncryptPayload([]byte("test-passphrase-long-enough"), []byte("secret"), "payload-a")
	if err != nil {
		t.Fatal(err)
	}

	_, err = DecryptPayload([]byte("test-passphrase-long-enough"), encrypted, "payload-b")
	if err == nil {
		t.Error("expected error with wrong payload name")
	}
}

func TestDecryptPayloadTampered(t *testing.T) {
	encrypted, err := EncryptPayload([]byte("test-passphrase-long-enough"), []byte("secret"), "test")
	if err != nil {
		t.Fatal(err)
	}

	// Flip a bit in the ciphertext (after header).
	encrypted[len(encrypted)-1] ^= 0x01

	_, err = DecryptPayload([]byte("test-passphrase-long-enough"), encrypted, "test")
	if err == nil {
		t.Error("expected error for tampered payload")
	}
}

func TestDecryptPayloadBadVersion(t *testing.T) {
	encrypted, err := EncryptPayload([]byte("test-passphrase-long-enough"), []byte("secret"), "test")
	if err != nil {
		t.Fatal(err)
	}

	// Change the version byte.
	encrypted[0] = 99

	_, err = DecryptPayload([]byte("test-passphrase-long-enough"), encrypted, "test")
	if err == nil {
		t.Error("expected error for bad version")
	}
}

func FuzzDecryptPayload(f *testing.F) {
	// Seed with a valid encrypted payload.
	encrypted, _ := EncryptPayload([]byte("test-passphrase-long-enough"), []byte("hello"), "test")
	f.Add(encrypted)
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})

	f.Fuzz(func(t *testing.T, data []byte) {
		// DecryptPayload should never panic, only return errors.
		_, _ = DecryptPayload([]byte("test-passphrase-long-enough"), data, "test")
	})
}

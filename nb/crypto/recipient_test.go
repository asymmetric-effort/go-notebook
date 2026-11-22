// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"crypto/hpke"
	"testing"
)

func generateXWingKey(t *testing.T) (hpke.PrivateKey, hpke.PublicKey) {
	t.Helper()
	kem := hpke.MLKEM768X25519()
	priv, err := kem.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return priv, priv.PublicKey()
}

func TestEncryptDecryptOneRecipient(t *testing.T) {
	priv, pub := generateXWingKey(t)
	plaintext := []byte("secret data for recipient")

	ciphertext, headers, err := EncryptForRecipients(plaintext, []hpke.PublicKey{pub}, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 1 {
		t.Fatalf("got %d headers, want 1", len(headers))
	}

	decrypted, err := DecryptForRecipient(ciphertext, headers, priv, "test")
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecryptMultipleRecipients(t *testing.T) {
	priv1, pub1 := generateXWingKey(t)
	priv2, pub2 := generateXWingKey(t)

	plaintext := []byte("shared secret")
	ciphertext, headers, err := EncryptForRecipients(plaintext, []hpke.PublicKey{pub1, pub2}, "multi")
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 2 {
		t.Fatalf("got %d headers, want 2", len(headers))
	}

	// Both recipients should be able to decrypt.
	dec1, err := DecryptForRecipient(ciphertext, headers, priv1, "multi")
	if err != nil {
		t.Fatal(err)
	}
	if string(dec1) != string(plaintext) {
		t.Error("recipient 1 decryption failed")
	}

	dec2, err := DecryptForRecipient(ciphertext, headers, priv2, "multi")
	if err != nil {
		t.Fatal(err)
	}
	if string(dec2) != string(plaintext) {
		t.Error("recipient 2 decryption failed")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	_, pub := generateXWingKey(t)
	wrongPriv, _ := generateXWingKey(t)

	ciphertext, headers, err := EncryptForRecipients([]byte("secret"), []hpke.PublicKey{pub}, "test")
	if err != nil {
		t.Fatal(err)
	}

	_, err = DecryptForRecipient(ciphertext, headers, wrongPriv, "test")
	if err == nil {
		t.Error("expected error with wrong key")
	}
}

func TestEncryptNoRecipients(t *testing.T) {
	_, _, err := EncryptForRecipients([]byte("data"), nil, "test")
	if err == nil {
		t.Error("expected error with no recipients")
	}
}

func TestMarshalUnmarshalHeaders(t *testing.T) {
	_, pub := generateXWingKey(t)
	_, headers, err := EncryptForRecipients([]byte("data"), []hpke.PublicKey{pub}, "test")
	if err != nil {
		t.Fatal(err)
	}

	data := MarshalRecipientHeaders(headers)
	restored, consumed, err := UnmarshalRecipientHeaders(data)
	if err != nil {
		t.Fatal(err)
	}
	if consumed != len(data) {
		t.Errorf("consumed %d bytes, want %d", consumed, len(data))
	}
	if len(restored) != len(headers) {
		t.Fatalf("got %d headers, want %d", len(restored), len(headers))
	}
	if len(restored[0].EncryptedKey) != len(headers[0].EncryptedKey) {
		t.Error("key length mismatch")
	}
}

func TestUnmarshalHeadersTooShort(t *testing.T) {
	_, _, err := UnmarshalRecipientHeaders([]byte{0})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestUnmarshalHeadersTruncated(t *testing.T) {
	// Valid count but truncated key data.
	data := []byte{0, 1, 0, 100} // 1 header, key length 100 but no data.
	_, _, err := UnmarshalRecipientHeaders(data)
	if err == nil {
		t.Error("expected error for truncated key")
	}
}

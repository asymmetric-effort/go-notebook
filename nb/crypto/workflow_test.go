// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"bytes"
	"crypto/hpke"
	"testing"
)

// TestEndToEndPassphraseWorkflow exercises the full passphrase-mode workflow:
// generate salt → derive key → encrypt payload → decrypt payload.
func TestEndToEndPassphraseWorkflow(t *testing.T) {
	passphrase := []byte("a-secure-passphrase-at-least-12")
	plaintext := []byte("sensitive report data with charts and cell outputs")
	payloadName := "report.html"

	// Encrypt.
	encrypted, err := EncryptPayload(passphrase, plaintext, payloadName)
	if err != nil {
		t.Fatal(err)
	}

	// Verify encrypted data is not plaintext.
	if bytes.Contains(encrypted, plaintext) {
		t.Error("encrypted data should not contain plaintext")
	}

	// Decrypt.
	decrypted, err := DecryptPayload(passphrase, encrypted, payloadName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("decrypted data should match original")
	}
}

// TestEndToEndRecipientWorkflow exercises the full recipient-mode workflow:
// generate keys → encrypt for recipients → decrypt by recipient.
func TestEndToEndRecipientWorkflow(t *testing.T) {
	kem := hpke.MLKEM768X25519()

	// Generate two recipients.
	priv1, err := kem.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	priv2, err := kem.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("report data for multiple recipients")
	payloadName := "report.html"

	// Encrypt for both recipients.
	ciphertext, headers, err := EncryptForRecipients(
		plaintext,
		[]hpke.PublicKey{priv1.PublicKey(), priv2.PublicKey()},
		payloadName,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Serialize and deserialize headers (simulating storage).
	headerBytes := MarshalRecipientHeaders(headers)
	restoredHeaders, _, err := UnmarshalRecipientHeaders(headerBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Recipient 1 decrypts.
	dec1, err := DecryptForRecipient(ciphertext, restoredHeaders, priv1, payloadName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, dec1) {
		t.Error("recipient 1 decryption failed")
	}

	// Recipient 2 decrypts.
	dec2, err := DecryptForRecipient(ciphertext, restoredHeaders, priv2, payloadName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, dec2) {
		t.Error("recipient 2 decryption failed")
	}
}

// TestTamperPassphrasePayload verifies every tamper case fails.
func TestTamperPassphrasePayload(t *testing.T) {
	passphrase := []byte("a-secure-passphrase-at-least-12")
	plaintext := []byte("secret data")

	encrypted, err := EncryptPayload(passphrase, plaintext, "test")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		modify func([]byte) []byte
	}{
		{"flip first byte", func(d []byte) []byte {
			c := make([]byte, len(d))
			copy(c, d)
			c[0] ^= 0xFF
			return c
		}},
		{"flip last byte", func(d []byte) []byte {
			c := make([]byte, len(d))
			copy(c, d)
			c[len(c)-1] ^= 0x01
			return c
		}},
		{"flip middle byte", func(d []byte) []byte {
			c := make([]byte, len(d))
			copy(c, d)
			c[len(c)/2] ^= 0x42
			return c
		}},
		{"truncate", func(d []byte) []byte {
			return d[:len(d)-10]
		}},
		{"empty", func(d []byte) []byte {
			return []byte{}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tampered := tt.modify(encrypted)
			_, err := DecryptPayload(passphrase, tampered, "test")
			if err == nil {
				t.Errorf("tamper case %q should fail", tt.name)
			}
		})
	}
}

// TestSwapPayloadName verifies swapping the payload name fails (AAD mismatch).
func TestSwapPayloadName(t *testing.T) {
	passphrase := []byte("a-secure-passphrase-at-least-12")
	encrypted, err := EncryptPayload(passphrase, []byte("data"), "payload-a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptPayload(passphrase, encrypted, "payload-b")
	if err == nil {
		t.Error("swapped payload name should fail")
	}
}

// TestWrongPassphraseFails verifies the wrong passphrase fails.
func TestWrongPassphraseFails(t *testing.T) {
	encrypted, err := EncryptPayload([]byte("correct-passphrase!!"), []byte("data"), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptPayload([]byte("wrong-passphrase!!!!"), encrypted, "test")
	if err == nil {
		t.Error("wrong passphrase should fail")
	}
}

// TestWrongRecipientKeyFails verifies the wrong recipient key fails.
func TestWrongRecipientKeyFails(t *testing.T) {
	kem := hpke.MLKEM768X25519()
	_, pub := generateKey(t, kem)
	wrongPriv, _ := generateKey(t, kem)

	ciphertext, headers, err := EncryptForRecipients([]byte("data"), []hpke.PublicKey{pub}, "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptForRecipient(ciphertext, headers, wrongPriv, "test")
	if err == nil {
		t.Error("wrong recipient key should fail")
	}
}

func generateKey(t *testing.T, kem hpke.KEM) (hpke.PrivateKey, hpke.PublicKey) {
	t.Helper()
	priv, err := kem.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return priv, priv.PublicKey()
}

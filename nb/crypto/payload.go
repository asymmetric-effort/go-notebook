// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package crypto implements the encryption format for embedded data and
// secrets in report binaries.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

// FormatVersion identifies the encryption format for forward compatibility.
const FormatVersion = 1

// Argon2id parameters (RFC 9106 lower-memory option).
const (
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024 // 64 MiB
	Argon2Threads = 4
	Argon2KeyLen  = 32
	SaltLen       = 16
	NonceLen      = 12
)

// Header holds the encryption metadata stored with the payload.
type Header struct {
	Version       uint8
	Argon2Time    uint32
	Argon2Memory  uint32
	Argon2Threads uint8
	Salt          [SaltLen]byte
}

// DeriveKey derives a master key from a passphrase and salt using Argon2id.
func DeriveKey(passphrase []byte, salt []byte) []byte {
	return argon2.IDKey(passphrase, salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
}

// DerivePayloadKey derives a per-payload key from the master key and payload name.
func DerivePayloadKey(masterKey []byte, payloadName string) ([]byte, error) {
	info := []byte("gobook-payload-" + payloadName)
	hk := hkdf.New(sha256.New, masterKey, nil, info)
	key := make([]byte, Argon2KeyLen)
	if _, err := io.ReadFull(hk, key); err != nil {
		return nil, fmt.Errorf("derive payload key: %w", err)
	}
	return key, nil
}

// Encrypt seals data with AES-256-GCM. The payloadName is included as
// authenticated data so payloads cannot be swapped undetected.
func Encrypt(key []byte, plaintext []byte, payloadName string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	nonce := make([]byte, NonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	ad := authenticatedData(payloadName)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, ad)
	return ciphertext, nil
}

// Decrypt opens data sealed with Encrypt.
func Decrypt(key []byte, ciphertext []byte, payloadName string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	if len(ciphertext) < NonceLen {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:NonceLen]
	data := ciphertext[NonceLen:]
	ad := authenticatedData(payloadName)
	plaintext, err := gcm.Open(nil, nonce, data, ad)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// GenerateSalt creates a random salt for Argon2id.
func GenerateSalt() ([SaltLen]byte, error) {
	var salt [SaltLen]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return salt, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

// authenticatedData creates the AAD for AES-256-GCM.
func authenticatedData(payloadName string) []byte {
	return []byte(fmt.Sprintf("gobook-v%d-%s", FormatVersion, payloadName))
}

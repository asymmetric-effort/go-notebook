// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package crypto implements the encryption format for embedded data and
// secrets in report binaries.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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

// HeaderSize is the fixed size of a serialized Header.
const HeaderSize = 1 + 4 + 4 + 1 + SaltLen // version + time + memory + threads + salt = 26

// MarshalBinary serializes the header to a fixed-size binary format.
func (h Header) MarshalBinary() []byte {
	buf := make([]byte, HeaderSize)
	buf[0] = h.Version
	binary.BigEndian.PutUint32(buf[1:5], h.Argon2Time)
	binary.BigEndian.PutUint32(buf[5:9], h.Argon2Memory)
	buf[9] = h.Argon2Threads
	copy(buf[10:], h.Salt[:])
	return buf
}

// UnmarshalHeader deserializes a header from binary data.
func UnmarshalHeader(data []byte) (Header, error) {
	if len(data) < HeaderSize {
		return Header{}, fmt.Errorf("header too short: got %d bytes, need %d", len(data), HeaderSize)
	}
	var h Header
	h.Version = data[0]
	h.Argon2Time = binary.BigEndian.Uint32(data[1:5])
	h.Argon2Memory = binary.BigEndian.Uint32(data[5:9])
	h.Argon2Threads = data[9]
	copy(h.Salt[:], data[10:10+SaltLen])
	return h, nil
}

// EncryptPayload encrypts data with a passphrase and returns the complete
// binary payload (header + ciphertext). The header contains the Argon2id
// parameters and salt needed for decryption.
func EncryptPayload(passphrase []byte, plaintext []byte, payloadName string) ([]byte, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}
	masterKey := DeriveKey(passphrase, salt[:])
	payloadKey, err := DerivePayloadKey(masterKey, payloadName)
	if err != nil {
		return nil, err
	}
	ciphertext, err := Encrypt(payloadKey, plaintext, payloadName)
	if err != nil {
		return nil, err
	}

	header := Header{
		Version:       FormatVersion,
		Argon2Time:    Argon2Time,
		Argon2Memory:  Argon2Memory,
		Argon2Threads: Argon2Threads,
		Salt:          salt,
	}
	return append(header.MarshalBinary(), ciphertext...), nil
}

// DecryptPayload decrypts a complete binary payload (header + ciphertext)
// using the given passphrase and payload name.
func DecryptPayload(passphrase []byte, data []byte, payloadName string) ([]byte, error) {
	header, err := UnmarshalHeader(data)
	if err != nil {
		return nil, err
	}
	if header.Version != FormatVersion {
		return nil, fmt.Errorf("unsupported format version %d", header.Version)
	}
	// Validate Argon2 parameters to prevent OOM from crafted headers.
	if header.Argon2Memory > 4*1024*1024 { // Max 4 GiB
		return nil, fmt.Errorf("argon2 memory parameter too large: %d KiB", header.Argon2Memory)
	}
	if header.Argon2Time > 100 {
		return nil, fmt.Errorf("argon2 time parameter too large: %d", header.Argon2Time)
	}
	if header.Argon2Threads == 0 || header.Argon2Threads > 64 {
		return nil, fmt.Errorf("argon2 threads parameter invalid: %d", header.Argon2Threads)
	}

	masterKey := argon2.IDKey(passphrase, header.Salt[:],
		header.Argon2Time, header.Argon2Memory, header.Argon2Threads, Argon2KeyLen)
	payloadKey, err := DerivePayloadKey(masterKey, payloadName)
	if err != nil {
		return nil, err
	}
	return Decrypt(payloadKey, data[HeaderSize:], payloadName)
}

// authenticatedData creates the AAD for AES-256-GCM.
func authenticatedData(payloadName string) []byte {
	return []byte(fmt.Sprintf("gobook-v%d-%s", FormatVersion, payloadName))
}

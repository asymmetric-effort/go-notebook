// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"crypto/hpke"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// RecipientHeader stores the encrypted data key for one recipient.
type RecipientHeader struct {
	EncryptedKey []byte // HPKE-sealed data key.
}

// EncryptForRecipients encrypts data for multiple recipients using HPKE
// with X-Wing (MLKEM768X25519). Each recipient's public key wraps the
// same random data key. Returns the encrypted payload and per-recipient
// headers.
func EncryptForRecipients(plaintext []byte, pubKeys []hpke.PublicKey, payloadName string) ([]byte, []RecipientHeader, error) {
	if len(pubKeys) == 0 {
		return nil, nil, fmt.Errorf("at least one recipient required")
	}

	// Generate a random data key.
	dataKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dataKey); err != nil {
		return nil, nil, fmt.Errorf("generate data key: %w", err)
	}

	// Encrypt the data key for each recipient.
	info := []byte("gobook-recipient-" + payloadName)
	var headers []RecipientHeader
	for i, pk := range pubKeys {
		sealed, err := hpke.Seal(pk, hpke.HKDFSHA256(), hpke.AES256GCM(), info, dataKey)
		if err != nil {
			return nil, nil, fmt.Errorf("encrypt for recipient %d: %w", i, err)
		}
		headers = append(headers, RecipientHeader{EncryptedKey: sealed})
	}

	// Derive a payload key from the data key and encrypt.
	payloadKey, err := DerivePayloadKey(dataKey, payloadName)
	if err != nil {
		return nil, nil, err
	}
	ciphertext, err := Encrypt(payloadKey, plaintext, payloadName)
	if err != nil {
		return nil, nil, err
	}

	return ciphertext, headers, nil
}

// DecryptForRecipient decrypts a payload using the recipient's private key.
// It tries each header until one succeeds.
func DecryptForRecipient(ciphertext []byte, headers []RecipientHeader, privKey hpke.PrivateKey, payloadName string) ([]byte, error) {
	info := []byte("gobook-recipient-" + payloadName)

	for _, h := range headers {
		dataKey, err := hpke.Open(privKey, hpke.HKDFSHA256(), hpke.AES256GCM(), info, h.EncryptedKey)
		if err != nil {
			continue // Not for this recipient.
		}

		payloadKey, err := DerivePayloadKey(dataKey, payloadName)
		if err != nil {
			return nil, err
		}
		return Decrypt(payloadKey, ciphertext, payloadName)
	}
	return nil, fmt.Errorf("no recipient header could be decrypted with the given key")
}

// MarshalRecipientHeaders serializes recipient headers to binary.
func MarshalRecipientHeaders(headers []RecipientHeader) []byte {
	var buf []byte
	// Number of recipients (uint16).
	countBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(countBuf, uint16(len(headers)))
	buf = append(buf, countBuf...)
	for _, h := range headers {
		// Length of encrypted key (uint16) + encrypted key.
		lenBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBuf, uint16(len(h.EncryptedKey)))
		buf = append(buf, lenBuf...)
		buf = append(buf, h.EncryptedKey...)
	}
	return buf
}

// UnmarshalRecipientHeaders deserializes recipient headers from binary.
func UnmarshalRecipientHeaders(data []byte) ([]RecipientHeader, int, error) {
	if len(data) < 2 {
		return nil, 0, fmt.Errorf("recipient header too short")
	}
	count := int(binary.BigEndian.Uint16(data[:2]))
	offset := 2
	var headers []RecipientHeader
	for i := 0; i < count; i++ {
		if offset+2 > len(data) {
			return nil, 0, fmt.Errorf("truncated recipient header %d", i)
		}
		keyLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if offset+keyLen > len(data) {
			return nil, 0, fmt.Errorf("truncated recipient key %d", i)
		}
		headers = append(headers, RecipientHeader{EncryptedKey: data[offset : offset+keyLen]})
		offset += keyLen
	}
	return headers, offset, nil
}

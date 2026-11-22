// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package identity implements post-quantum key generation and management
// for gobook's recipient and signing modes.
package identity

import (
	"crypto/mldsa"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"
)

// Identity holds a post-quantum key pair: ML-DSA-65 for signing.
type Identity struct {
	SigningKey *mldsa.PrivateKey
	VerifyKey  *mldsa.PublicKey
}

// Generate creates a new post-quantum identity with ML-DSA-65 keys.
func Generate() (*Identity, error) {
	priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		return nil, fmt.Errorf("generate ML-DSA-65: %w", err)
	}
	return &Identity{
		SigningKey: priv,
		VerifyKey:  priv.PublicKey(),
	}, nil
}

// Fingerprint returns a human-readable fingerprint of the public key.
func (id *Identity) Fingerprint() string {
	pubBytes := id.VerifyKey.Bytes()
	hash := sha256.Sum256(pubBytes)
	return fmt.Sprintf("SHA256:%x", hash[:16])
}

// WriteKeys writes the public and private keys to PEM files.
// The private key file has owner-only permissions (0600).
func (id *Identity) WriteKeys(baseName string) error {
	// Write public key.
	pubBytes := id.VerifyKey.Bytes()
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "GOBOOK ML-DSA-65 PUBLIC KEY",
		Bytes: pubBytes,
	})
	pubPath := baseName + ".pub"
	if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}

	// Write private key with owner-only permissions.
	privBytes := id.SigningKey.Bytes()
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "GOBOOK ML-DSA-65 PRIVATE KEY",
		Bytes: privBytes,
	})
	privPath := baseName + ".key"
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}

	return nil
}

// ReadPublicKey reads a public key from a PEM file.
func ReadPublicKey(path string) (*mldsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}
	pub, err := mldsa.NewPublicKey(mldsa.MLDSA65(), block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	return pub, nil
}

// ReadPrivateKey reads a private key from a PEM file.
func ReadPrivateKey(path string) (*mldsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}
	priv, err := mldsa.NewPrivateKey(mldsa.MLDSA65(), block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return priv, nil
}

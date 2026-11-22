// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package identity

import (
	"crypto/mldsa"
	"encoding/pem"
	"fmt"
	"os"
)

// Sign signs the given data with the private key and returns the signature.
func Sign(priv *mldsa.PrivateKey, data []byte) ([]byte, error) {
	sig, err := priv.Sign(nil, data, nil)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	return sig, nil
}

// Verify checks a signature against the public key and data.
func Verify(pub *mldsa.PublicKey, data []byte, sig []byte) error {
	return mldsa.Verify(pub, data, sig, nil)
}

// SignFile signs the contents of a file and writes a detached .sig file.
func SignFile(dataPath, keyPath string) error {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", dataPath, err)
	}
	priv, err := ReadPrivateKey(keyPath)
	if err != nil {
		return err
	}
	sig, err := Sign(priv, data)
	if err != nil {
		return err
	}

	sigPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "GOBOOK ML-DSA-65 SIGNATURE",
		Bytes: sig,
	})
	sigPath := dataPath + ".sig"
	if err := os.WriteFile(sigPath, sigPEM, 0o644); err != nil {
		return fmt.Errorf("write signature: %w", err)
	}
	return nil
}

// VerifyFile verifies a detached signature against a data file and public key.
func VerifyFile(dataPath, sigPath, pubKeyPath string) error {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", dataPath, err)
	}
	sigData, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", sigPath, err)
	}
	block, _ := pem.Decode(sigData)
	if block == nil {
		return fmt.Errorf("no PEM block in %s", sigPath)
	}
	pub, err := ReadPublicKey(pubKeyPath)
	if err != nil {
		return err
	}
	if err := Verify(pub, data, block.Bytes); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package identity

import (
	"crypto/mldsa"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if id.SigningKey == nil {
		t.Error("signing key should not be nil")
	}
	if id.VerifyKey == nil {
		t.Error("verify key should not be nil")
	}
}

func TestFingerprint(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	fp := id.Fingerprint()
	if !strings.HasPrefix(fp, "SHA256:") {
		t.Errorf("fingerprint = %q, want SHA256: prefix", fp)
	}
	if len(fp) < 30 {
		t.Errorf("fingerprint too short: %q", fp)
	}
}

func TestFingerprintDeterministic(t *testing.T) {
	id, _ := Generate()
	fp1 := id.Fingerprint()
	fp2 := id.Fingerprint()
	if fp1 != fp2 {
		t.Error("fingerprint should be deterministic")
	}
}

func TestFingerprintUnique(t *testing.T) {
	id1, _ := Generate()
	id2, _ := Generate()
	if id1.Fingerprint() == id2.Fingerprint() {
		t.Error("different keys should have different fingerprints")
	}
}

func TestWriteAndReadKeys(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	baseName := filepath.Join(dir, "test-key")
	if err := id.WriteKeys(baseName); err != nil {
		t.Fatal(err)
	}

	// Check files exist.
	if _, err := os.Stat(baseName + ".pub"); err != nil {
		t.Error("public key file not found")
	}
	if _, err := os.Stat(baseName + ".key"); err != nil {
		t.Error("private key file not found")
	}

	// Check private key permissions.
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(baseName + ".key")
		if info.Mode().Perm() != 0o600 {
			t.Errorf("private key permissions = %o, want 0600", info.Mode().Perm())
		}
	}

	// Read back and verify.
	pub, err := ReadPublicKey(baseName + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	if pub == nil {
		t.Error("read public key should not be nil")
	}

	priv, err := ReadPrivateKey(baseName + ".key")
	if err != nil {
		t.Fatal(err)
	}
	if priv == nil {
		t.Error("read private key should not be nil")
	}

	// Verify the keys work: sign and verify.
	msg := []byte("test message")
	sig, err := priv.Sign(nil, msg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mldsa.Verify(pub, msg, sig, nil); err != nil {
		t.Fatal(err)
	}
}

func TestReadPublicKeyMissing(t *testing.T) {
	_, err := ReadPublicKey("/nonexistent/key.pub")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadPrivateKeyMissing(t *testing.T) {
	_, err := ReadPrivateKey("/nonexistent/key.key")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadPublicKeyBadPEM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pub")
	os.WriteFile(path, []byte("not pem"), 0o644)
	_, err := ReadPublicKey(path)
	if err == nil {
		t.Error("expected error for bad PEM")
	}
}

func TestReadPrivateKeyBadPEM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.key")
	os.WriteFile(path, []byte("not pem"), 0o600)
	_, err := ReadPrivateKey(path)
	if err == nil {
		t.Error("expected error for bad PEM")
	}
}

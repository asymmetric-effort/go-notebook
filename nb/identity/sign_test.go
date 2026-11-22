// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("test message to sign")
	sig, err := Sign(id.SigningKey, msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(id.VerifyKey, msg, sig); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWrongMessage(t *testing.T) {
	id, _ := Generate()
	sig, _ := Sign(id.SigningKey, []byte("original"))
	if err := Verify(id.VerifyKey, []byte("tampered"), sig); err == nil {
		t.Error("expected verification failure for wrong message")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	id1, _ := Generate()
	id2, _ := Generate()
	sig, _ := Sign(id1.SigningKey, []byte("message"))
	if err := Verify(id2.VerifyKey, []byte("message"), sig); err == nil {
		t.Error("expected verification failure for wrong key")
	}
}

func TestSignAndVerifyFile(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	baseName := filepath.Join(dir, "test-key")
	id.WriteKeys(baseName)

	// Create a data file.
	dataPath := filepath.Join(dir, "report.html")
	os.WriteFile(dataPath, []byte("<html>report</html>"), 0o644)

	// Sign.
	if err := SignFile(dataPath, baseName+".key"); err != nil {
		t.Fatal(err)
	}

	// Verify.
	if err := VerifyFile(dataPath, dataPath+".sig", baseName+".pub"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFileTampered(t *testing.T) {
	id, _ := Generate()
	dir := t.TempDir()
	baseName := filepath.Join(dir, "key")
	id.WriteKeys(baseName)

	dataPath := filepath.Join(dir, "data.txt")
	os.WriteFile(dataPath, []byte("original"), 0o644)
	SignFile(dataPath, baseName+".key")

	// Tamper with the data.
	os.WriteFile(dataPath, []byte("tampered"), 0o644)

	err := VerifyFile(dataPath, dataPath+".sig", baseName+".pub")
	if err == nil {
		t.Error("expected verification failure for tampered data")
	}
}

func TestSignFileMissingData(t *testing.T) {
	err := SignFile("/nonexistent/file", "/nonexistent/key")
	if err == nil {
		t.Error("expected error for missing data file")
	}
}

func TestSignFileMissingKey(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "data.txt")
	os.WriteFile(dataPath, []byte("data"), 0o644)

	err := SignFile(dataPath, "/nonexistent/key.key")
	if err == nil {
		t.Error("expected error for missing key file")
	}
}

func TestVerifyFileMissingSig(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "data.txt")
	os.WriteFile(dataPath, []byte("data"), 0o644)

	err := VerifyFile(dataPath, "/nonexistent/data.sig", "/nonexistent/key.pub")
	if err == nil {
		t.Error("expected error for missing sig file")
	}
}

func TestVerifyFileBadSigPEM(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "data.txt")
	sigPath := filepath.Join(dir, "data.sig")
	os.WriteFile(dataPath, []byte("data"), 0o644)
	os.WriteFile(sigPath, []byte("not pem"), 0o644)

	err := VerifyFile(dataPath, sigPath, "/nonexistent/key.pub")
	if err == nil {
		t.Error("expected error for bad signature PEM")
	}
}

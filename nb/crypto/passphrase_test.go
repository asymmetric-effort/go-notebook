// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePassphraseMinLength(t *testing.T) {
	// 12 characters should pass.
	if err := ValidatePassphrase("123456789012"); err != nil {
		t.Errorf("12 chars should pass: %v", err)
	}
}

func TestValidatePassphraseTooShort(t *testing.T) {
	if err := ValidatePassphrase("short"); err == nil {
		t.Error("5 chars should fail")
	}
}

func TestValidatePassphraseEmpty(t *testing.T) {
	// Empty passphrase means no encryption — always valid.
	if err := ValidatePassphrase(""); err != nil {
		t.Errorf("empty passphrase should pass (no encryption): %v", err)
	}
}

func TestValidatePassphraseRule13InError(t *testing.T) {
	err := ValidatePassphrase("short")
	if err == nil {
		t.Fatal("short passphrase should fail")
	}
	if !strings.Contains(err.Error(), "rule 13") {
		t.Errorf("error should reference rule 13: %v", err)
	}
}

func TestValidatePassphraseUnicode(t *testing.T) {
	// 12 Unicode code points (some multi-byte).
	pass := "日本語テスト用パスワード十二" // 12 runes, more than 12 bytes
	if err := ValidatePassphrase(pass); err != nil {
		t.Errorf("12 Unicode code points should pass: %v", err)
	}
}

func TestValidatePassphraseUnicodeTooShort(t *testing.T) {
	pass := "日本語" // 3 runes
	if err := ValidatePassphrase(pass); err == nil {
		t.Error("3 Unicode code points should fail")
	}
}

func TestReadPassphraseFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	os.WriteFile(path, []byte("my-long-passphrase\n"), 0o600)

	pass, err := ReadPassphraseFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if pass != "my-long-passphrase" {
		t.Errorf("pass = %q", pass)
	}
}

func TestReadPassphraseFromFileNoNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	os.WriteFile(path, []byte("passphrase-no-newline"), 0o600)

	pass, err := ReadPassphraseFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if pass != "passphrase-no-newline" {
		t.Errorf("pass = %q", pass)
	}
}

func TestReadPassphraseFromFileMissing(t *testing.T) {
	_, err := ReadPassphraseFromFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadPassphraseFromFileWindowsNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	os.WriteFile(path, []byte("passphrase\r\n"), 0o600)

	pass, err := ReadPassphraseFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if pass != "passphrase" {
		t.Errorf("pass = %q (should trim \\r\\n)", pass)
	}
}

func TestMinPassphraseLength(t *testing.T) {
	if MinPassphraseLength != 12 {
		t.Errorf("MinPassphraseLength = %d, want 12", MinPassphraseLength)
	}
}

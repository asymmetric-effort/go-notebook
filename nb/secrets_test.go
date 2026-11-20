// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package nb

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveSecretFromEnv(t *testing.T) {
	t.Setenv("GOBOOK_SECRET_DB", "mypassword")
	val, err := ResolveSecret("db", "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if val != "mypassword" {
		t.Errorf("val = %q, want %q", val, "mypassword")
	}
}

func TestResolveSecretFromFile(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	secretsFile := notebook + ".secrets"

	// Write secrets file with owner-only permissions.
	if err := os.WriteFile(secretsFile, []byte("db=file_password\napi=key123\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Ensure env var is NOT set.
	t.Setenv("GOBOOK_SECRET_DB", "")

	val, err := ResolveSecret("db", notebook)
	if err != nil {
		t.Fatal(err)
	}
	if val != "file_password" {
		t.Errorf("val = %q, want %q", val, "file_password")
	}
}

func TestResolveSecretFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions work differently on Windows")
	}

	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	secretsFile := notebook + ".secrets"

	// Write secrets file with too-wide permissions.
	os.WriteFile(secretsFile, []byte("db=password\n"), 0o644)
	t.Setenv("GOBOOK_SECRET_DB", "")

	_, err := ResolveSecret("db", notebook)
	if err == nil {
		t.Error("expected error for wide permissions")
	}
}

func TestResolveSecretNotFound(t *testing.T) {
	t.Setenv("GOBOOK_SECRET_MISSING", "")
	_, err := ResolveSecret("missing", "/nonexistent/test.md")
	if err == nil {
		t.Error("expected error when secret not found anywhere")
	}
}

func TestReadSecretsFileComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.secrets")
	content := "# comment\n\ndb=value\n# another comment\napi=key\n"
	os.WriteFile(path, []byte(content), 0o600)

	val, err := readSecretsFile(path, "db")
	if err != nil {
		t.Fatal(err)
	}
	if val != "value" {
		t.Errorf("val = %q", val)
	}

	val, err = readSecretsFile(path, "api")
	if err != nil {
		t.Fatal(err)
	}
	if val != "key" {
		t.Errorf("val = %q", val)
	}
}

func TestReadSecretsFileNotFound(t *testing.T) {
	_, err := readSecretsFile("/nonexistent/file", "db")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadSecretsFileSecretNotInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.secrets")
	os.WriteFile(path, []byte("other=value\n"), 0o600)

	_, err := readSecretsFile(path, "db")
	if err == nil {
		t.Error("expected error for secret not in file")
	}
}

func TestResolveSecretFromPrompt(t *testing.T) {
	// Inject a mock prompt.
	original := PromptFunc
	defer func() { PromptFunc = original }()
	PromptFunc = func(name string) (string, error) {
		return "prompted_value", nil
	}

	t.Setenv("GOBOOK_SECRET_PROMPTED", "")
	val, err := ResolveSecret("prompted", "/nonexistent/test.md")
	if err != nil {
		t.Fatal(err)
	}
	if val != "prompted_value" {
		t.Errorf("val = %q, want %q", val, "prompted_value")
	}
}

func TestResolveSecretEnvPriority(t *testing.T) {
	// Env should take priority over file.
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	secretsFile := notebook + ".secrets"
	os.WriteFile(secretsFile, []byte("db=file_value\n"), 0o600)

	t.Setenv("GOBOOK_SECRET_DB", "env_value")

	val, err := ResolveSecret("db", notebook)
	if err != nil {
		t.Fatal(err)
	}
	if val != "env_value" {
		t.Errorf("env should take priority: got %q, want %q", val, "env_value")
	}
}

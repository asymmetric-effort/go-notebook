// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package nb

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestEnsureGitignoreCreatesFile(t *testing.T) {
	dir := t.TempDir()
	ensureGitignore(dir, "test.md.secrets")

	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test.md.secrets") {
		t.Errorf(".gitignore content = %q, want test.md.secrets", content)
	}
}

func TestEnsureGitignoreAppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.exe\n"), 0o644)

	ensureGitignore(dir, "test.md.secrets")

	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test.md.secrets") {
		t.Error("pattern not appended")
	}
	if !strings.Contains(string(content), "*.exe") {
		t.Error("existing content removed")
	}
}

func TestEnsureGitignoreNoDuplicate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("test.md.secrets\n"), 0o644)

	ensureGitignore(dir, "test.md.secrets")

	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	// Should only appear once.
	count := strings.Count(string(content), "test.md.secrets")
	if count != 1 {
		t.Errorf("pattern appears %d times, want 1", count)
	}
}

func TestEnsureGitignoreEmptyDir(t *testing.T) {
	// Save current dir and change to temp.
	orig, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(orig)

	ensureGitignore("", "test.secrets")
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err == nil && !strings.Contains(string(content), "test.secrets") {
		t.Error("empty dir should default to '.'")
	}
}

func TestEnsureGitignoreNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	// File without trailing newline.
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.exe"), 0o644)
	ensureGitignore(dir, "test.secrets")
	content, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !strings.Contains(string(content), "\ntest.secrets") {
		t.Error("should add newline before pattern")
	}
}

func TestResolveSecretUpdatesGitignore(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	secretsFile := notebook + ".secrets"
	os.WriteFile(secretsFile, []byte("db=myvalue\n"), 0o600)

	val, err := ResolveSecret("db", notebook)
	if err != nil {
		t.Fatal(err)
	}
	if val != "myvalue" {
		t.Errorf("val = %q, want myvalue", val)
	}

	// .gitignore should now contain the secrets file pattern.
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test.md.secrets") {
		t.Errorf(".gitignore = %q, want test.md.secrets", content)
	}
}

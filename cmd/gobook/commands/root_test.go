// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

func TestCmdCheckClean(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test\nfmt.Println(\"hi\")\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdCheck(path); err != nil {
		t.Fatalf("cmdCheck = %v", err)
	}
}

func TestCmdCheckErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook exported(val int)\nval = 1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdCheck(path); err == nil {
		t.Error("expected error for exported without name")
	}
}

func TestCmdCheckMissingFile(t *testing.T) {
	if err := cmdCheck("/nonexistent/file.md"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestCmdFmt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test\nx:=1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdFmt(path); err != nil {
		t.Fatalf("cmdFmt = %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) == src {
		t.Error("file should have been reformatted")
	}
}

func TestCmdGen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "gen")
	if err := cmdGen(path, outDir); err != nil {
		t.Fatalf("cmdGen = %v", err)
	}
}

func TestCmdRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdRun(path); err != nil {
		t.Fatalf("cmdRun = %v", err)
	}
}

func TestCmdRunWithPrint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(val int)\nval = 42\nfmt.Println(\"hello from cell\")\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdRun(path)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun = %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "hello from cell") {
		t.Errorf("expected cell output in stdout, got: %q", output)
	}
}

func TestCmdRunWritesOutputBlocks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=greet exported(msg string)\nmsg = \"hello\"\nfmt.Println(\"greeting:\", msg)\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	err := cmdRun(path)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(updated)
	if !strings.Contains(content, "```output name=greet") {
		t.Errorf("expected output block written back to file.\nFile content:\n%s", content)
	}
	if !strings.Contains(content, "greeting:") {
		t.Errorf("expected cell output in output block.\nFile content:\n%s", content)
	}
}

func TestLastNamedCell(t *testing.T) {
	src := "```go:notebook name=a\n```\n\n```go:notebook\n```\n\n```go:notebook name=b\n```\n"
	result := core.Parse(src)
	last := lastNamedCell(result)
	if last != "b" {
		t.Errorf("lastNamedCell = %q, want %q", last, "b")
	}
}

func TestLastNamedCellEmpty(t *testing.T) {
	src := "```go:notebook\n```\n"
	result := core.Parse(src)
	last := lastNamedCell(result)
	if last != "" {
		t.Errorf("lastNamedCell = %q, want empty", last)
	}
}

func TestCmdRunPanicRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\npanic(\"test\")\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	err := cmdRun(path)
	if err != nil {
		t.Logf("cmdRun error (may be expected): %v", err)
	}
}

func TestCmdVet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdVet(path); err != nil {
		t.Fatalf("cmdVet = %v", err)
	}
}

func TestCountCells(t *testing.T) {
	src := "```go:notebook name=a\n```\n\n```go:notebook name=b\n```\n"
	result := core.Parse(src)
	if n := countCells(result); n != 2 {
		t.Errorf("countCells = %d, want 2", n)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute returns nil for --help (cobra handles it).
	os.Args = []string{"gobook", "--help"}
	err := Execute()
	if err != nil {
		t.Errorf("Execute(--help) = %v", err)
	}
}

func TestResolvePassphraseNone(t *testing.T) {
	// No flags set — no passphrase needed.
	pass, err := resolvePassphrase(false, "")
	if err != nil {
		t.Fatal(err)
	}
	if pass != "" {
		t.Errorf("expected empty passphrase, got %q", pass)
	}
}

func TestResolvePassphraseFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	os.WriteFile(path, []byte("a-long-enough-passphrase\n"), 0o600)

	pass, err := resolvePassphrase(false, path)
	if err != nil {
		t.Fatal(err)
	}
	if pass != "a-long-enough-passphrase" {
		t.Errorf("passphrase = %q", pass)
	}
}

func TestResolvePassphraseFileTooShort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	os.WriteFile(path, []byte("short\n"), 0o600)

	_, err := resolvePassphrase(false, path)
	if err == nil {
		t.Error("expected error for short passphrase")
	}
	if !strings.Contains(err.Error(), "passphrase rejected") {
		t.Errorf("error = %v, want 'passphrase rejected'", err)
	}
}

func TestResolvePassphraseFileMissing(t *testing.T) {
	_, err := resolvePassphrase(false, "/nonexistent/pass.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

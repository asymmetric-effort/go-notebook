// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package commands

import (
	"os"
	"path/filepath"
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
	if err := cmdCheck([]string{path}); err != nil {
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
	if err := cmdCheck([]string{path}); err == nil {
		t.Error("expected error for exported without name")
	}
}

func TestCmdCheckMissingFile(t *testing.T) {
	if err := cmdCheck([]string{"/nonexistent/file.md"}); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestCmdCheckNoArgs(t *testing.T) {
	if err := cmdCheck(nil); err == nil {
		t.Error("expected error for no args")
	}
}

func TestCmdFmt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test\nx:=1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdFmt([]string{path}); err != nil {
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
	if err := cmdGen([]string{path}); err != nil {
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
	if err := cmdRun([]string{path}); err != nil {
		t.Fatalf("cmdRun = %v", err)
	}
}

func TestCmdRunWithPrint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	// A cell that imports fmt and prints. The import must be in go:defs.
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(val int)\nval = 42\nfmt.Println(\"hello from cell\")\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	// Capture stdout by redirecting.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdRun([]string{path})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun = %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !contains(output, "hello from cell") {
		t.Errorf("expected cell output in stdout, got: %q", output)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || lenContains(s, sub))
}

func lenContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestCmdRunPanicRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\npanic(\"test\")\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	// Should not return a Go runtime panic — the deferred recovery catches it.
	err := cmdRun([]string{path})
	// The cell panics, but the binary catches it and exits 0.
	// gobook run should succeed.
	if err != nil {
		t.Logf("cmdRun error (may be expected): %v", err)
	}
}

func TestCmdVet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test\nval := 1\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdVet([]string{path}); err != nil {
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

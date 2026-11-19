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

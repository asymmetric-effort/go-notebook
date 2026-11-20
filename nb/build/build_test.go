// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

func TestRunMarkdown(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Report\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.md")
	err := Run(notebook, Options{Format: "md", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# Report") {
		t.Error("missing prose in output")
	}
}

func TestRunHTML(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Report\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.html")
	err := Run(notebook, Options{Format: "html", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE in HTML output")
	}
	if !strings.Contains(string(data), "Content-Security-Policy") {
		t.Error("missing CSP in HTML output")
	}
}

func TestRunDefaultFormat(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	if err := os.WriteFile(notebook, []byte("# Title\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Run(notebook, Options{Output: filepath.Join(dir, "out.html")})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunCheckErrors(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook exported(val int)\nval = 1\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Run(notebook, Options{Output: filepath.Join(dir, "out.html")})
	if err == nil {
		t.Error("expected error for exported without name")
	}
}

func TestRunMissingFile(t *testing.T) {
	err := Run("/nonexistent/file.md", Options{})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunBadFormat(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	if err := os.WriteFile(notebook, []byte("# Title\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Run(notebook, Options{Format: "bad", Output: filepath.Join(dir, "out")})
	if err == nil {
		t.Error("expected error for bad format")
	}
}

func TestRunDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	if err := os.WriteFile(notebook, []byte("# Title\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir so default output goes there.
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	err := Run(notebook, Options{Format: "md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(notebook + ".md"); err != nil {
		t.Error("default output file not created")
	}
}

func TestBuildDocument(t *testing.T) {
	src := "# Title\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test\nfmt.Println(\"hi\")\n```\n\n```output name=test\nhi\n```\n"
	result := core.Parse(src)
	doc := buildDocument(result)
	if doc.Len() == 0 {
		t.Error("document should have elements")
	}
}

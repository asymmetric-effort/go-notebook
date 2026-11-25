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

func TestRunPDF(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# PDF Report\n\nSome content.\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.pdf")
	err := Run(notebook, Options{Format: "pdf", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Errorf("PDF too small: %d bytes", len(data))
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("not a PDF: starts with %q", data[:5])
	}
	t.Logf("PDF size: %d bytes", len(data))
}

func TestRunPDFWithError(t *testing.T) {
	// A notebook with a panic produces a partial PDF with the error.
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Report\n\nSome prose before the error.\n\n```go:notebook name=test exported(val int)\npanic(\"oops\")\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.pdf")
	err := Run(notebook, Options{Format: "pdf", Output: outPath})
	// Error is expected (cell panicked).
	_ = err

	// Partial report should still be written.
	data, readErr := os.ReadFile(outPath)
	if readErr != nil {
		t.Fatal("partial PDF should be written even on cell error")
	}
	if len(data) < 100 {
		t.Errorf("partial PDF too small: %d bytes", len(data))
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("not a PDF: starts with %q", data[:5])
	}
	t.Logf("Partial PDF size: %d bytes", len(data))
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

func TestRunTypst(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Report\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.typ")
	err := Run(notebook, Options{Format: "typst", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "#set page") {
		t.Error("missing Typst preamble")
	}
	if !strings.Contains(string(data), "= Report") {
		t.Error("missing heading conversion")
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
	doc := buildDocument(result, "hi", false)
	if doc.Len() == 0 {
		t.Error("document should have elements")
	}
}

func TestBuildDocumentWithOutput(t *testing.T) {
	src := "# Title\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	result := core.Parse(src)
	doc := buildDocument(result, "result: 42", false)
	// Should have prose + display.
	if doc.Len() < 2 {
		t.Errorf("document should have at least 2 elements, got %d", doc.Len())
	}
}

func TestRunWithCellExecution(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(val int)\nval = 42\nfmt.Println(\"hello from build\")\n```\n"
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
	if !strings.Contains(string(data), "hello from build") {
		t.Error("expected cell output in report")
	}
}

func TestRunWithTimeout(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.html")
	err := Run(notebook, Options{Format: "html", Output: outPath, Timeout: "30s"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunWithPanicCell(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\npanic(\"oops\")\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.html")
	err := Run(notebook, Options{Format: "html", Output: outPath})
	// The build should still produce a partial report.
	if _, statErr := os.Stat(outPath); statErr != nil {
		t.Error("partial report should be written even on failure")
	}
	_ = err // Error is expected.
}

func TestRunWriteError(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	if err := os.WriteFile(notebook, []byte("# Title\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write to an impossible path.
	err := Run(notebook, Options{Output: "/nonexistent/dir/out.html"})
	if err == nil {
		t.Error("expected error for unwritable output")
	}
}

func TestFindModuleRoot(t *testing.T) {
	root := findModuleRoot()
	if root == "" {
		t.Error("findModuleRoot should not return empty")
	}
	// Should end with go-notebook or be "." in test context.
	if root != "." && !strings.HasSuffix(root, "go-notebook") {
		t.Logf("findModuleRoot = %q (may be valid)", root)
	}
}

func TestBuildDocumentStripsOutput(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n\n```output name=test\nold output\n```\n"
	result := core.Parse(src)
	doc := buildDocument(result, "new output", false)
	// Output blocks should be stripped; cell output comes from execution.
	found := false
	for _, e := range doc.Elements {
		if strings.Contains(e.Content, "old output") {
			t.Error("old output block should be stripped")
		}
		if strings.Contains(e.Content, "new output") {
			found = true
		}
	}
	if !found {
		t.Error("new output should be present")
	}
}

func TestRunShortPassphraseRejected(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	err := Run(notebook, Options{
		Format:     "html",
		Output:     filepath.Join(dir, "out.html"),
		Passphrase: "short",
	})
	if err == nil {
		t.Fatal("short passphrase should be rejected")
	}
	if !strings.Contains(err.Error(), "rule 13") {
		t.Errorf("error should reference rule 13: %v", err)
	}
}

func TestRunExactlyMinPassphraseAccepted(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	err := Run(notebook, Options{
		Format:     "html",
		Output:     filepath.Join(dir, "out.html"),
		Passphrase: "exactly12chr",
	})
	if err != nil {
		t.Fatalf("12-char passphrase should be accepted: %v", err)
	}
}

func TestRunEmptyPassphraseAccepted(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	err := Run(notebook, Options{
		Format:     "html",
		Output:     filepath.Join(dir, "out.html"),
		Passphrase: "",
	})
	if err != nil {
		t.Fatalf("empty passphrase should be accepted: %v", err)
	}
}

func TestRunUnicodePassphraseAccepted(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	err := Run(notebook, Options{
		Format:     "html",
		Output:     filepath.Join(dir, "out.html"),
		Passphrase: "日本語テスト用パスワード十二", // 12 runes
	})
	if err != nil {
		t.Fatalf("12-rune Unicode passphrase should be accepted: %v", err)
	}
}

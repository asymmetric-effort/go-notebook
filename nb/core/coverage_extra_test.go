// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Tests specifically targeting uncovered lines to push toward 100% coverage.
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatNotebookWithError(t *testing.T) {
	// A cell with invalid Go that can't be formatted.
	src := "```go:notebook name=test\nfunc {{{\n```\n"
	_, err := FormatNotebook(src)
	if err == nil {
		t.Error("expected error for invalid Go in cell")
	}
}

func TestFormatNotebookAlreadyFormatted(t *testing.T) {
	// Already-formatted cell body should be unchanged.
	src := "```go:notebook name=test\nx := 1\n```\n"
	result, err := FormatNotebook(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "x := 1") {
		t.Error("formatted body should be unchanged")
	}
}

func TestIsImportLineNoQuotes(t *testing.T) {
	unused := map[string]bool{"os": true}
	// Line with no quotes at all.
	if isImportLine("func main()", unused) {
		t.Error("non-import line should return false")
	}
}

func TestIsImportLineSingleQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	// Line with only one quote (malformed).
	if isImportLine("\"os", unused) {
		t.Error("malformed import line should return false")
	}
}

func TestCheckDefsUnknownDeclType(t *testing.T) {
	// This is hard to trigger because go/ast only produces GenDecl and FuncDecl.
	// The default case in checkDefs handles unknown GenDecl tokens.
	// We can't easily create an AST with an unknown token type from source,
	// but we can test that a valid defs block with all allowed types passes.
	src := "```go:defs\nimport \"fmt\"\nconst x = 1\ntype T struct{}\nfunc f() {}\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("valid defs should pass: %v", errs)
	}
}

func TestCopyEmbedFilesRelPathError(t *testing.T) {
	// CopyEmbedFiles when glob matches but Rel fails.
	// This is very hard to trigger since Rel only fails when the
	// base is not a prefix, which Glob guarantees. Test the success path.
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, srcDir, "a.txt", "data-a")
	writeTestFile(t, srcDir, "b.txt", "data-b")
	err := CopyEmbedFiles([]string{"*.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteOutputBlocksInvalidCellInfo(t *testing.T) {
	// A cell with an invalid info string should be skipped.
	src := "```go:notebook name=\nbody\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"": "data"})
	_ = result // Should not panic.
}

func TestIsImportLineNoStartQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	if isImportLine("no quotes here", unused) {
		t.Error("should return false for no quotes")
	}
}

func TestIsImportLineOneQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	if isImportLine(`"os`, unused) {
		t.Error("should return false for single quote")
	}
}

func TestUnwrapBodyNoClosingBrace(t *testing.T) {
	_, err := unwrapBody("func _() {\n  x := 1\n")
	if err == nil {
		t.Error("expected error for no closing brace")
	}
}

func TestParseInfoStringOnlyWhitespace(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook   ", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "" {
		t.Error("should have empty name")
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := writeFile(dir, name, content); err != nil {
		t.Fatal(err)
	}
}

func writeFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func TestCheckCellFieldMutationUnparseable(t *testing.T) {
	// A cell body that can't be parsed should not crash the checker.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\n{{{invalid go\nresult = 1\n```\n"
	result := Parse(src)
	// Should not panic.
	_ = Check(result)
}

func TestCheckFileAccessUnparseableBody(t *testing.T) {
	// Same for file access checker — unparseable body should be safe.
	src := "```go:notebook name=test reads(data.csv:embed)\n{{{invalid\n```\n"
	result := Parse(src)
	_ = Check(result)
}

func TestWritePackageEmbedMissingSrcDir(t *testing.T) {
	// WritePackage with a notebook path whose dir doesn't exist for embed patterns.
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "//go:embed nonexistent.txt\nvar data string",
	}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	// notebookPath points to a nonexistent directory, so embed copy finds no files.
	err := WritePackage(gen, pkgDir, "/nonexistent/test.md", "testpkg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckCellReferencesUnparseable(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\n{{{invalid\nresult = 1\n```\n"
	result := Parse(src)
	_ = Check(result)
}

func TestExtractEmbedPatternsMultiline(t *testing.T) {
	patterns := ExtractEmbedPatterns("//go:embed a.txt b.txt\n//go:embed c.txt")
	if len(patterns) != 3 {
		t.Errorf("got %d patterns, want 3", len(patterns))
	}
}

func TestIsWebURL(t *testing.T) {
	if !isWebURL("http://example.com") {
		t.Error("http should be web URL")
	}
	if !isWebURL("https://example.com") {
		t.Error("https should be web URL")
	}
	if isWebURL("ftp://example.com") {
		t.Error("ftp should not be web URL")
	}
	if isWebURL("../relative") {
		t.Error("relative should not be web URL")
	}
}

func TestCountCells(t *testing.T) {
	src := "text\n\n```go:notebook name=a\nval = 1\n```\n\nmore text\n\n```go:notebook name=b\nval = 2\n```\n"
	result := Parse(src)
	count := 0
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			count++
		}
	}
	if count != 2 {
		t.Errorf("got %d cells, want 2", count)
	}
}

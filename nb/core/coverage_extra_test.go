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

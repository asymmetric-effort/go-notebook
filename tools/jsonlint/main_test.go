// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalJSON(t *testing.T) {
	input := []byte(`{"a":1,"b":[2,3]}`)
	got, err := canonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n\t\"a\": 1,\n\t\"b\": [\n\t\t2,\n\t\t3\n\t]\n}\n"
	if string(got) != want {
		t.Errorf("canonicalJSON = %q, want %q", got, want)
	}
}

func TestCanonicalJSONInvalid(t *testing.T) {
	_, err := canonicalJSON([]byte(`{bad`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLintFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.json")
	content, err := canonicalJSON([]byte(`{"name":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := lintFile(path); err != nil {
		t.Errorf("lintFile(good) = %v, want nil", err)
	}
}

func TestLintFileInvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"a": }`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := lintFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	// Should mention file and position.
	if got := err.Error(); len(got) == 0 {
		t.Error("empty error message")
	}
}

func TestLintFileNonCanonical(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ugly.json")
	// Valid but not canonically formatted.
	if err := os.WriteFile(path, []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := lintFile(path)
	if err == nil {
		t.Fatal("expected error for non-canonical JSON")
	}
}

func TestLintFileMissing(t *testing.T) {
	err := lintFile("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFindSyntaxErrorOffset(t *testing.T) {
	offset := findSyntaxErrorOffset([]byte(`{"a": }`))
	if offset <= 0 {
		t.Errorf("offset = %d, want > 0", offset)
	}
}

func TestFindSyntaxErrorOffsetValid(t *testing.T) {
	offset := findSyntaxErrorOffset([]byte(`{"a": 1}`))
	if offset != 0 {
		t.Errorf("offset = %d, want 0 for valid JSON", offset)
	}
}

func TestOffsetToLineCol(t *testing.T) {
	data := []byte("{\n  \"a\": 1\n}")
	line, col := offsetToLineCol(data, 5)
	if line != 2 {
		t.Errorf("line = %d, want 2", line)
	}
	if col < 1 {
		t.Errorf("col = %d, want >= 1", col)
	}
}

func TestOffsetToLineColZero(t *testing.T) {
	line, col := offsetToLineCol([]byte("{}"), 0)
	if line != 1 || col != 1 {
		t.Errorf("got %d:%d, want 1:1", line, col)
	}
}

func TestOffsetToLineColBeyondEnd(t *testing.T) {
	line, col := offsetToLineCol([]byte("{}"), 100)
	if line != 1 || col != 1 {
		t.Errorf("got %d:%d, want 1:1", line, col)
	}
}

func TestFindJSONFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "c.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := findJSONFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1 (node_modules should be skipped)", len(files))
	}
}

func TestFindJSONFilesSkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := findJSONFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("got %d files, want 0 (.git should be skipped)", len(files))
	}
}

func TestRunWithNoJSON(t *testing.T) {
	dir := t.TempDir()
	code := run([]string{dir})
	if code != 0 {
		t.Errorf("run() = %d, want 0 for empty directory", code)
	}
}

func TestRunWithValidJSON(t *testing.T) {
	dir := t.TempDir()
	content, _ := canonicalJSON([]byte(`{"key":"value"}`))
	if err := os.WriteFile(filepath.Join(dir, "test.json"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	code := run([]string{dir})
	if code != 0 {
		t.Errorf("run() = %d, want 0", code)
	}
}

func TestRunWithInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{bad}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run([]string{dir})
	if code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
}

func TestRunWithNonCanonicalJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ugly.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run([]string{dir})
	if code != 1 {
		t.Errorf("run() = %d, want 1", code)
	}
}

func TestRunDefaultDir(t *testing.T) {
	code := run(nil)
	// Should succeed since no .json files in tools/jsonlint.
	if code != 0 {
		t.Errorf("run(nil) = %d, want 0", code)
	}
}

func TestRunBadDir(t *testing.T) {
	code := run([]string{"/nonexistent/dir"})
	if code != 1 {
		t.Errorf("run(bad dir) = %d, want 1", code)
	}
}

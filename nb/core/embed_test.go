// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyEmbedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files.
	if err := os.WriteFile(filepath.Join(srcDir, "data.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := CopyEmbedFiles([]string{"data.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dstDir, "data.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("copied data = %q, want %q", data, "hello")
	}
}

func TestCopyEmbedFilesGlob(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	for _, name := range []string{"a.txt", "b.txt", "c.go"} {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	err := CopyEmbedFiles([]string{"*.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have copied a.txt and b.txt but not c.go.
	for _, name := range []string{"a.txt", "b.txt"} {
		if _, err := os.Stat(filepath.Join(dstDir, name)); err != nil {
			t.Errorf("missing %s", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dstDir, "c.go")); err == nil {
		t.Error("c.go should not be copied")
	}
}

func TestCopyEmbedFilesNoMatch(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	err := CopyEmbedFiles([]string{"*.nonexistent"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopyEmbedFilesSubdir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := CopyEmbedFiles([]string{"sub/file.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dstDir, "sub", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Errorf("data = %q", data)
	}
}

func TestCopyEmbedFilesBadPattern(t *testing.T) {
	err := CopyEmbedFiles([]string{"["}, t.TempDir(), t.TempDir())
	if err == nil {
		t.Error("expected error for bad glob pattern")
	}
}

func TestCopyFileSourceMissing(t *testing.T) {
	err := copyFile("/nonexistent/src", filepath.Join(t.TempDir(), "dst"))
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestCopyFileDestError(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "file.txt")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Try to copy to a path where the parent is a file, not a directory.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := copyFile(src, filepath.Join(blocker, "sub", "dst"))
	if err == nil {
		t.Error("expected error when destination dir can't be created")
	}
}

func TestCopyEmbedFilesRelError(t *testing.T) {
	// CopyEmbedFiles with actual files tests the rel path code.
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CopyEmbedFiles([]string{"x.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopyEmbedFilesMissingSrc(t *testing.T) {
	dstDir := t.TempDir()
	err := CopyEmbedFiles([]string{"file.txt"}, "/nonexistent/dir", dstDir)
	// No matches should be fine (glob returns empty).
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopyFileReadOnlyDest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only dirs work differently on Windows")
	}
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "file.txt")
	os.WriteFile(src, []byte("data"), 0o644)

	// Create read-only directory so os.Create fails.
	roDir := filepath.Join(t.TempDir(), "readonly")
	os.MkdirAll(roDir, 0o555)
	defer os.Chmod(roDir, 0o755)

	err := copyFile(src, filepath.Join(roDir, "file.txt"))
	if err == nil {
		t.Error("expected error when dest dir is read-only")
	}
}

func TestCopyFileSuccess(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "file.txt")
	os.WriteFile(src, []byte("content"), 0o644)

	dst := filepath.Join(dstDir, "copy.txt")
	err := copyFile(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "content" {
		t.Errorf("copied data = %q", data)
	}
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestExtractEmbedPatterns(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"single pattern", "//go:embed data.txt\nvar data string", []string{"data.txt"}},
		{"multiple patterns one line", "//go:embed *.txt *.csv\nvar data embed.FS", []string{"*.txt", "*.csv"}},
		{"multiple directives", "//go:embed a.txt\nvar a string\n//go:embed b.txt\nvar b string", []string{"a.txt", "b.txt"}},
		{"no directives", "var x = 1\nconst y = 2", nil},
		{"not a directive", "// go:embed data.txt\nvar data string", nil},
		{"empty source", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractEmbedPatterns(tt.src)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("pattern[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWritePackageCopiesEmbedFiles(t *testing.T) {
	// Create a notebook directory with an embed file.
	nbDir := t.TempDir()
	nbPath := filepath.Join(nbDir, "test.md")
	embedData := "embedded content"
	os.WriteFile(filepath.Join(nbDir, "data.txt"), []byte(embedData), 0o644)

	// A notebook with //go:embed in the defs block.
	src := "```go:defs\nimport _ \"embed\"\n\n//go:embed data.txt\nvar data string\n```\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(nbPath, []byte(src), 0o644)

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(t.TempDir(), "pkg")
	err = WritePackage(gen, outDir, nbPath, "testpkg")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the embed file was copied.
	copied, err := os.ReadFile(filepath.Join(outDir, "data.txt"))
	if err != nil {
		t.Fatalf("embed file not copied: %v", err)
	}
	if string(copied) != embedData {
		t.Errorf("copied data = %q, want %q", copied, embedData)
	}

	// Verify the generated source contains the embed directive.
	mainSrc, _ := os.ReadFile(filepath.Join(outDir, "main.go"))
	if !strings.Contains(string(mainSrc), "//go:embed") {
		t.Error("generated source missing //go:embed directive")
	}
}

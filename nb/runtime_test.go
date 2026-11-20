// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package nb

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestContextOpen(t *testing.T) {
	fsys := fstest.MapFS{
		"test.txt": {Data: []byte("hello")},
	}
	ctx := NewContext(fsys)
	f, err := ctx.Open("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, 5)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "hello" {
		t.Errorf("read = %q, want %q", buf[:n], "hello")
	}
}

func TestContextOpenMissing(t *testing.T) {
	fsys := fstest.MapFS{}
	ctx := NewContext(fsys)
	_, err := ctx.Open("missing.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestContextOpenNilFS(t *testing.T) {
	ctx := NewContext(nil)
	_, err := ctx.Open("test.txt")
	if err == nil {
		t.Error("expected error for nil FS")
	}
}

func TestContextDisplay(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Display("text/markdown", "# Hello")
	ctx.Display("text/markdown", "World")

	displays := ctx.Displays()
	if len(displays) != 2 {
		t.Fatalf("got %d displays, want 2", len(displays))
	}
	if displays[0].Data != "# Hello" {
		t.Errorf("displays[0].Data = %q", displays[0].Data)
	}
}

func TestContextSecret(t *testing.T) {
	ctx := NewContext(nil)
	ctx.SetSecret("db", "password123")
	if got := ctx.Secret("db"); got != "password123" {
		t.Errorf("Secret(db) = %q, want %q", got, "password123")
	}
	if got := ctx.Secret("missing"); got != "" {
		t.Errorf("Secret(missing) = %q, want empty", got)
	}
}

func TestContextCreateNoDir(t *testing.T) {
	ctx := NewContext(nil)
	_, err := ctx.Create("test.txt")
	if err == nil {
		t.Error("Create without output dir should error")
	}
}

func TestContextCreateWritesFile(t *testing.T) {
	dir := t.TempDir()
	ctx := NewContext(nil)
	ctx.SetOutputDir(dir)

	w, err := ctx.Create("output.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("hello from cell"))
	w.Close()

	data, err := os.ReadFile(filepath.Join(dir, "output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello from cell" {
		t.Errorf("file content = %q", data)
	}
}

func TestContextCreateSubdir(t *testing.T) {
	dir := t.TempDir()
	ctx := NewContext(nil)
	ctx.SetOutputDir(dir)

	w, err := ctx.Create("sub/dir/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("nested"))
	w.Close()

	data, _ := os.ReadFile(filepath.Join(dir, "sub", "dir", "file.txt"))
	if string(data) != "nested" {
		t.Errorf("nested file content = %q", data)
	}
}

func TestPanicErrorWithError(t *testing.T) {
	original := errors.New("test error")
	err := PanicError(original)
	if err == nil {
		t.Fatal("PanicError returned nil")
	}
	if !errors.Is(err, original) {
		t.Error("PanicError should wrap the original error")
	}
}

func TestPanicErrorWithString(t *testing.T) {
	err := PanicError("something went wrong")
	if err == nil {
		t.Fatal("PanicError returned nil")
	}
	if err.Error() != "panic: something went wrong" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestCellError(t *testing.T) {
	inner := errors.New("file not found")
	ce := NewCellError("load", 42, inner)
	if ce.Error() != `cell "load" failed at line 42: file not found` {
		t.Errorf("Error() = %q", ce.Error())
	}
	if !errors.Is(ce, inner) {
		t.Error("CellError should unwrap to inner error")
	}
}

func TestContextDisplaysConcurrent(t *testing.T) {
	ctx := NewContext(nil)
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			ctx.Display("text/markdown", "test")
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	displays := ctx.Displays()
	if len(displays) != 10 {
		t.Errorf("got %d displays, want 10", len(displays))
	}
}

func TestContextDisplaysEmpty(t *testing.T) {
	ctx := NewContext(nil)
	displays := ctx.Displays()
	if len(displays) != 0 {
		t.Errorf("got %d displays, want 0", len(displays))
	}
}

func TestContextSecretNoSecrets(t *testing.T) {
	ctx := NewContext(nil)
	if got := ctx.Secret("anything"); got != "" {
		t.Errorf("Secret on nil map = %q, want empty", got)
	}
}

// Verify fs.File interface satisfaction.
var _ fs.FS = fstest.MapFS{}

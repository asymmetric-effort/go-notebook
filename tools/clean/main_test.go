// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func inTmpDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
}

func TestRunCleansDirectories(t *testing.T) {
	inTmpDir(t)
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(d, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	for _, d := range dirs {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("directory %s still exists after clean", d)
		}
	}
}

func TestRunCleansCoverprofiles(t *testing.T) {
	inTmpDir(t)
	for _, f := range []string{"unit.coverprofile", "int.coverprofile"} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	matches, _ := filepath.Glob("*.coverprofile")
	if len(matches) > 0 {
		t.Errorf("coverprofiles remain: %v", matches)
	}
}

func TestRunCleansCoverageFiles(t *testing.T) {
	inTmpDir(t)
	for _, f := range coverageFiles {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	for _, f := range coverageFiles {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("file %s still exists", f)
		}
	}
}

func TestRunNoopOnCleanDirectory(t *testing.T) {
	inTmpDir(t)
	if code := run(); code != 0 {
		t.Fatalf("run() on clean directory = %d, want 0", code)
	}
}

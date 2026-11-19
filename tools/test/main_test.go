// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestRunReturnsZero(t *testing.T) {
	// Skip if already inside a test run to prevent infinite recursion.
	if os.Getenv("GOBOOK_TEST_RECURSIVE_GUARD") == "1" {
		t.Skip("skipping to prevent recursive test invocation")
	}
	t.Setenv("GOBOOK_TEST_RECURSIVE_GUARD", "1")
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func TestRunFailsOnBadPackage(t *testing.T) {
	// Verify run() returns non-zero when tests fail.
	tmp := t.TempDir()
	// Create a Go file that won't compile.
	if err := os.WriteFile(tmp+"/go.mod", []byte("module bad\ngo 1.27.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmp+"/bad_test.go", []byte("package bad\nfunc TestBad(t *testing.T) { t.Fatal(\"fail\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Build a temporary version of the test tool and run it in the bad directory.
	bin := tmp + "/testrunner"
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir, _ = os.Getwd()
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test tool: %v\n%s", err, out)
	}

	cmd := exec.Command(bin)
	cmd.Dir = tmp
	if err := cmd.Run(); err == nil {
		t.Fatal("expected non-zero exit from test tool on failing tests")
	}
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestGitHooksDir(t *testing.T) {
	dir, err := gitHooksDir()
	if err != nil {
		t.Fatalf("gitHooksDir() = _, %v", err)
	}
	if dir == "" {
		t.Fatal("gitHooksDir() returned empty string")
	}
}

func TestModuleRoot(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatalf("moduleRoot() = _, %v", err)
	}
	if root == "" {
		t.Fatal("moduleRoot() returned empty string")
	}
}

func TestRunInstallsHook(t *testing.T) {
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestGoVetPasses(t *testing.T) {
	if err := goVet(); err != nil {
		t.Fatalf("goVet() = %v, want nil", err)
	}
}

func TestGoFmtPasses(t *testing.T) {
	if err := goFmt(); err != nil {
		t.Fatalf("goFmt() = %v, want nil", err)
	}
}

func TestLintersReturnsChecks(t *testing.T) {
	checks := linters()
	if len(checks) < 5 {
		t.Fatalf("linters() returned %d checks, want at least 5", len(checks))
	}
}

func TestRunReturnsZero(t *testing.T) {
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func TestEsLintNoFiles(t *testing.T) {
	// Should pass when no TS files exist.
	if err := esLint(); err != nil {
		t.Fatalf("esLint() = %v, want nil", err)
	}
}

func TestTscCheckNoFiles(t *testing.T) {
	// Should pass when no TS files exist.
	if err := tscCheck(); err != nil {
		t.Fatalf("tscCheck() = %v, want nil", err)
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

func TestHasTypeScriptFilesNone(t *testing.T) {
	// Currently no .ts files in the project.
	if hasTypeScriptFiles() {
		t.Error("hasTypeScriptFiles() = true, want false (no TS files yet)")
	}
}

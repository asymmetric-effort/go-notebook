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
	if len(checks) < 2 {
		t.Fatalf("linters() returned %d checks, want at least 2", len(checks))
	}
	if checks[0].name != "go vet" {
		t.Errorf("first linter = %q, want %q", checks[0].name, "go vet")
	}
	if checks[1].name != "gofmt" {
		t.Errorf("second linter = %q, want %q", checks[1].name, "gofmt")
	}
}

func TestRunReturnsZero(t *testing.T) {
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

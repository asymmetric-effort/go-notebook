// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExceptionsNonexistent(t *testing.T) {
	exc, err := loadExceptions("/nonexistent/file")
	if err != nil {
		t.Fatalf("loadExceptions(nonexistent) = _, %v", err)
	}
	if exc != nil {
		t.Errorf("expected nil map, got %v", exc)
	}
}

func TestLoadExceptionsValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exceptions")
	content := `# This is a comment
GO-2024-1234
GO-2024-5678

# Another comment
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	exc, err := loadExceptions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(exc) != 2 {
		t.Fatalf("expected 2 exceptions, got %d", len(exc))
	}
	if !exc["GO-2024-1234"] {
		t.Error("missing GO-2024-1234")
	}
	if !exc["GO-2024-5678"] {
		t.Error("missing GO-2024-5678")
	}
}

func TestLoadExceptionsInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exceptions")
	if err := os.WriteFile(path, []byte("BAD-ID\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadExceptions(path)
	if err == nil {
		t.Fatal("expected error for invalid exception ID")
	}
}

func TestParseVulnIDs(t *testing.T) {
	output := `govulncheck is scanning for vulnerabilities...

Vulnerability #1: GO-2024-1234
  Description...

Vulnerability #2: GO-2024-5678
  Description...
`
	ids := parseVulnIDs(output)
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d: %v", len(ids), ids)
	}
	if ids[0] != "GO-2024-1234" {
		t.Errorf("ids[0] = %q, want %q", ids[0], "GO-2024-1234")
	}
	if ids[1] != "GO-2024-5678" {
		t.Errorf("ids[1] = %q, want %q", ids[1], "GO-2024-5678")
	}
}

func TestParseVulnIDsNoVulns(t *testing.T) {
	ids := parseVulnIDs("No vulnerabilities found.")
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(ids))
	}
}

func TestLoadExceptionsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exceptions")
	if err := os.WriteFile(path, []byte("# only comments\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	exc, err := loadExceptions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(exc) != 0 {
		t.Errorf("expected 0 exceptions, got %d", len(exc))
	}
}

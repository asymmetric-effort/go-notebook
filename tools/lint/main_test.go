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
	if len(checks) < 9 {
		t.Fatalf("linters() returned %d checks, want at least 9", len(checks))
	}
}

func TestRunReturnsZero(t *testing.T) {
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func TestMarkdownLintPasses(t *testing.T) {
	if err := markdownLint(); err != nil {
		t.Fatalf("markdownLint() = %v, want nil", err)
	}
}

func TestCheckMakePasses(t *testing.T) {
	if err := checkMake(); err != nil {
		t.Fatalf("checkMake() = %v, want nil", err)
	}
}

func TestYamlFmtNoFiles(t *testing.T) {
	if err := yamlFmt(); err != nil {
		t.Fatalf("yamlFmt() = %v, want nil", err)
	}
}

func TestXmlLintNoFiles(t *testing.T) {
	if err := xmlLint(); err != nil {
		t.Fatalf("xmlLint() = %v, want nil", err)
	}
}

func TestEsLintNoFiles(t *testing.T) {
	if err := esLint(); err != nil {
		t.Fatalf("esLint() = %v, want nil", err)
	}
}

func TestTscCheckNoFiles(t *testing.T) {
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

func TestHasTypeScriptFiles(t *testing.T) {
	// The repo has TypeScript files (nb/server/static/editor.ts).
	if !hasTypeScriptFiles() {
		t.Error("hasTypeScriptFiles() = false, want true")
	}
}

func TestFindFilesByExtMd(t *testing.T) {
	root, _ := moduleRoot()
	files := findFilesByExt(root, ".md", nil)
	if len(files) == 0 {
		t.Error("findFilesByExt(.md) returned no files")
	}
}

func TestFindFilesByExtNoMatch(t *testing.T) {
	root, _ := moduleRoot()
	files := findFilesByExt(root, ".nonexistent", nil)
	if len(files) != 0 {
		t.Errorf("findFilesByExt(.nonexistent) = %d files, want 0", len(files))
	}
}

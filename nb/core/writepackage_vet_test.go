// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestWritePackageVetClean verifies that generated code produces no go vet findings.
func TestWritePackageVetClean(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 42\n```\n\n" +
		"```go:notebook name=b exported(result int)\nresult = a.val + 1\n```\n\n" +
		"```go:notebook\n_ = b.result\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "vetpkg")
	WritePackage(gen, pkgDir, "test.md", "vetpkg")

	moduleRoot := findModuleRoot(t)
	goMod := "module vetpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		genSrc, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
		t.Fatalf("go vet found issues:\n%s\n---generated---\n%s", out, genSrc)
	}
	t.Log("go vet passed — no findings on generated code")
}

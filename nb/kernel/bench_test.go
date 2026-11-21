// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

// BenchmarkGoRunColdCache measures the cost of `go run` on a simple cell
// with a cold build cache. This is the worst case — first run after changing
// a cell's code or imports.
func BenchmarkGoRunColdCache(b *testing.B) {
	pkgDir := setupBenchPackage(b, "```go:notebook name=test exported(val int)\nval = 42\n```\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear the Go build cache for cold measurement.
		exec.Command("go", "clean", "-cache").Run()

		result, err := GoRun(context.Background(), pkgDir)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("exit %d: %s", result.ExitCode, result.Stderr)
		}
	}
}

// BenchmarkGoRunWarmCache measures the cost of `go run` on a simple cell
// with a warm build cache. This is the common case — rerunning an unchanged cell.
func BenchmarkGoRunWarmCache(b *testing.B) {
	pkgDir := setupBenchPackage(b, "```go:notebook name=test exported(val int)\nval = 42\n```\n")

	// Warm the cache with one run.
	GoRun(context.Background(), pkgDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := GoRun(context.Background(), pkgDir)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("exit %d: %s", result.ExitCode, result.Stderr)
		}
	}
}

// BenchmarkGoRunWithImports measures the cost with common imports.
func BenchmarkGoRunWithImports(b *testing.B) {
	src := "```go:defs\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n```\n\n" +
		"```go:notebook name=test exported(val string)\nval = strings.ToUpper(\"hello\")\nfmt.Println(val)\n```\n"
	pkgDir := setupBenchPackage(b, src)

	// Warm cache.
	GoRun(context.Background(), pkgDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := GoRun(context.Background(), pkgDir)
		if err != nil {
			b.Fatal(err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("exit %d: %s", result.ExitCode, result.Stderr)
		}
	}
}

func setupBenchPackage(b *testing.B, notebookSrc string) string {
	b.Helper()
	result := core.Parse(notebookSrc)
	gen, err := core.Generate(result, "main")
	if err != nil {
		b.Fatal(err)
	}

	dir := b.TempDir()
	pkgDir := filepath.Join(dir, "bench")
	if err := core.WritePackage(gen, pkgDir, "bench.md", "bench"); err != nil {
		b.Fatal(err)
	}

	moduleRoot := findBenchModuleRoot(b)
	goMod := "module bench\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	return pkgDir
}

func findBenchModuleRoot(b *testing.B) string {
	b.Helper()
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		b.Fatal(err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

// TestBenchmarkSmoke just verifies the benchmark setup works.
func TestBenchmarkSmoke(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	result := core.Parse(src)
	gen, err := core.Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "smoke")
	core.WritePackage(gen, pkgDir, "smoke.md", "smoke")

	moduleRoot := findModuleRoot(t)
	goMod := "module smoke\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cellResult, err := GoRun(ctx, pkgDir)
	if err != nil {
		t.Fatal(err)
	}
	if cellResult.ExitCode != 0 {
		t.Errorf("exit %d: %s", cellResult.ExitCode, cellResult.Stderr)
	}
}

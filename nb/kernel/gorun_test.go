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

// TestGoRunCellEndToEnd is the critical integration test: it parses a
// notebook, generates a Go package, runs `go run` on it, and verifies
// the process completes successfully.
func TestGoRunCellEndToEnd(t *testing.T) {
	// A simple notebook with one cell that prints to stdout.
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n"

	result := core.Parse(src)
	gen, err := core.Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cellpkg")
	moduleRoot := findModuleRoot(t)

	err = core.WritePackage(gen, pkgDir, "test.md", "cellpkg")
	if err != nil {
		t.Fatal(err)
	}

	// Add replace directive for local module.
	goMod := "module cellpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	// Run via the kernel.
	k := New()
	cellResult, err := k.RunCell(context.Background(), "test", 30*time.Second,
		func(ctx context.Context) (CellResult, error) {
			return GoRun(ctx, pkgDir)
		})

	if err != nil {
		t.Fatalf("RunCell: %v", err)
	}
	if cellResult.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0\nStderr: %s", cellResult.ExitCode, cellResult.Stderr)
	}
	t.Logf("Cell ran successfully. Stdout: %q, Stderr: %q", cellResult.Stdout, cellResult.Stderr)
}

// TestGoRunCellTimeout verifies that a slow cell is killed by the timeout.
func TestGoRunCellTimeout(t *testing.T) {
	// Create a package that sleeps forever.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "slowpkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module slowpkg\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte(`package main
import "time"
func main() { time.Sleep(10 * time.Minute) }
`), 0o644)

	k := New()
	_, err := k.RunCell(context.Background(), "slow", 200*time.Millisecond,
		func(ctx context.Context) (CellResult, error) {
			return GoRun(ctx, pkgDir)
		})

	if err == nil {
		t.Error("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %q, want timeout", err.Error())
	}
}

// TestGoRunCellPanicRecovered verifies a panicking cell doesn't crash.
func TestGoRunCellPanicRecovered(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\npanic(\"intentional\")\n```\n"

	result := core.Parse(src)
	gen, err := core.Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "panicpkg")
	moduleRoot := findModuleRoot(t)

	core.WritePackage(gen, pkgDir, "test.md", "panicpkg")
	goMod := "module panicpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	k := New()
	cellResult, err := k.RunCell(context.Background(), "test", 30*time.Second,
		func(ctx context.Context) (CellResult, error) {
			return GoRun(ctx, pkgDir)
		})

	if err != nil {
		t.Fatalf("RunCell error: %v", err)
	}
	// The binary should exit 0 (panic recovered) with no stack trace.
	if cellResult.ExitCode != 0 {
		// It's ok if exit code is non-zero due to the error, but no stack trace.
		if strings.Contains(cellResult.Stderr, "goroutine") {
			t.Error("stack trace in stderr — panic was not recovered")
		}
	}
	t.Logf("Panic cell: exit=%d stdout=%q stderr=%q", cellResult.ExitCode, cellResult.Stdout, cellResult.Stderr)
}

// TestGoRunCellExitError verifies a cell that exits non-zero.
func TestGoRunCellExitError(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "exitpkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module exitpkg\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\nimport \"os\"\nfunc main() { os.Exit(1) }\n"), 0o644)

	result, err := GoRun(context.Background(), pkgDir)
	if err != nil {
		t.Fatalf("GoRun returned error (should return CellResult with exit code): %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
}

// TestGoRunCellCompileError verifies a package that doesn't compile.
func TestGoRunCellCompileError(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "badpkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module badpkg\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\nfunc main() { undefined() }\n"), 0o644)

	result, err := GoRun(context.Background(), pkgDir)
	if err != nil {
		t.Logf("GoRun error (expected for compile failure): %v", err)
		return
	}
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit for compile error")
	}
}

// TestGoRunCellBadDir verifies error when directory doesn't exist.
func TestGoRunCellBadDir(t *testing.T) {
	_, err := GoRun(context.Background(), "/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// TestGoRunContextCancel verifies the ctx.Err() path in GoRun.
func TestGoRunContextCancel(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "sleeppkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module sleeppkg\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	// Cancel context immediately — go run will fail because it's already cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GoRun(ctx, pkgDir)
	if err == nil {
		t.Fatal("expected error from GoRun with cancelled context")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("error = %q, want timeout or context error", err)
	}
}

// TestGoRunWithEnvContextCancel verifies the ctx.Err() path in GoRunWithEnv.
func TestGoRunWithEnvContextCancel(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "sleeppkg2")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module sleeppkg2\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GoRunWithEnv(ctx, pkgDir, []string{"FOO=bar"})
	if err == nil {
		t.Fatal("expected error from GoRunWithEnv with cancelled context")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("error = %q, want timeout or context error", err)
	}
}

// TestGoRunWithEnvBadDir verifies error when directory doesn't exist.
func TestGoRunWithEnvBadDir(t *testing.T) {
	_, err := GoRunWithEnv(context.Background(), "/nonexistent/path", nil)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// buildFixture creates a standalone Go program in a temp dir with its own
// go.mod, builds it, and returns the binary path.
func buildFixture(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	pkg := filepath.Join(dir, name)
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "go.mod"), []byte("module "+name+"\ngo 1.27.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, name+"-bin")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = pkg
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build fixture %s: %v\n%s", name, err, out)
	}
	return bin
}

func TestRunCLICapturesStdout(t *testing.T) {
	bin := buildFixture(t, "stdout", `package main
import ("fmt"; "os")
func main() { fmt.Fprint(os.Stdout, "hello") }
`)
	result, err := RunCLI(bin, nil, "", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "hello")
	}
}

func TestRunCLICapturesStderr(t *testing.T) {
	bin := buildFixture(t, "stderr", `package main
import ("fmt"; "os")
func main() { fmt.Fprint(os.Stderr, "oops"); os.Exit(1) }
`)
	result, err := RunCLI(bin, nil, "", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", result.ExitCode)
	}
	if result.Stderr != "oops" {
		t.Errorf("stderr = %q, want %q", result.Stderr, "oops")
	}
}

func TestRunCLISuppliesStdin(t *testing.T) {
	bin := buildFixture(t, "stdin", `package main
import ("fmt"; "io"; "os")
func main() { d, _ := io.ReadAll(os.Stdin); fmt.Fprint(os.Stdout, string(d)) }
`)
	result, err := RunCLI(bin, nil, "input data", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if result.Stdout != "input data" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "input data")
	}
}

func TestRunCLITimeout(t *testing.T) {
	bin := buildFixture(t, "sleeper", `package main
import "time"
func main() { time.Sleep(10 * time.Minute) }
`)
	_, err := RunCLI(bin, nil, "", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunCLIDefaultTimeout(t *testing.T) {
	bin := buildFixture(t, "quick", `package main
func main() {}
`)
	result, err := RunCLI(bin, nil, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestRunCLIWithEnv(t *testing.T) {
	bin := buildFixture(t, "envfix", `package main
import ("fmt"; "os")
func main() { fmt.Fprint(os.Stdout, os.Getenv("TEST_VAR")) }
`)
	result, err := RunCLIWithEnv(bin, nil, "", 10*time.Second, []string{"TEST_VAR=hello_env"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Stdout != "hello_env" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "hello_env")
	}
}

func TestRunCLIWithEnvDir(t *testing.T) {
	bin := buildFixture(t, "dirfix", `package main
import ("fmt"; "os")
func main() { wd, _ := os.Getwd(); fmt.Fprint(os.Stdout, wd) }
`)
	workDir := t.TempDir()
	result, err := RunCLIWithEnv(bin, nil, "", 10*time.Second, nil, workDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Stdout != workDir {
		t.Errorf("stdout = %q, want %q", result.Stdout, workDir)
	}
}

func TestRunCLIWithEnvDefaultTimeout(t *testing.T) {
	bin := buildFixture(t, "quick2", `package main
func main() {}
`)
	result, err := RunCLIWithEnv(bin, nil, "", 0, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestRunCLIWithEnvTimeout(t *testing.T) {
	bin := buildFixture(t, "sleeper2", `package main
import "time"
func main() { time.Sleep(10 * time.Minute) }
`)
	_, err := RunCLIWithEnv(bin, nil, "", 100*time.Millisecond, nil, "")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunCLINonexistentBinary(t *testing.T) {
	_, err := RunCLI("/nonexistent/binary", nil, "", time.Second)
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestRunCLIWithEnvNonexistentBinary(t *testing.T) {
	_, err := RunCLIWithEnv("/nonexistent/binary", nil, "", time.Second, nil, "")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestRunCLIWithEnvStdin(t *testing.T) {
	bin := buildFixture(t, "stdin2", `package main
import ("fmt"; "io"; "os")
func main() { d, _ := io.ReadAll(os.Stdin); fmt.Fprint(os.Stdout, string(d)) }
`)
	result, err := RunCLIWithEnv(bin, nil, "env_input", 10*time.Second, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Stdout != "env_input" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "env_input")
	}
}

func TestExitCodeWithExitError(t *testing.T) {
	// Run a command that exits with code 42.
	bin := buildFixture(t, "exit42", `package main
import "os"
func main() { os.Exit(42) }
`)
	cmd := exec.Command(bin)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error")
	}
	if code := exitCode(err); code != 42 {
		t.Errorf("exitCode() = %d, want 42", code)
	}
}

func TestExitCodeWithNonExitError(t *testing.T) {
	err := fmt.Errorf("not an exit error")
	if code := exitCode(err); code != -1 {
		t.Errorf("exitCode() = %d, want -1", code)
	}
}

func TestBuildTestBinaryFailsOnBadPackage(t *testing.T) {
	dir := t.TempDir()
	_, err := BuildTestBinary(filepath.Join(dir, "nonexistent"), dir)
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}

func TestBuildTestBinarySuccess(t *testing.T) {
	// Build one of the existing tools to test the success path.
	dir := t.TempDir()
	bin, err := BuildTestBinary("github.com/asymmetric-effort/go-notebook/tools/clean", dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Errorf("binary not found at %s: %v", bin, err)
	}
}

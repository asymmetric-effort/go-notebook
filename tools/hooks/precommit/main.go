// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// precommit is the compiled pre-commit hook for the go-notebook repository.
// It runs leakdetector, make lint and make test, failing the commit if any
// check fails.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// check represents a named pre-commit check.
type check struct {
	name string
	fn   func() error
}

// checks returns the ordered list of pre-commit checks.
// leakdetector runs first so secrets are caught before other checks.
func checks() []check {
	return []check{
		{name: "leakdetector", fn: runLeakDetector},
		{name: "make lint", fn: runMakeLint},
		{name: "make test", fn: runMakeTest},
	}
}

// runLeakDetector runs leakdetector --skip-history.
func runLeakDetector() error {
	cmd := exec.Command("leakdetector", "--skip-history")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("leakdetector failed: %w", err)
	}
	return nil
}

// runMakeLint runs make lint.
func runMakeLint() error {
	cmd := exec.Command("make", "lint")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make lint failed: %w", err)
	}
	return nil
}

// runMakeTest runs make test.
func runMakeTest() error {
	cmd := exec.Command("make", "test")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make test failed: %w", err)
	}
	return nil
}

// expandPath adds common Go and local tool directories to PATH so the
// hook finds tools installed via go install or placed in ~/.local/bin.
func expandPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	extra := []string{
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".local", "bin"),
	}
	current := os.Getenv("PATH")
	for _, dir := range extra {
		if !strings.Contains(current, dir) {
			current = dir + string(os.PathListSeparator) + current
		}
	}
	os.Setenv("PATH", current)
}

func run() int {
	// Ensure common tool paths are in PATH for git hooks which run
	// with a minimal environment.
	expandPath()

	for _, c := range checks() {
		fmt.Printf("pre-commit: running %s...\n", c.name)
		if err := c.fn(); err != nil {
			fmt.Fprintf(os.Stderr, "pre-commit: %v\n", err)
			return 1
		}
	}
	fmt.Println("pre-commit: all checks passed")
	return 0
}

func main() {
	os.Exit(run())
}

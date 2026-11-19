// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// hooks builds and installs the pre-commit hook.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// hookSource is the Go source for the pre-commit hook.
const hookSource = `tools/hooks/precommit`

// gitHooksDir returns the path to the .git/hooks directory.
func gitHooksDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("find git dir: %w", err)
	}
	return filepath.Join(strings.TrimSpace(string(out)), "hooks"), nil
}

// moduleRoot returns the root directory of the Go module.
func moduleRoot() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	modFile := strings.TrimSpace(string(out))
	if modFile == "" || modFile == os.DevNull {
		return "", fmt.Errorf("not inside a Go module")
	}
	return filepath.Dir(modFile), nil
}

func run() int {
	hooksDir, err := gitHooksDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create hooks dir: %v\n", err)
		return 1
	}

	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	// Build the pre-commit hook binary.
	hookBin := filepath.Join(hooksDir, "pre-commit")
	cmd := exec.Command("go", "build", "-o", hookBin, "./"+hookSource)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: build hook: %v\n", err)
		return 1
	}

	fmt.Printf("hooks: installed pre-commit to %s\n", hookBin)
	return 0
}

func main() {
	os.Exit(run())
}

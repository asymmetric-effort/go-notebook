// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// lint runs all linters configured for the repository.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// linter represents a single lint check.
type linter struct {
	name string
	fn   func() error
}

// linters returns the ordered list of lint checks.
func linters() []linter {
	return []linter{
		{name: "go vet", fn: goVet},
		{name: "gofmt", fn: goFmt},
		{name: "jsonlint", fn: jsonLint},
	}
}

// goVet runs go vet on all packages.
func goVet() error {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}
	return nil
}

// goFmt checks that all Go files are gofmt-formatted.
func goFmt() error {
	cmd := exec.Command("gofmt", "-l", ".")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}
	unformatted := strings.TrimSpace(string(out))
	if unformatted != "" {
		return fmt.Errorf("unformatted files:\n%s", unformatted)
	}
	return nil
}

// jsonLint runs the in-repo JSON linter on the project root.
func jsonLint() error {
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	cmd := exec.Command("go", "run", "./tools/jsonlint", root)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("jsonlint failed: %w", err)
	}
	return nil
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
	failed := false
	for _, l := range linters() {
		if err := l.fn(); err != nil {
			fmt.Fprintf(os.Stderr, "lint %s: %v\n", l.name, err)
			failed = true
		}
	}
	if failed {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// lint runs all linters configured for the repository.
package main

import (
	"fmt"
	"os"
	"os/exec"
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

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package testutil provides shared test helpers for the go-notebook project.
// It is internal and must not be imported outside the module.
package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CLIResult holds the captured output of a command invocation.
type CLIResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// exitCode extracts the exit code from an exec error. If the error is not an
// ExitError, it returns -1.
func exitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// RunCLI executes a built binary with the given arguments and optional stdin,
// returning the captured stdout, stderr and exit code. It uses only standard
// go test tooling with no third-party test frameworks.
func RunCLI(binary string, args []string, stdin string, timeout time.Duration) (CLIResult, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	cmd := exec.Command(binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return CLIResult{}, fmt.Errorf("start %s: %w", binary, err)
	}

	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		result := CLIResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
		if err != nil {
			result.ExitCode = exitCode(err)
		}
		return result, nil
	case <-time.After(timeout):
		cmd.Process.Kill()
		return CLIResult{}, fmt.Errorf("%s timed out after %v", binary, timeout)
	}
}

// RunCLIWithEnv is like RunCLI but also sets environment variables and a
// working directory for the command.
func RunCLIWithEnv(binary string, args []string, stdin string, timeout time.Duration, env []string, dir string) (CLIResult, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	cmd := exec.Command(binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), env...)

	if dir != "" {
		cmd.Dir = dir
	}

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return CLIResult{}, fmt.Errorf("start %s: %w", binary, err)
	}

	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		result := CLIResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
		if err != nil {
			result.ExitCode = exitCode(err)
		}
		return result, nil
	case <-time.After(timeout):
		cmd.Process.Kill()
		return CLIResult{}, fmt.Errorf("%s timed out after %v", binary, timeout)
	}
}

// BuildTestBinary compiles a Go main package into a temporary binary and
// returns the path. The caller should defer removal or use t.TempDir().
func BuildTestBinary(pkg string, outputDir string) (string, error) {
	name := filepath.Base(pkg)
	bin := filepath.Join(outputDir, name)
	cmd := exec.Command("go", "build", "-o", bin, pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build %s: %w\n%s", pkg, err, out)
	}
	return bin, nil
}

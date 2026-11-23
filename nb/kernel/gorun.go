// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// GoRun executes `go run .` in the given package directory and returns
// the cell result. This is the actual process execution that the kernel
// uses to run notebook cells.
//
// The context controls the timeout — if it expires, the process is killed.
func GoRun(ctx context.Context, pkgDir string) (CellResult, error) {
	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CellResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if ctx.Err() != nil {
			// Context was cancelled (timeout).
			return result, fmt.Errorf("cell timed out: %w", ctx.Err())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil // Non-zero exit is a cell error, not a system error.
		}
		return result, fmt.Errorf("go run failed: %w", err)
	}

	return result, nil
}

// GoRunWithEnv executes `go run .` with additional environment variables.
func GoRunWithEnv(ctx context.Context, pkgDir string, extraEnv []string) (CellResult, error) {
	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	cmd.Env = append(cmd.Env, extraEnv...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CellResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if ctx.Err() != nil {
			return result, fmt.Errorf("cell timed out: %w", ctx.Err())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return result, fmt.Errorf("go run failed: %w", err)
	}

	return result, nil
}

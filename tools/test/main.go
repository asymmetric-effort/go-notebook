// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// test runs the project test suite with coverage enforcement.
// The coverage gate is added by story e0.s9 (#11).
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func run() int {
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: tests failed: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}

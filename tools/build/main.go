// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// build compiles the project binaries.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func run() int {
	cmd := exec.Command("go", "build", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: build failed: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}

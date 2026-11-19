// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// clean removes build artifacts and generated files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// dirs lists directories removed by clean.
var dirs = []string{"bin", "dist", "gen"}

// coverageFiles lists individual files removed by clean.
var coverageFiles = []string{"coverage.out", "coverage.html"}

func run() int {
	for _, d := range dirs {
		if err := os.RemoveAll(d); err != nil {
			fmt.Fprintf(os.Stderr, "error: clean %s: %v\n", d, err)
			return 1
		}
	}
	matches, err := filepath.Glob("*.coverprofile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: clean coverprofiles: %v\n", err)
		return 1
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			fmt.Fprintf(os.Stderr, "error: clean %s: %v\n", m, err)
			return 1
		}
	}
	for _, f := range coverageFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: clean %s: %v\n", f, err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run())
}

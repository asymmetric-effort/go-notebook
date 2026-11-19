// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// gobook is the CLI for go-notebook.
package main

import (
	"fmt"
	"os"

	"github.com/asymmetric-effort/go-notebook/cmd/gobook/commands"
)

func run() int {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}

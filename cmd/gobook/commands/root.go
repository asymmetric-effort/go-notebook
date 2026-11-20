// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package commands implements the gobook CLI commands.
package commands

import (
	"fmt"
	"os"

	"github.com/asymmetric-effort/go-notebook/nb/build"
	"github.com/asymmetric-effort/go-notebook/nb/core"
)

// Execute runs the gobook CLI. For now this is a simple dispatch without
// cobra, which will be added when the report binary needs it.
func Execute() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: gobook <command> [args]\n\nCommands: build, check, gen, generate, run, fmt, vet")
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "build":
		return cmdBuild(args)
	case "check":
		return cmdCheck(args)
	case "fmt":
		return cmdFmt(args)
	case "gen", "generate":
		return cmdGen(args)
	case "run":
		return cmdRun(args)
	case "vet":
		return cmdVet(args)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func cmdBuild(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	opts := build.Options{}
	// Parse flags: --format, -o, --cell-timeout
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				opts.Format = args[i+1]
				i++
			}
		case "-o":
			if i+1 < len(args) {
				opts.Output = args[i+1]
				i++
			}
		case "--cell-timeout":
			if i+1 < len(args) {
				opts.Timeout = args[i+1]
				i++
			}
		}
	}
	return build.Run(path, opts)
}

func requireNotebookArg(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("usage: gobook <command> <notebook.md>")
	}
	return args[0], nil
}

func cmdCheck(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}
	return nil
}

func cmdFmt(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	formatted, err := core.FormatNotebook(string(src))
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	if formatted != string(src) {
		if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func cmdGen(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}
	// Print generated code summary.
	fmt.Printf("Generated %d cells from %s\n", len(gen.Cells), path)
	for _, c := range gen.Cells {
		if c.StructName != "" {
			fmt.Printf("  %s -> %s\n", c.FuncName, c.StructName)
		} else {
			fmt.Printf("  %s (display only)\n", c.FuncName)
		}
	}
	return nil
}

func cmdRun(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	// For now, run means: parse, check, generate, report.
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}
	fmt.Printf("gobook run: %s checked successfully (%d cells)\n", path, countCells(result))
	return nil
}

func cmdVet(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d findings", len(errs))
	}
	return nil
}

func countCells(result core.ParseResult) int {
	n := 0
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			n++
		}
	}
	return n
}

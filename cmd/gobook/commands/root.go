// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package commands implements the gobook CLI commands.
package commands

import (
	"fmt"
	"os"

	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/build"
	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/kernel"
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

	// Determine output directory.
	outDir := "gen"
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--out" {
			outDir = args[i+1]
		}
	}

	// Write the generated package to disk.
	if err := core.WritePackage(gen, outDir, path, "gobook-gen"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	fmt.Printf("gobook gen: wrote %d cells to %s/\n", len(gen.Cells), outDir)
	return nil
}

func cmdRun(args []string) error {
	path, err := requireNotebookArg(args)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	// Parse and check.
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}

	// Generate the Go package.
	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}

	// Write to a temp directory.
	tmpDir, err := os.MkdirTemp("", "gobook-run-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, path, "gobook-run"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	// Add replace directive pointing to the notebook's module.
	moduleRoot := findModuleRootFromPath(path)
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	// Execute via the kernel.
	k := kernel.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cellResult, err := k.RunCell(ctx, "all", 5*time.Minute,
		func(ctx context.Context) (kernel.CellResult, error) {
			return kernel.GoRun(ctx, pkgDir)
		})
	if err != nil {
		return fmt.Errorf("run %s: %w", path, err)
	}

	// Print output.
	if cellResult.Stdout != "" {
		fmt.Print(cellResult.Stdout)
	}
	if cellResult.Stderr != "" {
		fmt.Fprint(os.Stderr, cellResult.Stderr)
	}
	if cellResult.ExitCode != 0 {
		return fmt.Errorf("cell exited with status %d", cellResult.ExitCode)
	}

	fmt.Fprintf(os.Stderr, "gobook run: %s completed successfully\n", path)
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

	// First run the checker.
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check findings", len(errs))
	}

	// Generate the package and run go vet on it.
	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}

	tmpDir, err := os.MkdirTemp("", "gobook-vet-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, path, "gobook-vet"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	moduleRoot := findModuleRootFromPath(path)
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))
		return fmt.Errorf("go vet found issues")
	}

	fmt.Fprintf(os.Stderr, "gobook vet: %s passed\n", path)
	return nil
}

// findModuleRootFromPath finds the go-notebook module root. For now, uses
// `go env GOMOD` which works when gobook is run from within the module.
func findModuleRootFromPath(notebookPath string) string {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "."
	}
	modFile := strings.TrimSpace(string(out))
	if modFile == "" || modFile == os.DevNull {
		return "."
	}
	return filepath.Dir(modFile)
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

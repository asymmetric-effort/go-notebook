// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// lint runs all linters configured for the repository.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// isWindows returns true if running on Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

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
		{name: "markdownlint", fn: markdownLint},
		{name: "checkmake", fn: checkMake},
		{name: "yamlfmt", fn: yamlFmt},
		{name: "xmllint", fn: xmlLint},
		{name: "eslint", fn: esLint},
		{name: "tsc", fn: tscCheck},
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

// markdownLint runs markdownlint-cli2 on Markdown files.
func markdownLint() error {
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	files := findFilesByExt(root, ".md", nil)
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"markdownlint-cli2"}, files...)
	cmd := exec.Command("npx", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("markdownlint failed: %w", err)
	}
	return nil
}

// checkMake runs checkmake on Makefile and .mk files. Skips if checkmake is not installed.
func checkMake() error {
	if _, err := exec.LookPath("checkmake"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: checkmake not found, skipping Makefile lint\n")
		return nil
	}
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	mainMakefile := filepath.Join(root, "Makefile")
	if _, err := os.Stat(mainMakefile); err == nil {
		cmd := exec.Command("checkmake", mainMakefile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("checkmake Makefile failed: %w", err)
		}
	}
	return nil
}

// yamlFmt runs yamlfmt -lint on YAML files. Skips if yamlfmt is not installed
// or on Windows where CRLF handling causes false positives.
func yamlFmt() error {
	if _, err := exec.LookPath("yamlfmt"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: yamlfmt not found, skipping YAML lint\n")
		return nil
	}
	if isWindows() {
		fmt.Fprintf(os.Stderr, "warning: skipping yamlfmt on Windows (CRLF issues)\n")
		return nil
	}
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	files := findFilesByExt(root, ".yaml", []string{".yml"})
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"-lint"}, files...)
	cmd := exec.Command("yamlfmt", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yamlfmt failed: %w", err)
	}
	return nil
}

// xmlLint runs xmllint on XML and SVG files. Skips if xmllint is not installed.
func xmlLint() error {
	if _, err := exec.LookPath("xmllint"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: xmllint not found, skipping XML/SVG lint\n")
		return nil
	}
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	files := findFilesByExt(root, ".xml", []string{".svg"})
	if len(files) == 0 {
		return nil
	}
	for _, f := range files {
		cmd := exec.Command("xmllint", "--noout", f)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("xmllint %s failed: %w", f, err)
		}
	}
	return nil
}

// findFilesByExt walks root and returns files matching the primary extension
// or any of the extra extensions. It skips node_modules, .git and vendor.
func findFilesByExt(root string, primary string, extra []string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if info.IsDir() && (base == "node_modules" || base == ".git" || base == "vendor" || base == "target" || base == "testdata") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == primary {
			files = append(files, path)
			return nil
		}
		for _, e := range extra {
			if ext == e {
				files = append(files, path)
				return nil
			}
		}
		return nil
	})
	return files
}

// hasTypeScriptFiles checks if any .ts files exist in the project.
func hasTypeScriptFiles() bool {
	root, err := moduleRoot()
	if err != nil {
		return false
	}
	found := false
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if info.IsDir() && (base == "node_modules" || base == ".git") {
			return filepath.SkipDir
		}
		if filepath.Ext(path) == ".ts" {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// esLint runs ESLint on TypeScript files.
func esLint() error {
	if !hasTypeScriptFiles() {
		return nil
	}
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	cmd := exec.Command("npx", "eslint", ".")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("eslint failed: %w", err)
	}
	return nil
}

// tscCheck runs tsc --noEmit for type checking.
func tscCheck() error {
	if !hasTypeScriptFiles() {
		return nil
	}
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	cmd := exec.Command("npx", "tsc", "--noEmit")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tsc failed: %w", err)
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

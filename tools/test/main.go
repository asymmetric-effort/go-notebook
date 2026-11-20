// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// test runs the project test suite with coverage enforcement.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// coverageFloor is the minimum aggregate line coverage percentage.
// The goal is 100%; 98% is the enforced floor.
const coverageFloor = 98.0

// excludePrefixes lists package prefixes excluded from coverage enforcement.
// Tool main packages contain os.Exit calls that cannot be tested in-process.
var excludePrefixes = []string{
	"github.com/asymmetric-effort/go-notebook/tools/",
	"github.com/asymmetric-effort/go-notebook/node_modules/",
	"github.com/asymmetric-effort/go-notebook/cmd/",
	"github.com/asymmetric-effort/go-notebook/nb/engine/",
	"github.com/asymmetric-effort/go-notebook/nb/crypto/",
	// nb/secrets.go has terminal-dependent code (term.ReadPassword)
	"github.com/asymmetric-effort/go-notebook/nb/secrets.go",
}

func run() int {
	coverFile := "coverage.out"
	// List packages explicitly, excluding node_modules which may contain
	// Go packages from npm dependencies.
	pkgsCmd := exec.Command("go", "list", "./...")
	pkgsOut, err := pkgsCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list packages: %v\n", err)
		return 1
	}
	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(pkgsOut)), "\n") {
		if line != "" && !strings.Contains(line, "/node_modules/") {
			pkgs = append(pkgs, line)
		}
	}
	if len(pkgs) == 0 {
		fmt.Println("no packages to test")
		return 0
	}

	args := append([]string{"test", "-coverprofile", coverFile}, pkgs...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: tests failed: %v\n", err)
		return 1
	}

	pct, err := aggregateCoverage(coverFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: parse coverage: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "\naggregate coverage: %.1f%%\n", pct)

	if pct < coverageFloor {
		fmt.Fprintf(os.Stderr, "error: aggregate coverage %.1f%% is below %.0f%% floor\n", pct, coverageFloor)
		return 1
	}

	return 0
}

// isExcluded returns true if the file belongs to an excluded package.
func isExcluded(filePath string) bool {
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(filePath, prefix) {
			return true
		}
	}
	return false
}

// aggregateCoverage reads a Go coverage profile and computes the aggregate
// statement coverage percentage. Files from excluded prefixes are skipped.
func aggregateCoverage(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var total, covered int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		// field[0] is "file:startLine.startCol,endLine.endCol"
		filePath := strings.SplitN(fields[0], ":", 2)[0]
		if isExcluded(filePath) {
			continue
		}
		stmts, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		total += stmts
		if count > 0 {
			covered += stmts
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan %s: %w", path, err)
	}
	if total == 0 {
		return 100.0, nil
	}
	return float64(covered) / float64(total) * 100, nil
}

func main() {
	os.Exit(run())
}

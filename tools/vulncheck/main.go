// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// vulncheck runs govulncheck and filters results against an exceptions file.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// defaultExceptionsFile is the path to the exceptions file relative to the
// module root.
const defaultExceptionsFile = ".govulncheck-exceptions"

// loadExceptions reads the exceptions file and returns a set of excepted
// vulnerability IDs. Lines starting with # are comments. Empty lines are
// ignored.
func loadExceptions(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open exceptions: %w", err)
	}
	defer f.Close()

	exceptions := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Each non-comment line must be a vulnerability ID (e.g., GO-2024-1234).
		if !strings.HasPrefix(line, "GO-") {
			return nil, fmt.Errorf("%s:%d: invalid exception %q (must start with GO-)", path, lineNum, line)
		}
		exceptions[line] = true
	}
	return exceptions, scanner.Err()
}

// runGovulncheck runs govulncheck and returns its exit code, stdout and stderr.
func runGovulncheck() (int, string, string) {
	cmd := exec.Command("govulncheck", "./...")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return exitCode, stdout.String(), stderr.String()
}

// parseVulnIDs extracts vulnerability IDs from govulncheck output.
// IDs appear as lines like "Vulnerability #1: GO-2024-1234"
func parseVulnIDs(output string) []string {
	var ids []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Vulnerability #") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				ids = append(ids, parts[len(parts)-1])
			}
		}
	}
	return ids
}

func run() int {
	exitCode, stdout, stderr := runGovulncheck()
	fmt.Print(stdout)
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	if exitCode == 0 {
		return 0
	}

	// Load exceptions.
	exceptions, err := loadExceptions(defaultExceptionsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	// Check if all found vulnerabilities are excepted.
	vulnIDs := parseVulnIDs(stdout)
	var unexcepted []string
	for _, id := range vulnIDs {
		if !exceptions[id] {
			unexcepted = append(unexcepted, id)
		}
	}

	if len(unexcepted) > 0 {
		fmt.Fprintf(os.Stderr, "\nerror: unexcepted vulnerabilities:\n")
		for _, id := range unexcepted {
			fmt.Fprintf(os.Stderr, "  %s\n", id)
		}
		fmt.Fprintf(os.Stderr, "Add to %s to except, or resolve them.\n", defaultExceptionsFile)
		return 1
	}

	fmt.Println("All vulnerabilities are excepted.")
	return 0
}

func main() {
	os.Exit(run())
}

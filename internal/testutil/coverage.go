// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package testutil

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CoverageFloor is the minimum line coverage percentage required. The goal
// is 100%; 98% is the enforced floor.
const CoverageFloor = 98.0

// PackageCoverage holds the coverage result for a single package.
type PackageCoverage struct {
	Package  string
	Coverage float64
}

// ParseCoverageOutput parses the output of `go test -cover ./...` and returns
// per-package coverage. Packages with "[no test files]" are skipped.
func ParseCoverageOutput(r io.Reader) ([]PackageCoverage, error) {
	var results []PackageCoverage
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines that don't start with "ok" or "FAIL".
		if !strings.HasPrefix(line, "ok") && !strings.HasPrefix(line, "FAIL") {
			continue
		}

		// Skip packages with no test files.
		if strings.Contains(line, "[no test files]") {
			continue
		}

		// Parse "ok  <package>  <time>  coverage: XX.X% of statements"
		if strings.Contains(line, "coverage:") {
			pkg, cov, err := parseCoverageLine(line)
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", line, err)
			}
			results = append(results, PackageCoverage{Package: pkg, Coverage: cov})
		} else if strings.HasPrefix(line, "FAIL") {
			// A failing package has no coverage number.
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				results = append(results, PackageCoverage{Package: fields[1], Coverage: 0})
			}
		}
	}

	return results, scanner.Err()
}

// parseCoverageLine extracts package name and coverage from a line like:
// "ok  github.com/foo/bar  0.5s  coverage: 85.0% of statements"
func parseCoverageLine(line string) (string, float64, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", 0, fmt.Errorf("too few fields")
	}
	pkg := fields[1]

	for i, f := range fields {
		if f == "coverage:" && i+1 < len(fields) {
			pctStr := strings.TrimSuffix(fields[i+1], "%")
			cov, err := strconv.ParseFloat(pctStr, 64)
			if err != nil {
				return "", 0, fmt.Errorf("parse coverage %q: %w", fields[i+1], err)
			}
			return pkg, cov, nil
		}
	}

	return "", 0, fmt.Errorf("no coverage field found")
}

// CheckCoverage verifies that every package meets the coverage floor.
// It returns a list of packages below the floor. An empty return means all pass.
func CheckCoverage(results []PackageCoverage) []PackageCoverage {
	var failures []PackageCoverage
	for _, r := range results {
		if r.Coverage < CoverageFloor {
			failures = append(failures, r)
		}
	}
	return failures
}

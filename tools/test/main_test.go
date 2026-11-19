// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAggregateCoverageValid(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	content := `mode: set
github.com/foo/bar/main.go:10.2,12.1 2 1
github.com/foo/bar/main.go:14.2,16.1 3 0
github.com/foo/bar/main.go:18.2,20.1 5 1
`
	if err := os.WriteFile(profile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pct, err := aggregateCoverage(profile)
	if err != nil {
		t.Fatal(err)
	}
	if pct < 69.9 || pct > 70.1 {
		t.Errorf("coverage = %.1f%%, want 70.0%%", pct)
	}
}

func TestAggregateCoverageEmpty(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	if err := os.WriteFile(profile, []byte("mode: set\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pct, err := aggregateCoverage(profile)
	if err != nil {
		t.Fatal(err)
	}
	if pct != 100.0 {
		t.Errorf("coverage = %.1f%%, want 100.0%%", pct)
	}
}

func TestAggregateCoverageMissingFile(t *testing.T) {
	_, err := aggregateCoverage("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestAggregateCoverageMalformedLines(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	content := `mode: set
bad line
github.com/foo/bar/main.go:10.2,12.1 abc 1
github.com/foo/bar/main.go:10.2,12.1 2 xyz
github.com/foo/bar/main.go:10.2,12.1 2 1
`
	if err := os.WriteFile(profile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pct, err := aggregateCoverage(profile)
	if err != nil {
		t.Fatal(err)
	}
	if pct != 100.0 {
		t.Errorf("coverage = %.1f%%, want 100.0%%", pct)
	}
}

func TestAggregateCoverageExcludesToolPackages(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	content := `mode: set
github.com/asymmetric-effort/go-notebook/internal/foo/bar.go:10.2,12.1 10 10
github.com/asymmetric-effort/go-notebook/tools/clean/main.go:5.2,7.1 5 0
`
	if err := os.WriteFile(profile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pct, err := aggregateCoverage(profile)
	if err != nil {
		t.Fatal(err)
	}
	// Only the internal package should count: 10/10 = 100%
	if pct != 100.0 {
		t.Errorf("coverage = %.1f%%, want 100.0%% (tool package should be excluded)", pct)
	}
}

func TestIsExcluded(t *testing.T) {
	if !isExcluded("github.com/asymmetric-effort/go-notebook/tools/clean/main.go") {
		t.Error("expected tools/ to be excluded")
	}
	if isExcluded("github.com/asymmetric-effort/go-notebook/internal/testutil/cli.go") {
		t.Error("expected internal/ to not be excluded")
	}
}

func TestIsExcludedNoMatch(t *testing.T) {
	if isExcluded("github.com/other/repo/main.go") {
		t.Error("expected other repos to not be excluded")
	}
}

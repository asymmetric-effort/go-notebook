// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package testutil

import (
	"strings"
	"testing"
)

func TestParseCoverageOutput(t *testing.T) {
	input := `ok  	github.com/foo/bar	0.5s	coverage: 85.0% of statements
ok  	github.com/foo/baz	0.3s	coverage: 100.0% of statements
?   	github.com/foo/nop	[no test files]
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Package != "github.com/foo/bar" || results[0].Coverage != 85.0 {
		t.Errorf("results[0] = %+v", results[0])
	}
	if results[1].Package != "github.com/foo/baz" || results[1].Coverage != 100.0 {
		t.Errorf("results[1] = %+v", results[1])
	}
}

func TestParseCoverageOutputWithFail(t *testing.T) {
	input := `FAIL	github.com/foo/broken	0.1s
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Coverage != 0 {
		t.Errorf("broken package coverage = %f, want 0", results[0].Coverage)
	}
}

func TestParseCoverageOutputSkipsNonTestLines(t *testing.T) {
	input := `# some build output
ok  	github.com/foo/bar	0.5s	coverage: 90.0% of statements
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
}

func TestParseCoverageOutputEmpty(t *testing.T) {
	results, err := ParseCoverageOutput(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestParseCoverageOutputCachedLine(t *testing.T) {
	// Lines like "ok  pkg  (cached)  coverage: 100.0% of statements"
	input := `ok  	github.com/foo/cached	(cached)	coverage: 99.5% of statements
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Coverage != 99.5 {
		t.Errorf("coverage = %f, want 99.5", results[0].Coverage)
	}
}

func TestParseCoverageOutputOkNoCoverage(t *testing.T) {
	// "ok" line without "coverage:" (e.g., cached without coverage flag)
	input := `ok  	github.com/foo/bar	0.5s
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	// No coverage info means it's skipped.
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestParseCoverageLineShortLine(t *testing.T) {
	_, _, err := parseCoverageLine("ok")
	if err == nil {
		t.Error("expected error for short line")
	}
}

func TestParseCoverageLineNoCoverageField(t *testing.T) {
	_, _, err := parseCoverageLine("ok  github.com/foo/bar  0.5s")
	if err == nil {
		t.Error("expected error for missing coverage field")
	}
}

func TestParseCoverageLineBadPercent(t *testing.T) {
	_, _, err := parseCoverageLine("ok  github.com/foo/bar  0.5s  coverage: abc% of statements")
	if err == nil {
		t.Error("expected error for bad percentage")
	}
}

func TestCheckCoverageAllPass(t *testing.T) {
	results := []PackageCoverage{
		{Package: "a", Coverage: 100.0},
		{Package: "b", Coverage: 98.0},
	}
	failures := CheckCoverage(results)
	if len(failures) != 0 {
		t.Errorf("expected no failures, got %v", failures)
	}
}

func TestCheckCoverageWithFailure(t *testing.T) {
	results := []PackageCoverage{
		{Package: "a", Coverage: 100.0},
		{Package: "b", Coverage: 50.0},
		{Package: "c", Coverage: 97.9},
	}
	failures := CheckCoverage(results)
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures, got %d: %v", len(failures), failures)
	}
}

func TestCoverageFloorValue(t *testing.T) {
	if CoverageFloor != 98.0 {
		t.Errorf("CoverageFloor = %f, want 98.0", CoverageFloor)
	}
}

func TestCheckCoverageEmpty(t *testing.T) {
	failures := CheckCoverage(nil)
	if len(failures) != 0 {
		t.Errorf("expected no failures for nil input, got %v", failures)
	}
}

func TestParseCoverageOutputBadCoverageLine(t *testing.T) {
	input := `ok  	github.com/foo/bar	0.5s	coverage: notanumber% of statements
`
	_, err := ParseCoverageOutput(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for bad coverage line")
	}
}

func TestParseCoverageOutputFailSingleField(t *testing.T) {
	input := `FAIL
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	// FAIL with no package name should be skipped.
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestParseCoverageOutputNoTestFiles(t *testing.T) {
	input := `?   	github.com/foo/nop	[no test files]
`
	results, err := ParseCoverageOutput(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0 (no test files should be skipped)", len(results))
	}
}

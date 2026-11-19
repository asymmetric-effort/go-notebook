// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TB is the subset of testing.TB used by CompareGolden.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// UpdateGolden controls whether golden files are updated. Set the
// GOBOOK_UPDATE_GOLDEN environment variable to "1" to overwrite golden files
// with the actual output instead of comparing.
func UpdateGolden() bool {
	return os.Getenv("GOBOOK_UPDATE_GOLDEN") == "1"
}

// CompareGolden compares actual against the content of goldenPath. If they
// differ, it fails the test with a readable unified diff. If GOBOOK_UPDATE_GOLDEN
// is set, it overwrites the golden file instead.
func CompareGolden(t TB, goldenPath string, actual string) {
	t.Helper()

	if UpdateGolden() {
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("update golden file %s: %v", goldenPath, err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v\nRun with GOBOOK_UPDATE_GOLDEN=1 to create it.", goldenPath, err)
	}

	if string(expected) == actual {
		return
	}

	diff := unifiedDiff(string(expected), actual)
	t.Errorf("output does not match golden file %s:\n%s", goldenPath, diff)
}

// unifiedDiff produces a simple line-by-line diff between expected and actual.
func unifiedDiff(expected, actual string) string {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")

	var b strings.Builder
	b.WriteString("--- expected\n+++ actual\n")

	maxLen := len(expLines)
	if len(actLines) > maxLen {
		maxLen = len(actLines)
	}

	for i := 0; i < maxLen; i++ {
		var exp, act string
		if i < len(expLines) {
			exp = expLines[i]
		}
		if i < len(actLines) {
			act = actLines[i]
		}
		if exp != act {
			if i < len(expLines) {
				b.WriteString(fmt.Sprintf("-%s\n", exp))
			}
			if i < len(actLines) {
				b.WriteString(fmt.Sprintf("+%s\n", act))
			}
		} else {
			b.WriteString(fmt.Sprintf(" %s\n", exp))
		}
	}

	return b.String()
}

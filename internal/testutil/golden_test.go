// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareGoldenMatch(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "test.golden")
	if err := os.WriteFile(golden, []byte("expected output"), 0o644); err != nil {
		t.Fatal(err)
	}
	CompareGolden(t, golden, "expected output")
}

func TestCompareGoldenMismatch(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "test.golden")
	if err := os.WriteFile(golden, []byte("expected"), 0o644); err != nil {
		t.Fatal(err)
	}

	// CompareGolden calls t.Errorf on mismatch. Use a mock T to verify.
	var called bool
	mockT := &mockTHelper{errorf: func(format string, args ...any) { called = true }}
	CompareGolden(mockT, golden, "actual")
	if !called {
		t.Error("expected Errorf to be called on mismatch")
	}
}

func TestCompareGoldenMissingFile(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "nonexistent.golden")

	var called bool
	mockT := &mockTHelper{fatalf: func(format string, args ...any) {
		called = true
		panic("fatalf") // Simulate t.Fatalf behavior.
	}}

	func() {
		defer func() { recover() }()
		CompareGolden(mockT, golden, "anything")
	}()

	if !called {
		t.Error("expected Fatalf to be called for missing golden file")
	}
}

func TestCompareGoldenUpdate(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "sub", "test.golden")

	t.Setenv("GOBOOK_UPDATE_GOLDEN", "1")
	CompareGolden(t, golden, "new content")

	data, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new content" {
		t.Errorf("golden file = %q, want %q", string(data), "new content")
	}
}

func TestUpdateGoldenDefault(t *testing.T) {
	t.Setenv("GOBOOK_UPDATE_GOLDEN", "")
	if UpdateGolden() {
		t.Error("UpdateGolden() = true, want false")
	}
}

func TestUpdateGoldenEnabled(t *testing.T) {
	t.Setenv("GOBOOK_UPDATE_GOLDEN", "1")
	if !UpdateGolden() {
		t.Error("UpdateGolden() = false, want true")
	}
}

func TestUnifiedDiffIdentical(t *testing.T) {
	diff := unifiedDiff("same", "same")
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			t.Errorf("unexpected removal in identical diff: %q", line)
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			t.Errorf("unexpected addition in identical diff: %q", line)
		}
	}
}

func TestUnifiedDiffExpectedLonger(t *testing.T) {
	diff := unifiedDiff("line1\nline2\nline3", "line1")
	if !strings.Contains(diff, "-line2") {
		t.Errorf("diff missing removed lines:\n%s", diff)
	}
}

func TestUnifiedDiffActualLonger(t *testing.T) {
	diff := unifiedDiff("line1", "line1\nline2\nline3")
	if !strings.Contains(diff, "+line2") {
		t.Errorf("diff missing added lines:\n%s", diff)
	}
}

func TestCompareGoldenUpdateMkdirFails(t *testing.T) {
	dir := t.TempDir()
	// Create a file where a directory is needed.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join(blocker, "sub", "test.golden")

	t.Setenv("GOBOOK_UPDATE_GOLDEN", "1")
	var called bool
	mockT := &mockTHelper{fatalf: func(format string, args ...any) {
		called = true
		panic("fatalf")
	}}
	func() {
		defer func() { recover() }()
		CompareGolden(mockT, golden, "content")
	}()
	if !called {
		t.Error("expected Fatalf for MkdirAll failure")
	}
}

func TestCompareGoldenUpdateWriteFails(t *testing.T) {
	dir := t.TempDir()
	// Create a read-only directory.
	roDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(roDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(roDir, 0o755) })
	golden := filepath.Join(roDir, "test.golden")

	t.Setenv("GOBOOK_UPDATE_GOLDEN", "1")
	var called bool
	mockT := &mockTHelper{fatalf: func(format string, args ...any) {
		called = true
		panic("fatalf")
	}}
	func() {
		defer func() { recover() }()
		CompareGolden(mockT, golden, "content")
	}()
	if !called {
		t.Error("expected Fatalf for write failure")
	}
}

func TestUnifiedDiffReadable(t *testing.T) {
	diff := unifiedDiff("expected", "actual")
	if !strings.Contains(diff, "-expected") || !strings.Contains(diff, "+actual") {
		t.Errorf("diff not readable:\n%s", diff)
	}
}

// mockTHelper implements the subset of testing.TB that CompareGolden uses.
type mockTHelper struct {
	errorf func(string, ...any)
	fatalf func(string, ...any)
}

func (m *mockTHelper) Helper() {}
func (m *mockTHelper) Errorf(format string, args ...any) {
	if m.errorf != nil {
		m.errorf(format, args...)
	}
}
func (m *mockTHelper) Fatalf(format string, args ...any) {
	if m.fatalf != nil {
		m.fatalf(format, args...)
	}
}

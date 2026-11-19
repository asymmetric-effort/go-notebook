// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestWriteOutputBlocksInsert(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"test": "42"})
	if !strings.Contains(result, "```output name=test") {
		t.Errorf("missing output block:\n%s", result)
	}
	if !strings.Contains(result, "42") {
		t.Errorf("missing output content:\n%s", result)
	}
}

func TestWriteOutputBlocksReplace(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n\n```output name=test\nold value\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"test": "new value"})
	if strings.Contains(result, "old value") {
		t.Errorf("old output should be replaced:\n%s", result)
	}
	if !strings.Contains(result, "new value") {
		t.Errorf("new output missing:\n%s", result)
	}
}

func TestWriteOutputBlocksNoOutputs(t *testing.T) {
	src := "```go:notebook name=test\nval = 42\n```\n"
	result := WriteOutputBlocks(src, nil)
	if result != src {
		t.Error("no outputs should leave source unchanged")
	}
}

func TestWriteOutputBlocksUnnamedCell(t *testing.T) {
	src := "```go:notebook\nfmt.Println(\"hi\")\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"test": "42"})
	if strings.Contains(result, "output") {
		t.Error("unnamed cell should not get output block")
	}
}

func TestWriteOutputBlocksMultiple(t *testing.T) {
	src := "```go:notebook name=a exported(x int)\nx = 1\n```\n\n```go:notebook name=b exported(y int)\ny = 2\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"a": "1", "b": "2"})
	if !strings.Contains(result, "```output name=a") {
		t.Error("missing output for a")
	}
	if !strings.Contains(result, "```output name=b") {
		t.Error("missing output for b")
	}
}

func TestWriteOutputBlocksInvalidInfo(t *testing.T) {
	// Cell with invalid info string (name=) should be skipped.
	src := "```go:notebook name=\nbody\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"": "data"})
	_ = result // Should not panic.
}

func TestFindExistingOutputProse(t *testing.T) {
	// Prose between cell and next block should stop search.
	src := "```go:notebook name=a\nbody\n```\n\nSome prose here.\n\n```go:notebook name=b\nbody\n```\n"
	result := Parse(src)
	start, end := findExistingOutput(result.Blocks, 0, "a")
	if start != 0 || end != 0 {
		t.Errorf("expected no output found, got %d-%d", start, end)
	}
}

func TestFindExistingOutputWrongName(t *testing.T) {
	src := "```go:notebook name=a\nbody\n```\n\n```output name=b\ndata\n```\n"
	result := Parse(src)
	start, end := findExistingOutput(result.Blocks, 0, "a")
	if start != 0 || end != 0 {
		t.Errorf("expected no output for wrong name, got %d-%d", start, end)
	}
}

func TestFindExistingOutputNonOutput(t *testing.T) {
	// When next block is a cell, not an output.
	src := "```go:notebook name=a\n```\n\n```go:notebook name=b\n```\n"
	result := Parse(src)
	start, end := findExistingOutput(result.Blocks, 0, "a")
	if start != 0 || end != 0 {
		t.Errorf("expected no existing output, got %d-%d", start, end)
	}
}

func TestFormatOutputBlock(t *testing.T) {
	lines := formatOutputBlock("test", "hello world")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "```output name=test" {
		t.Errorf("line[0] = %q", lines[0])
	}
	if lines[1] != "hello world" {
		t.Errorf("line[1] = %q", lines[1])
	}
	if lines[2] != "```" {
		t.Errorf("line[2] = %q", lines[2])
	}
}

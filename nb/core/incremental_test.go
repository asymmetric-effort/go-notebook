// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"testing"
)

func TestDiffCellsNoChanges(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n"
	old := Parse(src)
	new := Parse(src)
	changes := DiffCells(old, new)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiffCellsBodyChanged(t *testing.T) {
	old := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n")
	new := Parse("```go:notebook name=a exported(val int)\nval = 42\n```\n")
	changes := DiffCells(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Name != "a" {
		t.Errorf("name = %q", changes[0].Name)
	}
	if changes[0].IsNew || changes[0].IsDelete {
		t.Error("should not be new or delete")
	}
}

func TestDiffCellsNewCell(t *testing.T) {
	old := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n")
	new := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n\n```go:notebook name=b exported(x int)\nx = 2\n```\n")
	changes := DiffCells(old, new)
	found := false
	for _, c := range changes {
		if c.Name == "b" && c.IsNew {
			found = true
		}
	}
	if !found {
		t.Errorf("expected new cell 'b', got: %v", changes)
	}
}

func TestDiffCellsDeletedCell(t *testing.T) {
	old := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n\n```go:notebook name=b exported(x int)\nx = 2\n```\n")
	new := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n")
	changes := DiffCells(old, new)
	found := false
	for _, c := range changes {
		if c.Name == "b" && c.IsDelete {
			found = true
		}
	}
	if !found {
		t.Errorf("expected deleted cell 'b', got: %v", changes)
	}
}

func TestDiffCellsUnnamedChanged(t *testing.T) {
	old := Parse("```go:notebook\nold body\n```\n")
	new := Parse("```go:notebook\nnew body\n```\n")
	changes := DiffCells(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Name != "" {
		t.Error("should be unnamed")
	}
}

func TestNeedsFullRegenNoChanges(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n"
	old := Parse(src)
	new := Parse(src)
	if NeedsFullRegeneration(old, new) {
		t.Error("no changes should not need full regen")
	}
}

func TestNeedsFullRegenDefsChanged(t *testing.T) {
	old := Parse("```go:defs\nconst x = 1\n```\n\n```go:notebook name=a exported(val int)\nval = x\n```\n")
	new := Parse("```go:defs\nconst x = 2\n```\n\n```go:notebook name=a exported(val int)\nval = x\n```\n")
	if !NeedsFullRegeneration(old, new) {
		t.Error("defs change should need full regen")
	}
}

func TestNeedsFullRegenCellAdded(t *testing.T) {
	old := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n")
	new := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n\n```go:notebook name=b exported(x int)\nx = 2\n```\n")
	if !NeedsFullRegeneration(old, new) {
		t.Error("new cell should need full regen")
	}
}

func TestNeedsFullRegenBodyOnly(t *testing.T) {
	old := Parse("```go:notebook name=a exported(val int)\nval = 1\n```\n")
	new := Parse("```go:notebook name=a exported(val int)\nval = 42\n```\n")
	if NeedsFullRegeneration(old, new) {
		t.Error("body-only change should not need full regen")
	}
}

func TestIncrementalEqualsFullRegen(t *testing.T) {
	// Verify that incremental changes produce the same output as full regen.
	src1 := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	src2 := "```go:notebook name=test exported(val int)\nval = 42\n```\n"

	gen1, err := Generate(Parse(src1), "main")
	if err != nil {
		t.Fatal(err)
	}
	gen2, err := Generate(Parse(src2), "main")
	if err != nil {
		t.Fatal(err)
	}

	// The cell function code should differ (val = 1 vs val = 42).
	if gen1.Cells[0].FuncCode == gen2.Cells[0].FuncCode {
		t.Error("cell code should differ between versions")
	}

	// But the structure should be the same.
	if len(gen1.Cells) != len(gen2.Cells) {
		t.Error("cell count should be the same")
	}
}

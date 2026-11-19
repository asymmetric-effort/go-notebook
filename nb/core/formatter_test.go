// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestFormatCellSimple(t *testing.T) {
	body := "x:=1\ny:=2"
	formatted, err := FormatCell(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted, "x := 1") {
		t.Errorf("formatted = %q, want spaces around :=", formatted)
	}
}

func TestFormatCellPreservesValid(t *testing.T) {
	body := "x := 1\ny := 2"
	formatted, err := FormatCell(body)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != body {
		t.Errorf("formatted = %q, want %q (already formatted)", formatted, body)
	}
}

func TestFormatCellInvalid(t *testing.T) {
	_, err := FormatCell("func {{{")
	if err == nil {
		t.Error("expected error for invalid Go")
	}
}

func TestFormatDefsSimple(t *testing.T) {
	body := "import(\n\"fmt\"\n)\n\nconst x=1"
	formatted, err := FormatDefs(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted, "const x = 1") {
		t.Errorf("formatted = %q", formatted)
	}
}

func TestFormatDefsInvalid(t *testing.T) {
	_, err := FormatDefs("var {{{")
	if err == nil {
		t.Error("expected error for invalid Go")
	}
}

func TestFormatNotebook(t *testing.T) {
	src := "# Title\n\n```go:defs\nimport(\n\"fmt\"\n)\n```\n\n```go:notebook name=test\nx:=1\n```\n"
	formatted, err := FormatNotebook(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted, "x := 1") {
		t.Errorf("cell body not formatted:\n%s", formatted)
	}
}

func TestFormatNotebookProse(t *testing.T) {
	src := "# Just prose\n\nNo code blocks.\n"
	formatted, err := FormatNotebook(src)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != src {
		t.Errorf("prose should be unchanged:\n%s", formatted)
	}
}

func TestWrapUnwrapRoundTrip(t *testing.T) {
	body := "x := 1\nfmt.Println(x)"
	wrapped := wrapBody(body)
	unwrapped, err := unwrapBody(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if unwrapped != body {
		t.Errorf("round trip failed:\ngot:  %q\nwant: %q", unwrapped, body)
	}
}

func TestUnwrapBodyInvalid(t *testing.T) {
	_, err := unwrapBody("no function here")
	if err == nil {
		t.Error("expected error for invalid wrapped output")
	}
}

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestGenerateNamedCell(t *testing.T) {
	src := "```go:notebook name=load exported(rows []Row)\nf, err := ctx.Open(\"sales.csv\")\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(gen.Cells))
	}
	cell := gen.Cells[0]
	if cell.StructName != "loadCell" {
		t.Errorf("StructName = %q, want %q", cell.StructName, "loadCell")
	}
	if cell.FuncName != "runLoad" {
		t.Errorf("FuncName = %q, want %q", cell.FuncName, "runLoad")
	}
	if !strings.Contains(cell.StructCode, "err error") {
		t.Error("struct missing err field")
	}
	if !strings.Contains(cell.StructCode, "rows []Row") {
		t.Error("struct missing exported field")
	}
	if !strings.Contains(cell.FuncCode, "defer func()") {
		t.Error("function missing deferred recovery")
	}
	if !strings.Contains(cell.FuncCode, "nb.PanicError") {
		t.Error("function missing panic recovery")
	}
	if !strings.Contains(cell.FuncCode, "out = loadCell{") {
		t.Error("function missing result assignment")
	}
}

func TestGenerateUnnamedCell(t *testing.T) {
	src := "```go:notebook\nfmt.Println(\"hello\")\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(gen.Cells))
	}
	cell := gen.Cells[0]
	if cell.StructName != "" {
		t.Errorf("StructName = %q, want empty", cell.StructName)
	}
	if cell.FuncName != "runCell1" {
		t.Errorf("FuncName = %q, want %q", cell.FuncName, "runCell1")
	}
	if !strings.Contains(cell.FuncCode, "defer func()") {
		t.Error("function missing deferred recovery")
	}
}

func TestGenerateMultipleCells(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n```go:notebook name=b exported(result int)\nresult = a.val + 1\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Cells) != 2 {
		t.Fatalf("got %d cells, want 2", len(gen.Cells))
	}
	// Second cell should have first cell as parameter.
	if !strings.Contains(gen.Cells[1].FuncCode, "a aCell") {
		t.Error("second cell should have first cell as parameter")
	}
}

func TestGenerateDefsCode(t *testing.T) {
	src := "```go:defs\nimport \"fmt\"\n\nconst x = 1\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gen.DefsCode, "import \"fmt\"") {
		t.Error("missing import in defs code")
	}
}

func TestGenerateLineDirective(t *testing.T) {
	src := "# Title\n\n```go:notebook name=test\nbody line\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gen.Cells[0].FuncCode, "//line notebook.md:4:1") {
		t.Errorf("missing or wrong line directive in:\n%s", gen.Cells[0].FuncCode)
	}
}

func TestGenerateMultipleExported(t *testing.T) {
	src := "```go:notebook name=test exported(a int, b string, c []float64)\na = 1\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	sc := gen.Cells[0].StructCode
	if !strings.Contains(sc, "a int") || !strings.Contains(sc, "b string") || !strings.Contains(sc, "c []float64") {
		t.Errorf("struct missing fields:\n%s", sc)
	}
}

func TestGenerateEmptyNotebook(t *testing.T) {
	result := Parse("# Just prose\n")
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Cells) != 0 {
		t.Errorf("got %d cells, want 0", len(gen.Cells))
	}
}

func TestParseExportedFields(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"rows []Row, total float64", 2},
		{"val int", 1},
		{"m map[string][]int, f func(int, int) error", 2},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseExportedFields(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseExportedFields(%q) = %d fields, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"load", "Load"},
		{"a", "A"},
		{"", ""},
		{"Load", "Load"},
	}
	for _, tt := range tests {
		if got := capitalize(tt.input); got != tt.want {
			t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIndentBody(t *testing.T) {
	body := "a := 1\nb := 2"
	got := indentBody(body)
	if !strings.Contains(got, "\ta := 1") {
		t.Errorf("indentBody not indenting:\n%s", got)
	}
}

func TestGenerateInvalidInfoString(t *testing.T) {
	src := "```go:notebook name=\nbody\n```\n"
	result := Parse(src)
	_, err := Generate(result, "main")
	if err == nil {
		t.Error("expected error for invalid info string")
	}
}

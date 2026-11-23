// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

func TestParseEmbedOutput(t *testing.T) {
	stdout := "some output\nGOBOOK_EMBED:load:aGVsbG8=\nmore output\nGOBOOK_EMBED:totals:d29ybGQ=\n"
	embedded := parseEmbedOutput(stdout)
	if len(embedded) != 2 {
		t.Fatalf("got %d embedded, want 2", len(embedded))
	}
	if embedded[0].CellName != "load" {
		t.Errorf("cell[0].Name = %q", embedded[0].CellName)
	}
	if string(embedded[0].EncryptedData) != "hello" {
		t.Errorf("cell[0].Data = %q", embedded[0].EncryptedData)
	}
	if embedded[1].CellName != "totals" {
		t.Errorf("cell[1].Name = %q", embedded[1].CellName)
	}
}

func TestParseEmbedOutputEmpty(t *testing.T) {
	embedded := parseEmbedOutput("no embed lines here\n")
	if len(embedded) != 0 {
		t.Errorf("got %d embedded, want 0", len(embedded))
	}
}

func TestParseEmbedOutputBadBase64(t *testing.T) {
	embedded := parseEmbedOutput("GOBOOK_EMBED:bad:!!!invalid!!!\n")
	if len(embedded) != 0 {
		t.Errorf("bad base64 should be skipped, got %d", len(embedded))
	}
}

func TestHasEmbeddedCells(t *testing.T) {
	gen := core.GenerateResult{
		Cells: []core.GeneratedCell{
			{StructName: "loadCell", DataMode: core.DataEmbedded},
			{StructName: "displayCell", DataMode: core.DataNone},
		},
	}
	if !hasEmbeddedCells(gen) {
		t.Error("should detect embedded cells")
	}
}

func TestHasNoEmbeddedCells(t *testing.T) {
	gen := core.GenerateResult{
		Cells: []core.GeneratedCell{
			{StructName: "testCell", DataMode: core.DataNone},
		},
	}
	if hasEmbeddedCells(gen) {
		t.Error("should not detect embedded cells")
	}
}

func TestBuildWithLiveCell(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int) data=live\nval = 42\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.html")
	err := Run(notebook, Options{Format: "html", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Error("should produce valid HTML")
	}
}

func TestBuildWithEmbeddedCell(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int) data=embedded\nval = 42\n```\n"
	if err := os.WriteFile(notebook, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out.html")
	err := Run(notebook, Options{Format: "html", Output: outPath})
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Error("should produce valid HTML")
	}
}

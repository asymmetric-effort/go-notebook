// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParityAllNotebooks runs every example notebook through md, html, and
// typst formats and verifies each produces valid output with consistent
// content. This is the parity guarantee: the same notebook renders
// equivalently across all backends.
func TestParityAllNotebooks(t *testing.T) {
	notebooks, err := filepath.Glob("testdata/notebooks/*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(notebooks) == 0 {
		t.Fatal("no test notebooks found in testdata/notebooks/")
	}

	for _, nb := range notebooks {
		name := strings.TrimSuffix(filepath.Base(nb), ".md")
		t.Run(name, func(t *testing.T) {
			testParityNotebook(t, nb, name)
		})
	}
}

func testParityNotebook(t *testing.T, notebookPath, name string) {
	t.Helper()
	dir := t.TempDir()

	// Copy the notebook to the temp dir so outputs land there.
	src, err := os.ReadFile(notebookPath)
	if err != nil {
		t.Fatal(err)
	}
	nbCopy := filepath.Join(dir, name+".md")
	os.WriteFile(nbCopy, src, 0o644)

	isPanic := name == "panic_recovery"

	// --- Markdown output ---
	mdOut := filepath.Join(dir, name+".out.md")
	mdErr := Run(nbCopy, Options{Format: "md", Output: mdOut})
	if !isPanic && mdErr != nil {
		t.Fatalf("md: %v", mdErr)
	}
	mdData, err := os.ReadFile(mdOut)
	if err != nil {
		t.Fatalf("md output not written: %v", err)
	}
	if len(mdData) == 0 {
		t.Error("md output is empty")
	}

	// --- HTML output ---
	htmlOut := filepath.Join(dir, name+".out.html")
	htmlErr := Run(nbCopy, Options{Format: "html", Output: htmlOut})
	if !isPanic && htmlErr != nil {
		t.Fatalf("html: %v", htmlErr)
	}
	htmlData, err := os.ReadFile(htmlOut)
	if err != nil {
		t.Fatalf("html output not written: %v", err)
	}
	if len(htmlData) == 0 {
		t.Error("html output is empty")
	}

	// --- Typst output ---
	typstOut := filepath.Join(dir, name+".out.typ")
	typstErr := Run(nbCopy, Options{Format: "typst", Output: typstOut})
	if !isPanic && typstErr != nil {
		t.Fatalf("typst: %v", typstErr)
	}
	typstData, err := os.ReadFile(typstOut)
	if err != nil {
		t.Fatalf("typst output not written: %v", err)
	}
	if len(typstData) == 0 {
		t.Error("typst output is empty")
	}

	// --- Parity checks ---
	// All formats must include the first heading.
	mdStr := string(mdData)
	htmlStr := string(htmlData)
	typstStr := string(typstData)

	// Extract first heading from source.
	for _, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(line, "# ") {
			heading := strings.TrimPrefix(line, "# ")
			if !strings.Contains(mdStr, heading) {
				t.Errorf("md missing heading %q", heading)
			}
			if !strings.Contains(htmlStr, heading) {
				t.Errorf("html missing heading %q", heading)
			}
			// Typst converts # Heading to = Heading
			if !strings.Contains(typstStr, heading) {
				t.Errorf("typst missing heading %q", heading)
			}
			break
		}
	}

	// HTML must have CSP and DOCTYPE.
	if !strings.Contains(htmlStr, "Content-Security-Policy") {
		t.Error("html missing CSP")
	}
	if !strings.Contains(htmlStr, "<!DOCTYPE html>") {
		t.Error("html missing DOCTYPE")
	}

	// Typst must have page setup.
	if !strings.Contains(typstStr, "#set page") {
		t.Error("typst missing page setup")
	}

	// Panic notebook: all formats should still produce output (partial report).
	if isPanic {
		if len(mdData) < 10 {
			t.Error("panic md output too small for partial report")
		}
		if len(htmlData) < 50 {
			t.Error("panic html output too small for partial report")
		}
		if len(typstData) < 10 {
			t.Error("panic typst output too small for partial report")
		}
	}

	t.Logf("parity OK: md=%d bytes, html=%d bytes, typst=%d bytes", len(mdData), len(htmlData), len(typstData))
}

// TestParityDisplayOutput verifies that cell output appears consistently
// across formats when a cell prints to stdout.
func TestParityDisplayOutput(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "display.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(msg string)\nmsg = \"parity-check-output\"\nfmt.Println(msg)\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	mdOut := filepath.Join(dir, "out.md")
	htmlOut := filepath.Join(dir, "out.html")

	if err := Run(notebook, Options{Format: "md", Output: mdOut}); err != nil {
		t.Fatal(err)
	}
	if err := Run(notebook, Options{Format: "html", Output: htmlOut}); err != nil {
		t.Fatal(err)
	}

	mdData, _ := os.ReadFile(mdOut)
	htmlData, _ := os.ReadFile(htmlOut)

	// Both formats should include the cell output.
	if !strings.Contains(string(mdData), "parity-check-output") {
		t.Error("md missing cell output")
	}
	if !strings.Contains(string(htmlData), "parity-check-output") {
		t.Error("html missing cell output")
	}
}

// TestParityMultiCellUpstream verifies upstream value passing produces
// consistent results across formats.
func TestParityMultiCellUpstream(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "multi.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n" +
		"```go:notebook name=base exported(val int)\nval = 7\n```\n\n" +
		"```go:notebook name=derived exported(result int)\nresult = base.val * 6\nfmt.Println(result)\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	mdOut := filepath.Join(dir, "out.md")
	htmlOut := filepath.Join(dir, "out.html")

	if err := Run(notebook, Options{Format: "md", Output: mdOut}); err != nil {
		t.Fatal(err)
	}
	if err := Run(notebook, Options{Format: "html", Output: htmlOut}); err != nil {
		t.Fatal(err)
	}

	mdData, _ := os.ReadFile(mdOut)
	htmlData, _ := os.ReadFile(htmlOut)

	// 7 * 6 = 42 should appear in both.
	if !strings.Contains(string(mdData), "42") {
		t.Error("md missing computed result 42")
	}
	if !strings.Contains(string(htmlData), "42") {
		t.Error("html missing computed result 42")
	}
}

// TestParityIdempotent verifies that running the same notebook twice
// produces identical output.
func TestParityIdempotent(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Idempotent\n\n```go:notebook name=test exported(val int)\nval = 99\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	out1 := filepath.Join(dir, "out1.html")
	out2 := filepath.Join(dir, "out2.html")

	if err := Run(notebook, Options{Format: "html", Output: out1}); err != nil {
		t.Fatal(err)
	}
	if err := Run(notebook, Options{Format: "html", Output: out2}); err != nil {
		t.Fatal(err)
	}

	data1, _ := os.ReadFile(out1)
	data2, _ := os.ReadFile(out2)

	if string(data1) != string(data2) {
		t.Error("two runs of the same notebook produced different HTML output")
	}
}

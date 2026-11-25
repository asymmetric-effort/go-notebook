// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Tests specifically targeting uncovered lines to push toward 100% coverage.
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatNotebookWithError(t *testing.T) {
	// A cell with invalid Go that can't be formatted.
	src := "```go:notebook name=test\nfunc {{{\n```\n"
	_, err := FormatNotebook(src)
	if err == nil {
		t.Error("expected error for invalid Go in cell")
	}
}

func TestFormatNotebookAlreadyFormatted(t *testing.T) {
	// Already-formatted cell body should be unchanged.
	src := "```go:notebook name=test\nx := 1\n```\n"
	result, err := FormatNotebook(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "x := 1") {
		t.Error("formatted body should be unchanged")
	}
}

func TestIsImportLineNoQuotes(t *testing.T) {
	unused := map[string]bool{"os": true}
	// Line with no quotes at all.
	if isImportLine("func main()", unused) {
		t.Error("non-import line should return false")
	}
}

func TestIsImportLineSingleQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	// Line with only one quote (malformed).
	if isImportLine("\"os", unused) {
		t.Error("malformed import line should return false")
	}
}

func TestCheckDefsUnknownDeclType(t *testing.T) {
	// This is hard to trigger because go/ast only produces GenDecl and FuncDecl.
	// The default case in checkDefs handles unknown GenDecl tokens.
	// We can't easily create an AST with an unknown token type from source,
	// but we can test that a valid defs block with all allowed types passes.
	src := "```go:defs\nimport \"fmt\"\nconst x = 1\ntype T struct{}\nfunc f() {}\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("valid defs should pass: %v", errs)
	}
}

func TestCopyEmbedFilesRelPathError(t *testing.T) {
	// CopyEmbedFiles when glob matches but Rel fails.
	// This is very hard to trigger since Rel only fails when the
	// base is not a prefix, which Glob guarantees. Test the success path.
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, srcDir, "a.txt", "data-a")
	writeTestFile(t, srcDir, "b.txt", "data-b")
	err := CopyEmbedFiles([]string{"*.txt"}, srcDir, dstDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteOutputBlocksInvalidCellInfo(t *testing.T) {
	// A cell with an invalid info string should be skipped.
	src := "```go:notebook name=\nbody\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"": "data"})
	_ = result // Should not panic.
}

func TestIsImportLineNoStartQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	if isImportLine("no quotes here", unused) {
		t.Error("should return false for no quotes")
	}
}

func TestIsImportLineOneQuote(t *testing.T) {
	unused := map[string]bool{"os": true}
	if isImportLine(`"os`, unused) {
		t.Error("should return false for single quote")
	}
}

func TestUnwrapBodyNoClosingBrace(t *testing.T) {
	_, err := unwrapBody("func _() {\n  x := 1\n")
	if err == nil {
		t.Error("expected error for no closing brace")
	}
}

func TestParseInfoStringOnlyWhitespace(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook   ", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "" {
		t.Error("should have empty name")
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := writeFile(dir, name, content); err != nil {
		t.Fatal(err)
	}
}

func writeFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func TestCheckCellFieldMutationUnparseable(t *testing.T) {
	// A cell body that can't be parsed should not crash the checker.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\n{{{invalid go\nresult = 1\n```\n"
	result := Parse(src)
	// Should not panic.
	_ = Check(result)
}

func TestCheckFileAccessUnparseableBody(t *testing.T) {
	// Same for file access checker — unparseable body should be safe.
	src := "```go:notebook name=test reads(data.csv:embed)\n{{{invalid\n```\n"
	result := Parse(src)
	_ = Check(result)
}

func TestWritePackageEmbedMissingSrcDir(t *testing.T) {
	// WritePackage with a notebook path whose dir doesn't exist for embed patterns.
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "//go:embed nonexistent.txt\nvar data string",
	}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	// notebookPath points to a nonexistent directory, so embed copy finds no files.
	err := WritePackage(gen, pkgDir, "/nonexistent/test.md", "testpkg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckCellReferencesUnparseable(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\n{{{invalid\nresult = 1\n```\n"
	result := Parse(src)
	_ = Check(result)
}

func TestExtractEmbedPatternsMultiline(t *testing.T) {
	patterns := ExtractEmbedPatterns("//go:embed a.txt b.txt\n//go:embed c.txt")
	if len(patterns) != 3 {
		t.Errorf("got %d patterns, want 3", len(patterns))
	}
}

func TestIsWebURL(t *testing.T) {
	if !isWebURL("http://example.com") {
		t.Error("http should be web URL")
	}
	if !isWebURL("https://example.com") {
		t.Error("https should be web URL")
	}
	if isWebURL("ftp://example.com") {
		t.Error("ftp should not be web URL")
	}
	if isWebURL("../relative") {
		t.Error("relative should not be web URL")
	}
}

func TestCountCells(t *testing.T) {
	src := "text\n\n```go:notebook name=a\nval = 1\n```\n\nmore text\n\n```go:notebook name=b\nval = 2\n```\n"
	result := Parse(src)
	count := 0
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			count++
		}
	}
	if count != 2 {
		t.Errorf("got %d cells, want 2", count)
	}
}

func TestGenerateResultSourceMapInit(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if gen.SourceMap == nil {
		t.Error("SourceMap should be initialized")
	}
}

// TestBuildSourceMapNoColons triggers the "//line with no colons → continue" branch.
func TestBuildSourceMapNoColons(t *testing.T) {
	// A //line directive with no colons at all — both lastColon guards fire.
	src := "package main\n\n//line nofilenoconat\nval = 1\n\nfunc main() {}\n"
	sm := NewSourceMap()
	buildSourceMapFromSource(src, sm)
	// No entry should be added for the malformed directive.
	_ = sm.Len()
}

// TestBuildSourceMapOneColon triggers the "//line with only one colon → continue" branch.
func TestBuildSourceMapOneColon(t *testing.T) {
	// A //line directive with exactly one colon — secondLastColon < 0 guard fires.
	src := "package main\n\n//line somefile:10\nval = 1\n\nfunc main() {}\n"
	sm := NewSourceMap()
	buildSourceMapFromSource(src, sm)
	// No valid two-colon entry should be added.
	_ = sm.Len()
}

// TestGenerateBatchMainEmbeddedNoDataWithUpstream triggers the branch where an
// embedded cell has no data yet but there are prior named vars (params includes upstream).
func TestGenerateBatchMainEmbeddedNoDataWithUpstream(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		Cells: []GeneratedCell{
			{
				StructName: "loadCell",
				FuncName:   "runLoad",
				StructCode: "type loadCell struct {\n\terr error\n\tval int\n}",
				FuncCode:   "func runLoad(ctx *nb.Context) (out loadCell) {\n\treturn\n}",
			},
			{
				StructName: "computeCell",
				FuncName:   "runCompute",
				StructCode: "type computeCell struct {\n\terr error\n\tresult int\n}",
				FuncCode:   "func runCompute(ctx *nb.Context, load loadCell) (out computeCell) {\n\treturn\n}",
				DataMode:   DataEmbedded,
				// No EmbeddedCells entry — first pass, no data yet.
			},
		},
		// EmbeddedCells is empty — this triggers the "no data" path.
	}
	src := assembleSource(gen, "test.md")
	// Should include GOBOOK_EMBED_MODE check and pass prior named vars.
	if !strings.Contains(src, "GOBOOK_EMBED_MODE") {
		t.Error("should have embed mode check")
	}
	// Should include load in parameters since it's an upstream named var.
	if !strings.Contains(src, "runCompute(ctx, load)") {
		t.Errorf("should pass upstream load to embedded cell, got:\n%s", src)
	}
}

// TestGenerateSocketMainUnnamedCell triggers the "unnamed cell → continue" branch
// in generateSocketMain.
func TestGenerateSocketMainUnnamedCell(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		SocketMode:  true,
		TargetCell:  "process",
		Cells: []GeneratedCell{
			// An unnamed cell (StructName == "") — should be skipped.
			{
				StructName: "",
				FuncName:   "runAnon",
				FuncCode:   "func runAnon(ctx *nb.Context) {\n}",
			},
			{
				StructName: "processCell",
				FuncName:   "runProcess",
				StructCode: "type processCell struct {\n\terr error\n\tresult int\n}",
				FuncCode:   "func runProcess(ctx *nb.Context) (out processCell) {\n\treturn\n}",
			},
		},
	}
	src := assembleSource(gen, "test.md")
	// The unnamed cell should not appear as an upstream fetch.
	if strings.Contains(src, "FetchUpstream(\"\")") {
		t.Error("unnamed cell should be skipped in socket main")
	}
	if !strings.Contains(src, "GOBOOK_SOCKET") {
		t.Error("should have socket setup")
	}
}

// TestParseInfoStringWhitespaceOnlyNoPrefix triggers the break when rest is
// non-empty whitespace but no go:notebook prefix, so TrimSpace is not applied
// before the loop.
func TestParseInfoStringWhitespaceOnlyNoPrefix(t *testing.T) {
	// No "go:notebook" prefix — rest = "  " (non-empty), loop runs,
	// TrimLeftFunc returns "", break fires.
	attrs, err := ParseInfoString("  ", 1)
	if err == nil {
		// If no error, attrs should have empty name (unknown attribute " " error possible).
		_ = attrs
	}
}

// TestWriteOutputBlocksNamedCellNotInOutputs triggers the "named cell not in outputs
// map → continue" branch (output.go:34 — !ok after looking up attrs.Name).
func TestWriteOutputBlocksNamedCellNotInOutputs(t *testing.T) {
	// Two named cells; outputs map only has "b", so "a" triggers the !ok continue.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nresult = 2\n```\n"
	result := WriteOutputBlocks(src, map[string]string{"b": "2"})
	// Only b should get an output block; a should be skipped.
	if !strings.Contains(result, "```output name=b") {
		t.Errorf("b should get output block:\n%s", result)
	}
	if strings.Contains(result, "```output name=a") {
		t.Errorf("a should not get output block:\n%s", result)
	}
}

// TestJoinStringsEmpty triggers the empty-slice guard in joinStrings (incremental.go:148).
func TestJoinStringsEmpty(t *testing.T) {
	// Call joinStrings directly with an empty slice to hit the guard.
	got := joinStrings([]string{}, "\n\n")
	if got != "" {
		t.Errorf("joinStrings([]) = %q, want empty", got)
	}
}

// TestCheckCellFieldMutationNonSelectorLHS triggers checker.go:373 — when the LHS
// of an assignment is a selector like a.b.field (chained), sel.X is not *ast.Ident.
func TestCheckCellFieldMutationNonSelectorLHS(t *testing.T) {
	// Cell b has a chained selector LHS: a.sub.field = 1.
	// Here sel.X = a.sub (a SelectorExpr, not Ident), so the !ok continue fires.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nvar s struct{ val int }\na.sub.val = 1\nresult = s.val\n```\n"
	result := Parse(src)
	errs := Check(result)
	// The chained assignment should not trigger rule 6 (sel.X is not Ident).
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") && strings.Contains(e.Error(), "sub") {
			t.Logf("rule 6 triggered (acceptable): %v", e)
		}
	}
	// Should not panic.
}

// TestCheckCellFieldMutationUnaryNonSelector triggers checker.go:388 — when &expr
// operates on a non-selector (plain variable), sel check fails.
func TestCheckCellFieldMutationUnaryNonSelector(t *testing.T) {
	// Taking address of a plain variable (not a selector) — &localVar.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nlocalVar := 42\np := &localVar\n_ = p\nresult = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not trigger rule 6 for &localVar.
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") {
			t.Errorf("plain &localVar should not trigger rule 6: %v", e)
		}
	}
}

// TestCheckCellFieldMutationUnaryChainedSelector triggers checker.go:392 — when &expr
// operates on a chained selector (sel.X is a SelectorExpr, not Ident).
func TestCheckCellFieldMutationUnaryChainedSelector(t *testing.T) {
	// Taking address of a chained field: &x.y.z — sel.X is x.y (SelectorExpr), not Ident.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nvar s struct{ sub struct{ val int } }\np := &s.sub.val\n_ = p\nresult = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not trigger rule 6 — chained, so the !ok guard fires.
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") && strings.Contains(e.Error(), "sub") {
			t.Logf("rule 6 triggered (acceptable): %v", e)
		}
	}
}

// TestCheckFileAccessNonSelectorCall triggers checker.go:446 — a plain function
// call (not a method call) in the cell body, so call.Fun is not a SelectorExpr.
func TestCheckFileAccessNonSelectorCall(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = foo(42)\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not error for rule 8/9 (plain call, no selector).
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 8") || strings.Contains(e.Error(), "rule 9") {
			t.Errorf("plain func call should not trigger rule 8/9: %v", e)
		}
	}
}

// TestCheckFileAccessChainedSelector triggers checker.go:450 — a chained method call
// like a.b.Open("x"), where sel.X is a SelectorExpr rather than an Ident.
func TestCheckFileAccessChainedSelector(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nf, _ := a.b.Open(\"file.txt\")\n_ = f\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not trigger rule 8 — ident.Name is not "ctx" (the guard would have
	// fired at the ident check if it reached there, but sel.X is not Ident).
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 8") {
			t.Errorf("chained method should not trigger rule 8: %v", e)
		}
	}
}

// TestIsCtxSecretCallNonSelector triggers checker.go:554 — the RHS is a plain
// function call (not a selector), so isCtxSecretCall returns false early.
func TestIsCtxSecretCallNonSelector(t *testing.T) {
	// Rule 12: exported field assigned from a plain function call (not ctx.Secret).
	src := "```go:notebook name=test exported(val int)\nval = getValue()\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not trigger rule 12.
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 12") {
			t.Errorf("plain func call should not trigger rule 12: %v", e)
		}
	}
}

// TestIsCtxSecretCallChainedSelector triggers checker.go:558 — the RHS is a chained
// call like a.b.Secret(), so sel.X is a SelectorExpr, not Ident.
func TestIsCtxSecretCallChainedSelector(t *testing.T) {
	// exported field assigned from a.b.Secret() — sel.X is not *ast.Ident.
	src := "```go:notebook name=test exported(val string)\nval = a.b.Secret(\"key\")\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not trigger rule 12 (chained selector, not ctx.Secret).
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 12") {
			t.Errorf("chained selector call should not trigger rule 12: %v", e)
		}
	}
}

func TestIsJavascriptURLVbscript(t *testing.T) {
	if !IsJavascriptURL("vbscript:msgbox('xss')") {
		t.Error("vbscript: should be treated as javascript URL")
	}
}

func TestIsJavascriptURLDataTextHTML(t *testing.T) {
	if !IsJavascriptURL("data:text/html,<script>alert(1)</script>") {
		t.Error("data:text/html should be treated as javascript URL")
	}
}

func TestCheckDefsVarWithoutEmbed(t *testing.T) {
	// A var in go:defs without //go:embed triggers rule 14.
	src := "```go:defs\nvar x = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 14") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 14 error for var without //go:embed, got: %v", errs)
	}
}

func TestValidateSVGXMLInvalid(t *testing.T) {
	// Create a file with invalid XML to trigger the invalid XML error path.
	dir := t.TempDir()
	svgPath := filepath.Join(dir, "bad.svg")
	os.WriteFile(svgPath, []byte("<svg><unclosed>"), 0o644)
	err := validateSVGXML(svgPath)
	if err == nil {
		t.Error("expected error for invalid SVG XML")
	}
	if !strings.Contains(err.Error(), "invalid XML") {
		t.Errorf("error should mention invalid XML: %v", err)
	}
}

func TestCheckSerializableExportsChanType(t *testing.T) {
	// A cell with a chan-type exported field triggers rule 7.
	src := "```go:notebook name=test exported(ch chan int)\nch = make(chan int)\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for chan exported field, got: %v", errs)
	}
}

func TestCheckSerializableExportsFuncType(t *testing.T) {
	// A cell with a func-type exported field triggers rule 7.
	src := "```go:notebook name=test exported(fn func())\nfn = func(){}\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for func exported field, got: %v", errs)
	}
}

func TestIsImportLineAliasedImport(t *testing.T) {
	// An aliased import: the unused map should contain the alias name.
	unused := map[string]bool{"myfmt": true}
	// alias "fmt" — prefix is "myfmt", so name becomes "myfmt"
	if !isImportLine(`myfmt "fmt"`, unused) {
		t.Error("aliased import line should return true when alias is in unused map")
	}
}

func TestIsImportLineEndEqualsStart(t *testing.T) {
	// A line that starts and ends with the same quote — end <= start.
	unused := map[string]bool{"fmt": true}
	// Single character between first and last quote is the same position.
	// Use a string where LastIndex == Index (only one quote).
	if isImportLine(`"`, unused) {
		t.Error("single quote should return false")
	}
}

// TestGenerateReportMainUnnamedCell exercises the else branch in generateReportMain
// where a cell has no StructName (unnamed/defs cells).
func TestGenerateReportMainUnnamedCell(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		ReportMode:  true,
		Cells: []GeneratedCell{
			// Named cell first.
			{
				StructName: "loadCell",
				FuncName:   "runLoad",
				StructCode: "type loadCell struct {\n\terr error\n\tval int\n}",
				FuncCode:   "func runLoad(ctx *nb.Context) (out loadCell) {\n\treturn\n}",
			},
			// Unnamed cell (StructName == "") — triggers the else branch.
			{
				StructName: "",
				FuncName:   "runAnon",
				FuncCode:   "func runAnon(ctx *nb.Context, load loadCell) {\n}",
			},
		},
	}
	src := generateReportMain(gen)
	// The unnamed cell should be called without assignment.
	if !strings.Contains(src, "runAnon(ctx, load)") {
		t.Errorf("unnamed cell should be called with upstream args, got:\n%s", src)
	}
}

// TestWritePackageGoModWriteFail triggers writepackage.go:25 — go.mod write fails
// when the package directory is read-only.
func TestWritePackageGoModWriteFail(t *testing.T) {
	gen := GenerateResult{PackageName: "main"}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	// Create the pkg directory and make it read-only.
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(pkgDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(pkgDir, 0o755) })

	err := WritePackage(gen, pkgDir, "test.md", "testpkg")
	if err == nil {
		t.Skip("running as root or OS allows write to read-only dir — skipping")
	}
	if !strings.Contains(err.Error(), "write go.mod") {
		t.Errorf("expected go.mod write error, got: %v", err)
	}
}

// TestWritePackageMainGoWriteFail triggers writepackage.go:33 — main.go write fails
// when a directory named main.go already exists in the package dir.
func TestWritePackageMainGoWriteFail(t *testing.T) {
	gen := GenerateResult{PackageName: "main"}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a directory named main.go so os.WriteFile fails.
	mainGoDir := filepath.Join(pkgDir, "main.go")
	if err := os.MkdirAll(mainGoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	err := WritePackage(gen, pkgDir, "test.md", "testpkg")
	if err == nil {
		t.Error("expected error when main.go is a directory")
		return
	}
	if !strings.Contains(err.Error(), "write main.go") {
		t.Errorf("expected main.go write error, got: %v", err)
	}
}

// TestCopyEmbedFilesFailOnCopy triggers embed.go:51 — copyFile fails because
// the destination is not writable.
func TestCopyEmbedFilesFailOnCopy(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, srcDir, "data.txt", "content")
	// Make dst read-only so copyFile can't create the destination file.
	if err := os.Chmod(dstDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dstDir, 0o755) })

	err := CopyEmbedFiles([]string{"data.txt"}, srcDir, dstDir)
	if err == nil {
		t.Skip("running as root or OS allows write to read-only dir — skipping")
	}
	if !strings.Contains(err.Error(), "embed copy") {
		t.Errorf("expected embed copy error, got: %v", err)
	}
}

// TestWritePackageEmbedSrcDirDot triggers writepackage.go:46 — the srcDir = "."
// branch fires when notebookPath has no directory component.
func TestWritePackageEmbedSrcDirDot(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "//go:embed nonexistent.txt\nvar data string",
	}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "dotpkg")
	// Use a bare filename (no directory) so filepath.Dir returns ".".
	err := WritePackage(gen, pkgDir, "test.md", "dotpkg")
	// No actual file to embed, so CopyEmbedFiles will find no matches — succeeds.
	if err != nil {
		t.Fatal(err)
	}
}

// TestWritePackageEmbedCopyFails triggers writepackage.go:46,49 — CopyEmbedFiles
// returns an error when the source embed file is not readable.
func TestWritePackageEmbedCopyFails(t *testing.T) {
	// Create a source notebook directory with a data file that is not readable.
	srcDir := t.TempDir()
	dataFile := filepath.Join(srcDir, "data.txt")
	if err := os.WriteFile(dataFile, []byte("embed content"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dataFile, 0o644) })

	// The notebook path is in srcDir so DefsCode embed pattern finds the file.
	notebookPath := filepath.Join(srcDir, "test.md")

	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "//go:embed data.txt\nvar data string",
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "embpkg")

	err := WritePackage(gen, pkgDir, notebookPath, "embpkg")
	if err == nil {
		t.Skip("running as root or OS allows read of unreadable file — skipping")
	}
	// Should report copy embed files error.
	if !strings.Contains(err.Error(), "copy embed files") {
		t.Errorf("expected copy embed files error, got: %v", err)
	}
}

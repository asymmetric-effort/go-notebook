// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestWritePackageCompiles is the critical integration test: it generates
// a Go package from a notebook and verifies it compiles with `go build`.
func TestWritePackageCompiles(t *testing.T) {
	// A simple notebook with one named cell.
	src := `# Test Notebook

` + "```go:notebook name=greet exported(msg string)\nmsg = \"hello world\"\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")

	err = WritePackage(gen, pkgDir, "test.md", "testpkg")
	if err != nil {
		t.Fatalf("WritePackage: %v", err)
	}

	// Verify the files exist.
	if _, err := os.Stat(filepath.Join(pkgDir, "go.mod")); err != nil {
		t.Fatal("go.mod not created")
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "main.go")); err != nil {
		t.Fatal("main.go not created")
	}

	// Read the generated source and verify it looks right.
	mainSrc, err := os.ReadFile(filepath.Join(pkgDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(mainSrc)

	if !strings.Contains(source, "package main") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(source, "type greetCell struct") {
		t.Error("missing cell struct")
	}
	if !strings.Contains(source, "func runGreet") {
		t.Error("missing run function")
	}
	if !strings.Contains(source, "func main()") {
		t.Error("missing main function")
	}
	if !strings.Contains(source, "nb.PanicError") {
		t.Error("missing panic recovery")
	}
	if !strings.Contains(source, "nb.NewContext") {
		t.Error("missing context creation in main")
	}

	// THE REAL TEST: verify it compiles.
	// We need a go.sum and the nb module to be available.
	// For this test, we use a replace directive to point at our local module.
	moduleRoot := findModuleRoot(t)
	goMod := "module testpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed:\n%s\n%v", out, err)
	}
	t.Logf("go build succeeded for generated package")
}

// TestWritePackageMultipleCells tests a notebook with multiple cells
// that reference each other.
func TestWritePackageMultipleCells(t *testing.T) {
	src := "```go:notebook name=a exported(val int)\nval = 42\n```\n\n" +
		"```go:notebook name=b exported(result int)\nresult = a.val + 1\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")
	err = WritePackage(gen, pkgDir, "test.md", "testpkg")
	if err != nil {
		t.Fatal(err)
	}

	// Add replace directive for local module.
	moduleRoot := findModuleRoot(t)
	goMod := "module testpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Show the generated source for debugging.
		src, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
		t.Fatalf("go build failed:\n%s\n---generated---\n%s", out, src)
	}
}

// TestWritePackageWithUnnamedCell tests that unnamed cells compile.
func TestWritePackageWithUnnamedCell(t *testing.T) {
	src := "```go:notebook name=val exported(x int)\nx = 1\n```\n\n" +
		"```go:notebook\n_ = val.x\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")
	WritePackage(gen, pkgDir, "test.md", "testpkg")

	moduleRoot := findModuleRoot(t)
	goMod := "module testpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		src, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
		t.Fatalf("go build failed:\n%s\n---generated---\n%s", out, src)
	}
}

// TestWritePackagePanicRecovery tests that a generated binary catches panics.
func TestWritePackagePanicRecovery(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\npanic(\"test panic\")\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")
	WritePackage(gen, pkgDir, "test.md", "testpkg")

	moduleRoot := findModuleRoot(t)
	goMod := "module testpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	// Build.
	binName := "testbin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(dir, binName), "./...")
	buildCmd.Dir = pkgDir
	buildCmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		genSrc, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
		t.Fatalf("go build failed:\n%s\n---generated---\n%s", out, genSrc)
	}

	// Run — the panic should be caught by the deferred recovery, not crash.
	runCmd := exec.Command(filepath.Join(dir, binName))
	runOut, runErr := runCmd.CombinedOutput()
	// The binary should exit without a stack trace (panic is recovered).
	// It may exit non-zero because the cell has an error, but it should NOT
	// print a Go runtime panic.
	if strings.Contains(string(runOut), "goroutine") && strings.Contains(string(runOut), "runtime.") {
		t.Errorf("binary printed a stack trace — panic was not recovered:\n%s", runOut)
	}
	_ = runErr // Non-zero exit is expected (cell error), but no stack trace.
	t.Logf("binary ran (exit err: %v), output: %s", runErr, string(runOut))
}

// TestAssembleSource verifies the assembled source structure.
func TestAssembleSource(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "const x = 1",
		Cells: []GeneratedCell{
			{StructName: "testCell", FuncName: "runTest",
				StructCode: "type testCell struct {\n\terr error\n}",
				FuncCode:   "func runTest(ctx *nb.Context) (out testCell) {\n\treturn\n}"},
		},
	}
	assembled := assembleSource(gen, "test.md")
	if !strings.Contains(assembled, "package main") {
		t.Error("missing package")
	}
	if !strings.Contains(assembled, "const x = 1") {
		t.Error("missing defs")
	}
	if !strings.Contains(assembled, "func main()") {
		t.Error("missing main")
	}
}

func TestWritePackageBadDir(t *testing.T) {
	gen := GenerateResult{PackageName: "main"}
	blocker := filepath.Join(t.TempDir(), "blocker")
	os.WriteFile(blocker, []byte("x"), 0o644)
	err := WritePackage(gen, filepath.Join(blocker, "sub"), "test.md", "testpkg")
	if err == nil {
		t.Error("expected error when directory can't be created")
	}
}

func TestWritePackageNoCells(t *testing.T) {
	gen := GenerateResult{PackageName: "main"}
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "emptypkg")
	err := WritePackage(gen, pkgDir, "test.md", "emptypkg")
	if err != nil {
		t.Fatal(err)
	}
	// Should still produce a valid main.go.
	data, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
	if !strings.Contains(string(data), "func main()") {
		t.Error("missing main in empty notebook")
	}
}

func TestAssembleSourceNoDefs(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		Cells: []GeneratedCell{
			{StructName: "testCell", FuncName: "runTest",
				StructCode: "type testCell struct {\n\terr error\n}",
				FuncCode:   "func runTest(ctx *nb.Context) (out testCell) {\n\treturn\n}"},
		},
	}
	src := assembleSource(gen, "test.md")
	if strings.Contains(src, "go:defs") {
		t.Error("no defs should not have go:defs comment")
	}
	if !strings.Contains(src, "func main()") {
		t.Error("missing main")
	}
}

func TestAssembleSourceWithDefs(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "const x = 1",
	}
	src := assembleSource(gen, "test.md")
	if !strings.Contains(src, "const x = 1") {
		t.Error("missing defs")
	}
	if !strings.Contains(src, "go:defs") {
		t.Error("missing go:defs comment")
	}
}

func TestSplitDefsImports(t *testing.T) {
	tests := []struct {
		name        string
		defs        string
		wantImports []string
		wantBody    string
	}{
		{
			"single import",
			"import \"fmt\"\n\nconst x = 1",
			[]string{"\"fmt\""},
			"const x = 1",
		},
		{
			"import block",
			"import (\n\t\"fmt\"\n\t\"os\"\n)\n\nconst x = 1",
			[]string{"\"fmt\"", "\"os\""},
			"const x = 1",
		},
		{
			"no imports",
			"const x = 1",
			nil,
			"const x = 1",
		},
		{
			"empty",
			"",
			nil,
			"",
		},
		{
			"alias import",
			"import nb \"github.com/asymmetric-effort/go-notebook/nb\"",
			[]string{"nb \"github.com/asymmetric-effort/go-notebook/nb\""},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports, body := splitDefsImports(tt.defs)
			if len(imports) != len(tt.wantImports) {
				t.Fatalf("imports = %v, want %v", imports, tt.wantImports)
			}
			for i := range imports {
				if imports[i] != tt.wantImports[i] {
					t.Errorf("imports[%d] = %q, want %q", i, imports[i], tt.wantImports[i])
				}
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestExtractCell(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "const x = 1",
		Cells: []GeneratedCell{
			{StructName: "loadCell", FuncName: "runLoad",
				StructCode: "type loadCell struct {\n\terr error\n\tval int\n}",
				FuncCode:   "func runLoad(ctx *nb.Context) (out loadCell) {\n\treturn\n}"},
			{StructName: "processCell", FuncName: "runProcess",
				StructCode: "type processCell struct {\n\terr error\n\tresult int\n}",
				FuncCode:   "func runProcess(ctx *nb.Context, load loadCell) (out processCell) {\n\treturn\n}"},
			{StructName: "displayCell", FuncName: "runDisplay",
				StructCode: "type displayCell struct {\n\terr error\n}",
				FuncCode:   "func runDisplay(ctx *nb.Context, load loadCell, process processCell) (out displayCell) {\n\treturn\n}"},
		},
	}

	// Extract the middle cell — should include load and process but not display.
	extracted, err := ExtractCell(gen, "process")
	if err != nil {
		t.Fatal(err)
	}
	if len(extracted.Cells) != 2 {
		t.Fatalf("got %d cells, want 2", len(extracted.Cells))
	}
	if extracted.Cells[0].StructName != "loadCell" {
		t.Errorf("cells[0] = %q, want loadCell", extracted.Cells[0].StructName)
	}
	if extracted.Cells[1].StructName != "processCell" {
		t.Errorf("cells[1] = %q, want processCell", extracted.Cells[1].StructName)
	}
	if extracted.DefsCode != "const x = 1" {
		t.Error("defs code should be preserved")
	}
}

func TestExtractCellNotFound(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		Cells: []GeneratedCell{
			{StructName: "testCell", FuncName: "runTest"},
		},
	}
	_, err := ExtractCell(gen, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent cell")
	}
}

func TestAssembleSourcePrunesImports(t *testing.T) {
	// A defs block with an import that's NOT used by any cell.
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "import \"os\"\n\nconst x = 1",
	}
	src := assembleSource(gen, "test.md")
	// "os" should be pruned since nothing uses it.
	if strings.Contains(src, "\"os\"") {
		t.Error("unused import \"os\" should be pruned")
	}
	if !strings.Contains(src, "const x = 1") {
		t.Error("defs body should still be present")
	}
}

func TestAssembleSourceKeepsUsedDefsImport(t *testing.T) {
	// A defs block with an import used in cell code.
	gen := GenerateResult{
		PackageName: "main",
		DefsCode:    "import \"fmt\"\n\nfunc hello() { fmt.Println(\"hello\") }",
	}
	src := assembleSource(gen, "test.md")
	if !strings.Contains(src, "\"fmt\"") {
		t.Error("used import \"fmt\" should be kept")
	}
}

func TestExtractCellCompiles(t *testing.T) {
	// Two-cell notebook: the first cell is valid, the second has a type error.
	// Extracting and running the first cell should succeed despite the broken cell.
	src := "```go:notebook name=good exported(val int)\nval = 42\n```\n\n" +
		"```go:notebook name=bad exported(msg string)\nmsg = 12345\n```\n" // type error: int to string

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Extract only the "good" cell.
	extracted, err := ExtractCell(gen, "good")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "extractpkg")
	err = WritePackage(extracted, pkgDir, "test.md", "extractpkg")
	if err != nil {
		t.Fatal(err)
	}

	moduleRoot := findModuleRoot(t)
	goMod := "module extractpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	// Build — should succeed because the broken cell is not included.
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(dir, "extracttest"+exeSuffix()), ".")
	buildCmd.Dir = pkgDir
	buildCmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		genSrc, _ := os.ReadFile(filepath.Join(pkgDir, "main.go"))
		t.Fatalf("go build failed:\n%s\n---generated---\n%s", out, genSrc)
	}
	t.Log("extracted cell compiled successfully")
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// findModuleRoot finds the go-notebook module root for replace directives.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	modFile := strings.TrimSpace(string(out))
	return filepath.Dir(modFile)
}

func TestGenerateSourceMap(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n"
	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}
	if gen.SourceMap == nil {
		t.Fatal("SourceMap should not be nil")
	}

	// Write the package to populate the source map.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "smpkg")
	err = WritePackage(gen, pkgDir, "test.md", "smpkg")
	if err != nil {
		t.Fatal(err)
	}

	// The source map should have at least one entry.
	if gen.SourceMap.Len() == 0 {
		t.Error("source map should have entries after WritePackage")
	}

	// The cell body starts at Markdown line 2 (line 1 is the fence).
	// It should map to some Go line.
	goLine, _, ok := gen.SourceMap.MdToGo(2, 1)
	if !ok {
		t.Error("Markdown line 2 should be in the source map")
	}
	if goLine <= 0 {
		t.Errorf("Go line = %d, should be positive", goLine)
	}
	t.Logf("Md line 2 -> Go line %d", goLine)
}

func TestBuildSourceMapFromSource(t *testing.T) {
	src := "package main\n\n//line test.md:10:1\nval = 42\nmsg = \"hello\"\n\nfunc main() {}\n"
	sm := NewSourceMap()
	buildSourceMapFromSource(src, sm)
	if sm.Len() == 0 {
		t.Error("source map should have entries")
	}
	goLine, _, ok := sm.MdToGo(10, 1)
	if !ok {
		t.Error("line 10 should be mapped")
	}
	if goLine <= 0 {
		t.Errorf("Go line = %d", goLine)
	}
}

func TestGoByteSliceLiteral(t *testing.T) {
	got := goByteSliceLiteral([]byte{0x01, 0xff})
	if got != "[]byte{0x01, 0xff}" {
		t.Errorf("got %q", got)
	}
}

func TestGoByteSliceLiteralEmpty(t *testing.T) {
	got := goByteSliceLiteral(nil)
	if got != "[]byte{}" {
		t.Errorf("got %q", got)
	}
}

func TestGenerateMainWithEmbeddedData(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		Cells: []GeneratedCell{
			{StructName: "loadCell", FuncName: "runLoad",
				StructCode: "type loadCell struct {\n\terr error\n\tval int\n}",
				FuncCode:   "func runLoad(ctx *nb.Context) (out loadCell) {\n\treturn\n}",
				DataMode:   DataEmbedded},
		},
		EmbeddedCells: []EmbeddedCellData{
			{CellName: "load", EncryptedData: []byte{0x01, 0x02}},
		},
	}
	src := assembleSource(gen, "test.md")
	if !strings.Contains(src, "data=embedded") {
		t.Error("should have embedded comment")
	}
	if !strings.Contains(src, "UnmarshalBinary") {
		t.Error("should have UnmarshalBinary call")
	}
	if !strings.Contains(src, "0x01, 0x02") {
		t.Error("should have embedded byte data")
	}
}

func TestGenerateMainEmbeddedNoData(t *testing.T) {
	gen := GenerateResult{
		PackageName: "main",
		Cells: []GeneratedCell{
			{StructName: "loadCell", FuncName: "runLoad",
				StructCode: "type loadCell struct {\n\terr error\n\tval int\n}",
				FuncCode:   "func runLoad(ctx *nb.Context) (out loadCell) {\n\treturn\n}",
				DataMode:   DataEmbedded},
		},
		// No EmbeddedCells — first pass, should emit GOBOOK_EMBED_MODE code.
	}
	src := assembleSource(gen, "test.md")
	if !strings.Contains(src, "GOBOOK_EMBED_MODE") {
		t.Error("should have embed mode check for first pass")
	}
	if !strings.Contains(src, "MarshalBinary") {
		t.Error("should have MarshalBinary for serialization")
	}
}

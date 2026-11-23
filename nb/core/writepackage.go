// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WritePackage writes a complete, compilable Go package from a GenerateResult
// to the specified directory. The package includes a go.mod, the generated
// source file, and a main function that runs all cells in order.
//
// The generated package can be compiled with `go build` and run with `go run`.
func WritePackage(gen GenerateResult, dir string, notebookPath string, modulePath string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create package dir: %w", err)
	}

	// Write go.mod.
	goMod := fmt.Sprintf("module %s\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n", modulePath)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	// Assemble the main source file.
	src := assembleSource(gen, notebookPath)

	mainPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainPath, []byte(src), 0o644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	// Build source map from the generated source (uses //line directives).
	if gen.SourceMap != nil {
		buildSourceMapFromSource(src, gen.SourceMap)
	}

	// Copy //go:embed files from alongside the notebook into the package dir.
	patterns := ExtractEmbedPatterns(gen.DefsCode)
	if len(patterns) > 0 {
		srcDir := filepath.Dir(notebookPath)
		if srcDir == "" || srcDir == "." {
			srcDir = "."
		}
		if err := CopyEmbedFiles(patterns, srcDir, dir); err != nil {
			return fmt.Errorf("copy embed files: %w", err)
		}
	}

	return nil
}

// assembleSource combines all parts of the GenerateResult into a single
// compilable Go source file. Import pruning ensures no unused imports.
func assembleSource(gen GenerateResult, notebookPath string) string {
	var buf strings.Builder

	// Package declaration.
	fmt.Fprintf(&buf, "package %s\n\n", gen.PackageName)

	// Collect all imports into a single block.
	hasMarshal := false
	hasEmbedded := false
	hasSocket := gen.SocketMode
	for _, cell := range gen.Cells {
		if cell.StructCode != "" {
			hasMarshal = true
		}
		if cell.DataMode == DataEmbedded {
			hasEmbedded = true
		}
	}

	// Extract imports from defs code and separate them from other declarations.
	defsImports, defsBody := splitDefsImports(gen.DefsCode)

	// Build the import block, deduplicating between marshal imports and defs imports.
	seen := make(map[string]bool)
	fmt.Fprintf(&buf, "import (\n")
	if hasMarshal {
		for _, pkg := range []string{"\"bytes\"", "\"encoding/gob\"", "\"fmt\""} {
			fmt.Fprintf(&buf, "\t%s\n", pkg)
			seen[pkg] = true
		}
	}
	if hasEmbedded {
		for _, pkg := range []string{"\"encoding/base64\"", "\"os\""} {
			if !seen[pkg] {
				fmt.Fprintf(&buf, "\t%s\n", pkg)
				seen[pkg] = true
			}
		}
	}
	if hasSocket {
		for _, pkg := range []string{"\"os\"", "\"log\""} {
			if !seen[pkg] {
				fmt.Fprintf(&buf, "\t%s\n", pkg)
				seen[pkg] = true
			}
		}
	}
	fmt.Fprintf(&buf, "\tnb \"github.com/asymmetric-effort/go-notebook/nb\"\n")
	if hasSocket {
		fmt.Fprintf(&buf, "\t\"github.com/asymmetric-effort/go-notebook/nb/kernel\"\n")
	}
	seen["nb \"github.com/asymmetric-effort/go-notebook/nb\""] = true
	// Add imports extracted from defs blocks, skipping duplicates.
	for _, imp := range defsImports {
		if !seen[imp] {
			fmt.Fprintf(&buf, "\t%s\n", imp)
			seen[imp] = true
		}
	}
	fmt.Fprintf(&buf, ")\n\n")

	// Defs code (non-import declarations from go:defs blocks).
	if defsBody != "" {
		fmt.Fprintf(&buf, "// --- go:defs ---\n\n")
		fmt.Fprintf(&buf, "%s\n\n", defsBody)
	}

	if hasMarshal {
		// Suppress unused import warnings when no marshal calls happen.
		fmt.Fprintf(&buf, "var _ = bytes.NewReader\nvar _ = gob.NewEncoder\nvar _ = fmt.Sprintf\n\n")
	}
	if hasEmbedded {
		fmt.Fprintf(&buf, "var _ = base64.StdEncoding\nvar _ = os.Getenv\n\n")
	}
	if hasSocket {
		fmt.Fprintf(&buf, "var _ = kernel.DialSocket\nvar _ = log.Fatal\nvar _ = os.Getenv\n\n")
	}

	// Cell structs, marshal methods, and functions.
	for _, cell := range gen.Cells {
		if cell.StructCode != "" {
			fmt.Fprintf(&buf, "%s\n\n", cell.StructCode)
			// Generate marshal methods for named cells (deep copy support).
			fields := extractStructFields(cell.StructCode)
			if len(fields) > 0 {
				fmt.Fprintf(&buf, "%s\n\n", GenerateMarshal(cell.StructName, fields))
			}
		}
		fmt.Fprintf(&buf, "%s\n\n", cell.FuncCode)
	}

	// Main function that runs all cells in order.
	buf.WriteString(generateMain(gen, notebookPath))

	// Prune unused imports from the assembled source.
	src := buf.String()
	pruned, err := PruneImports(src)
	if err == nil {
		return pruned
	}
	return src
}

// splitDefsImports separates import declarations from other code in defs blocks.
// Returns the import specs (ready to include in an import block) and the
// remaining non-import code.
func splitDefsImports(defsCode string) (imports []string, body string) {
	if defsCode == "" {
		return nil, ""
	}
	var bodyLines []string
	lines := strings.Split(defsCode, "\n")
	inImportBlock := false
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "import (" {
			inImportBlock = true
			continue
		}
		if inImportBlock {
			if trimmed == ")" {
				inImportBlock = false
				continue
			}
			if trimmed != "" {
				imports = append(imports, trimmed)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "import ") && !strings.HasPrefix(trimmed, "import (") {
			// Single-line import: import "fmt"
			spec := strings.TrimPrefix(trimmed, "import ")
			imports = append(imports, spec)
			continue
		}
		bodyLines = append(bodyLines, lines[i])
	}
	return imports, strings.TrimSpace(strings.Join(bodyLines, "\n"))
}

// buildSourceMapFromSource scans generated Go source for //line directives
// and builds a source map. Each //line directive maps a Markdown position
// to the corresponding Go line in the generated source.
func buildSourceMapFromSource(src string, sm *SourceMap) {
	lines := strings.Split(src, "\n")
	for goLine, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//line ") {
			continue
		}
		// Parse: //line filename:mdLine:col
		rest := strings.TrimPrefix(trimmed, "//line ")
		// Find the last two colons (line:col).
		lastColon := strings.LastIndex(rest, ":")
		if lastColon < 0 {
			continue
		}
		beforeCol := rest[:lastColon]
		secondLastColon := strings.LastIndex(beforeCol, ":")
		if secondLastColon < 0 {
			continue
		}
		mdLineStr := beforeCol[secondLastColon+1:]
		mdLine := 0
		for _, ch := range mdLineStr {
			if ch >= '0' && ch <= '9' {
				mdLine = mdLine*10 + int(ch-'0')
			}
		}
		if mdLine > 0 {
			// Count how many non-directive lines follow until the next directive or function end.
			bodyLen := 0
			for j := goLine + 1; j < len(lines); j++ {
				t := strings.TrimSpace(lines[j])
				if strings.HasPrefix(t, "//line ") || t == "return" || t == "}" {
					break
				}
				bodyLen++
			}
			if bodyLen > 0 {
				sm.Add(mdLine, 1, goLine+2, 1, bodyLen) // goLine is 0-based, +1 for 1-based, +1 for line after directive
			}
		}
	}
}

// extractStructFields extracts field declarations from a struct definition string.
// Input like "type testCell struct {\n\terr error\n\tval int\n}" returns ["err error", "val int"].
func extractStructFields(structCode string) []string {
	var fields []string
	lines := strings.Split(structCode, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "type ") || trimmed == "}" {
			continue
		}
		fields = append(fields, trimmed)
	}
	return fields
}

// generateMain creates the main() function that executes cells in document
// order, checks errors, and renders the report.
func generateMain(gen GenerateResult, notebookPath string) string {
	if gen.SocketMode {
		return generateSocketMain(gen)
	}
	return generateBatchMain(gen, notebookPath)
}

// generateBatchMain creates a main() that runs all cells sequentially
// in one process (the default batch/report mode).
func generateBatchMain(gen GenerateResult, notebookPath string) string {
	var buf strings.Builder

	// Build a map of embedded cell data by name for quick lookup.
	embeddedByName := make(map[string][]byte)
	for _, ec := range gen.EmbeddedCells {
		embeddedByName[ec.CellName] = ec.EncryptedData
	}

	fmt.Fprintf(&buf, "func main() {\n")
	fmt.Fprintf(&buf, "\tctx := nb.NewContext(nil)\n")
	fmt.Fprintf(&buf, "\t_ = ctx\n") // Prevent unused variable if no cells.

	// Track all named cell variables in order for passing as parameters.
	var namedVars []string
	for _, cell := range gen.Cells {
		if cell.StructName != "" {
			varName := strings.TrimSuffix(cell.StructName, "Cell")

			if cell.DataMode == DataEmbedded {
				if encData, ok := embeddedByName[varName]; ok && len(encData) > 0 {
					// Embedded cell: decrypt and unmarshal instead of running.
					fmt.Fprintf(&buf, "\t// %s is data=embedded — using encrypted embedded data.\n", varName)
					fmt.Fprintf(&buf, "\tvar %s %s\n", varName, cell.StructName)
					fmt.Fprintf(&buf, "\t{\n")
					fmt.Fprintf(&buf, "\t\tembeddedData := %s\n", goByteSliceLiteral(encData))
					fmt.Fprintf(&buf, "\t\tif err := %s.UnmarshalBinary(embeddedData); err != nil {\n", varName)
					fmt.Fprintf(&buf, "\t\t\t%s.err = fmt.Errorf(\"unmarshal embedded %s: %%w\", err)\n", varName, varName)
					fmt.Fprintf(&buf, "\t\t}\n")
					fmt.Fprintf(&buf, "\t}\n")
				} else {
					// Embedded cell but no data yet — run normally and output serialized data.
					params := "ctx"
					if len(namedVars) > 0 {
						params = "ctx, " + strings.Join(namedVars, ", ")
					}
					fmt.Fprintf(&buf, "\t%s := %s(%s)\n", varName, cell.FuncName, params)
					// In embed mode, serialize and output the cell data.
					fmt.Fprintf(&buf, "\tif os.Getenv(\"GOBOOK_EMBED_MODE\") == \"1\" {\n")
					fmt.Fprintf(&buf, "\t\tif data, err := %s.MarshalBinary(); err == nil {\n", varName)
					fmt.Fprintf(&buf, "\t\t\tfmt.Printf(\"GOBOOK_EMBED:%s:%%s\\n\", base64.StdEncoding.EncodeToString(data))\n", varName)
					fmt.Fprintf(&buf, "\t\t}\n")
					fmt.Fprintf(&buf, "\t}\n")
				}
			} else {
				// Normal cell: call with all previous named cells, check err.
				params := "ctx"
				if len(namedVars) > 0 {
					params = "ctx, " + strings.Join(namedVars, ", ")
				}
				fmt.Fprintf(&buf, "\t%s := %s(%s)\n", varName, cell.FuncName, params)
			}

			fmt.Fprintf(&buf, "\tif %s.err != nil {\n", varName)
			fmt.Fprintf(&buf, "\t\tnb.NewCellError(%q, 0, %s.err)\n", varName, varName)
			fmt.Fprintf(&buf, "\t\treturn\n")
			fmt.Fprintf(&buf, "\t}\n")
			namedVars = append(namedVars, varName)
		} else {
			// Unnamed cell: call with all named cells.
			params := "ctx"
			if len(namedVars) > 0 {
				params = "ctx, " + strings.Join(namedVars, ", ")
			}
			fmt.Fprintf(&buf, "\t%s(%s)\n", cell.FuncName, params)
		}
	}

	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// generateSocketMain creates a main() for notebook mode. The cell process
// connects to the kernel's Unix socket (GOBOOK_SOCKET env var), fetches
// upstream cell data, runs the target cell, and sends the result back.
func generateSocketMain(gen GenerateResult) string {
	var buf strings.Builder

	// Identify the target cell (last named cell) and upstream cells.
	targetCell := gen.TargetCell
	var upstreamCells []GeneratedCell
	var targetGen *GeneratedCell

	for i := range gen.Cells {
		cell := &gen.Cells[i]
		if cell.StructName == "" {
			continue
		}
		varName := strings.TrimSuffix(cell.StructName, "Cell")
		if varName == targetCell {
			targetGen = cell
		} else {
			upstreamCells = append(upstreamCells, *cell)
		}
	}

	fmt.Fprintf(&buf, "func main() {\n")
	fmt.Fprintf(&buf, "\tsocketPath := os.Getenv(\"GOBOOK_SOCKET\")\n")
	fmt.Fprintf(&buf, "\tif socketPath == \"\" {\n")
	fmt.Fprintf(&buf, "\t\tlog.Fatal(\"GOBOOK_SOCKET not set\")\n")
	fmt.Fprintf(&buf, "\t}\n\n")

	// Connect to the kernel socket.
	fmt.Fprintf(&buf, "\tclient, err := kernel.DialSocket(socketPath)\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"connect to kernel: %%v\", err)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tdefer client.Close()\n\n")

	fmt.Fprintf(&buf, "\tctx := nb.NewContext(nil)\n")
	fmt.Fprintf(&buf, "\t_ = ctx\n\n")

	// Fetch and unmarshal each upstream cell from the kernel cache.
	var namedVars []string
	for _, cell := range upstreamCells {
		varName := strings.TrimSuffix(cell.StructName, "Cell")
		fmt.Fprintf(&buf, "\t// Fetch upstream cell %q from kernel.\n", varName)
		fmt.Fprintf(&buf, "\t%sData, err := client.FetchUpstream(%q)\n", varName, varName)
		fmt.Fprintf(&buf, "\tif err != nil {\n")
		fmt.Fprintf(&buf, "\t\t_ = client.SendError(fmt.Sprintf(\"fetch upstream %s: %%v\", err))\n", varName)
		fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"fetch upstream %s: %%v\", err)\n", varName)
		fmt.Fprintf(&buf, "\t}\n")
		fmt.Fprintf(&buf, "\tvar %s %s\n", varName, cell.StructName)
		fmt.Fprintf(&buf, "\tif err := %s.UnmarshalBinary(%sData); err != nil {\n", varName, varName)
		fmt.Fprintf(&buf, "\t\t_ = client.SendError(fmt.Sprintf(\"unmarshal %s: %%v\", err))\n", varName)
		fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"unmarshal %s: %%v\", err)\n", varName)
		fmt.Fprintf(&buf, "\t}\n\n")
		namedVars = append(namedVars, varName)
	}

	// Run the target cell.
	if targetGen != nil && targetGen.StructName != "" {
		params := "ctx"
		if len(namedVars) > 0 {
			params = "ctx, " + strings.Join(namedVars, ", ")
		}
		fmt.Fprintf(&buf, "\t// Run target cell %q.\n", targetCell)
		fmt.Fprintf(&buf, "\t%s := %s(%s)\n", targetCell, targetGen.FuncName, params)
		fmt.Fprintf(&buf, "\tif %s.err != nil {\n", targetCell)
		fmt.Fprintf(&buf, "\t\t_ = client.SendError(%s.err.Error())\n", targetCell)
		fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"cell %s error: %%v\", %s.err)\n", targetCell, targetCell)
		fmt.Fprintf(&buf, "\t}\n\n")

		// Send display output from ctx.
		fmt.Fprintf(&buf, "\t// Send display output to kernel.\n")
		fmt.Fprintf(&buf, "\tfor _, d := range ctx.Displays() {\n")
		fmt.Fprintf(&buf, "\t\tif err := client.SendDisplay(d.Mime, []byte(d.Data)); err != nil {\n")
		fmt.Fprintf(&buf, "\t\t\tlog.Fatalf(\"send display: %%v\", err)\n")
		fmt.Fprintf(&buf, "\t\t}\n")
		fmt.Fprintf(&buf, "\t}\n\n")

		// Marshal and send the result.
		fmt.Fprintf(&buf, "\t// Send serialized result to kernel.\n")
		fmt.Fprintf(&buf, "\tresultData, err := %s.MarshalBinary()\n", targetCell)
		fmt.Fprintf(&buf, "\tif err != nil {\n")
		fmt.Fprintf(&buf, "\t\t_ = client.SendError(fmt.Sprintf(\"marshal result: %%v\", err))\n")
		fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"marshal result: %%v\", err)\n")
		fmt.Fprintf(&buf, "\t}\n")
		fmt.Fprintf(&buf, "\tif err := client.SendResult(%q, resultData); err != nil {\n", targetCell)
		fmt.Fprintf(&buf, "\t\tlog.Fatalf(\"send result: %%v\", err)\n")
		fmt.Fprintf(&buf, "\t}\n")
	}

	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// goByteSliceLiteral converts a byte slice to a Go source literal.
func goByteSliceLiteral(data []byte) string {
	if len(data) == 0 {
		return "[]byte{}"
	}
	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("0x%02x", b))
	}
	return "[]byte{" + strings.Join(parts, ", ") + "}"
}

// ExtractCell creates a GenerateResult containing only the target cell and
// its upstream named-cell dependencies. This enables ad hoc runs of a single
// cell without compiling the entire notebook. Unnamed cells before the target
// are skipped (they are side-effect-only and not referenced by name).
func ExtractCell(gen GenerateResult, cellName string) (GenerateResult, error) {
	// Find the target cell.
	targetIdx := -1
	for i, cell := range gen.Cells {
		name := strings.TrimSuffix(cell.StructName, "Cell")
		if name == cellName {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return gen, fmt.Errorf("cell %q not found", cellName)
	}

	// Include all named cells up to and including the target, since they
	// may be upstream dependencies (passed as parameters).
	var cells []GeneratedCell
	for i := 0; i <= targetIdx; i++ {
		if gen.Cells[i].StructName != "" {
			cells = append(cells, gen.Cells[i])
		}
	}

	return GenerateResult{
		PackageName: gen.PackageName,
		Imports:     gen.Imports,
		DefsCode:    gen.DefsCode,
		Cells:       cells,
	}, nil
}

// ExtractCellForSocket creates a GenerateResult for notebook-mode socket
// execution. The generated main connects to the kernel socket, fetches
// upstream cells, runs the target, and sends the result back.
func ExtractCellForSocket(gen GenerateResult, cellName string) (GenerateResult, error) {
	result, err := ExtractCell(gen, cellName)
	if err != nil {
		return result, err
	}
	result.SocketMode = true
	result.TargetCell = cellName
	return result, nil
}

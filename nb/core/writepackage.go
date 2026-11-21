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

	return nil
}

// assembleSource combines all parts of the GenerateResult into a single
// compilable Go source file.
func assembleSource(gen GenerateResult, notebookPath string) string {
	var buf strings.Builder

	// Package declaration.
	fmt.Fprintf(&buf, "package %s\n\n", gen.PackageName)

	// Imports — always include the nb runtime and marshal dependencies.
	hasMarshal := false
	for _, cell := range gen.Cells {
		if cell.StructCode != "" {
			hasMarshal = true
			break
		}
	}
	fmt.Fprintf(&buf, "import (\n")
	if hasMarshal {
		// Only add marshal imports that aren't already in the defs code.
		if !strings.Contains(gen.DefsCode, "\"bytes\"") {
			fmt.Fprintf(&buf, "\t\"bytes\"\n")
		}
		if !strings.Contains(gen.DefsCode, "\"encoding/gob\"") {
			fmt.Fprintf(&buf, "\t\"encoding/gob\"\n")
		}
		if !strings.Contains(gen.DefsCode, "\"fmt\"") {
			fmt.Fprintf(&buf, "\t\"fmt\"\n")
		}
	}
	fmt.Fprintf(&buf, "\tnb \"github.com/asymmetric-effort/go-notebook/nb\"\n")
	fmt.Fprintf(&buf, ")\n\n")

	// Defs code (types, constants, functions from go:defs blocks).
	// Must come before var declarations since defs may contain imports.
	if gen.DefsCode != "" {
		fmt.Fprintf(&buf, "// --- go:defs ---\n\n")
		fmt.Fprintf(&buf, "%s\n\n", gen.DefsCode)
	}

	if hasMarshal {
		// Suppress unused import warnings when no marshal calls happen.
		fmt.Fprintf(&buf, "var _ = bytes.NewReader\nvar _ = gob.NewEncoder\nvar _ = fmt.Sprintf\n\n")
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

	return buf.String()
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
	var buf strings.Builder

	fmt.Fprintf(&buf, "func main() {\n")
	fmt.Fprintf(&buf, "\tctx := nb.NewContext(nil)\n")
	fmt.Fprintf(&buf, "\t_ = ctx\n") // Prevent unused variable if no cells.

	// Track all named cell variables in order for passing as parameters.
	var namedVars []string
	for _, cell := range gen.Cells {
		if cell.StructName != "" {
			// Named cell: call with all previous named cells, check err.
			varName := strings.TrimSuffix(cell.StructName, "Cell")
			params := "ctx"
			if len(namedVars) > 0 {
				params = "ctx, " + strings.Join(namedVars, ", ")
			}
			fmt.Fprintf(&buf, "\t%s := %s(%s)\n", varName, cell.FuncName, params)
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

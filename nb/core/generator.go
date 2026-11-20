// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"strings"
)

// upstreamCell tracks a named cell for use as a parameter in downstream cells.
type upstreamCell struct {
	name       string
	structName string
	attrs      CellAttrs
}

// GeneratedCell holds the generated Go source for a single cell.
type GeneratedCell struct {
	StructName string // e.g., "loadCell"
	FuncName   string // e.g., "runLoad"
	StructCode string // The struct type definition.
	FuncCode   string // The run function definition.
}

// GenerateResult holds the complete generated output for a notebook.
type GenerateResult struct {
	PackageName string
	Imports     string
	DefsCode    string
	Cells       []GeneratedCell
	MainCode    string
}

// Generate produces Go source code from parsed notebook blocks.
// It creates cell structs and run functions per the specification:
// - Named cells get a struct with err + exported fields and a runX function
// - Unnamed cells get a function with no return
// - Each function has a deferred panic recovery
// - Upstream cells are by-value parameters
func Generate(result ParseResult, pkgName string) (GenerateResult, error) {
	var gen GenerateResult
	gen.PackageName = pkgName

	// Collect defs and cells.
	var defsBlocks []Block
	var cellBlocks []Block
	for _, b := range result.Blocks {
		switch b.Kind {
		case BlockDefs:
			defsBlocks = append(defsBlocks, b)
		case BlockCell:
			cellBlocks = append(cellBlocks, b)
		}
	}

	// Merge defs blocks.
	var defsLines []string
	for _, d := range defsBlocks {
		defsLines = append(defsLines, d.Body)
	}
	gen.DefsCode = strings.Join(defsLines, "\n\n")

	// Track named cells for upstream references.
	var named []upstreamCell

	for i, cb := range cellBlocks {
		attrs, err := ParseInfoString(cb.Info, cb.StartPos.Line)
		if err != nil {
			return gen, err
		}

		if attrs.Name != "" {
			gc := generateNamedCell(attrs, cb, named)
			gen.Cells = append(gen.Cells, gc)
			named = append(named, upstreamCell{
				name:       attrs.Name,
				structName: gc.StructName,
				attrs:      attrs,
			})
		} else {
			gc := generateUnnamedCell(i, attrs, cb, named)
			gen.Cells = append(gen.Cells, gc)
		}
	}

	return gen, nil
}

// generateNamedCell generates the struct and function for a named cell.
func generateNamedCell(attrs CellAttrs, block Block, upstream []upstreamCell) GeneratedCell {
	structName := attrs.Name + "Cell"
	funcName := "run" + capitalize(attrs.Name)

	// Generate struct.
	var structBuf strings.Builder
	fmt.Fprintf(&structBuf, "type %s struct {\n", structName)
	fmt.Fprintf(&structBuf, "\terr error\n")
	if attrs.Exported != "" {
		for _, field := range parseExportedFields(attrs.Exported) {
			fmt.Fprintf(&structBuf, "\t%s\n", field)
		}
	}
	fmt.Fprintf(&structBuf, "}")

	// Generate function.
	var funcBuf strings.Builder
	params := buildParams(upstream)
	fmt.Fprintf(&funcBuf, "func %s(ctx *nb.Context%s) (out %s) {\n", funcName, params, structName)

	// Declare locals for exported values and err.
	fmt.Fprintf(&funcBuf, "\tvar err error\n")
	if attrs.Exported != "" {
		for _, field := range parseExportedFields(attrs.Exported) {
			parts := strings.Fields(field)
			if len(parts) >= 2 {
				fmt.Fprintf(&funcBuf, "\tvar %s\n", field)
			}
		}
	}

	// Deferred panic recovery.
	fmt.Fprintf(&funcBuf, "\tdefer func() {\n")
	fmt.Fprintf(&funcBuf, "\t\tif r := recover(); r != nil {\n")
	fmt.Fprintf(&funcBuf, "\t\t\terr = nb.PanicError(r)\n")
	fmt.Fprintf(&funcBuf, "\t\t}\n")
	fmt.Fprintf(&funcBuf, "\t\tout = %s{err: err", structName)
	if attrs.Exported != "" {
		for _, field := range parseExportedFields(attrs.Exported) {
			parts := strings.Fields(field)
			if len(parts) >= 1 {
				fmt.Fprintf(&funcBuf, ", %s: %s", parts[0], parts[0])
			}
		}
	}
	fmt.Fprintf(&funcBuf, "}\n")
	fmt.Fprintf(&funcBuf, "\t}()\n")

	// Line directive and body.
	fmt.Fprintf(&funcBuf, "//line %s:%d:1\n", "notebook.md", block.BodyPos.Line)
	fmt.Fprintf(&funcBuf, "%s\n", indentBody(block.Body))
	fmt.Fprintf(&funcBuf, "\treturn\n")
	fmt.Fprintf(&funcBuf, "}")

	return GeneratedCell{
		StructName: structName,
		FuncName:   funcName,
		StructCode: structBuf.String(),
		FuncCode:   funcBuf.String(),
	}
}

// generateUnnamedCell generates a display-only function for an unnamed cell.
func generateUnnamedCell(index int, attrs CellAttrs, block Block, upstream []upstreamCell) GeneratedCell {
	funcName := fmt.Sprintf("runCell%d", index+1)

	var funcBuf strings.Builder
	params := buildParams(upstream)
	fmt.Fprintf(&funcBuf, "func %s(ctx *nb.Context%s) {\n", funcName, params)

	// Deferred panic recovery (unnamed cells don't return a result).
	fmt.Fprintf(&funcBuf, "\tdefer func() {\n")
	fmt.Fprintf(&funcBuf, "\t\tif r := recover(); r != nil {\n")
	fmt.Fprintf(&funcBuf, "\t\t\t_ = nb.PanicError(r)\n")
	fmt.Fprintf(&funcBuf, "\t\t}\n")
	fmt.Fprintf(&funcBuf, "\t}()\n")

	// Line directive and body.
	fmt.Fprintf(&funcBuf, "//line %s:%d:1\n", "notebook.md", block.BodyPos.Line)
	fmt.Fprintf(&funcBuf, "%s\n", indentBody(block.Body))
	fmt.Fprintf(&funcBuf, "}")

	return GeneratedCell{
		StructName: "",
		FuncName:   funcName,
		FuncCode:   funcBuf.String(),
	}
}

// buildParams builds the parameter list from upstream named cells.
func buildParams(upstream []upstreamCell) string {
	if len(upstream) == 0 {
		return ""
	}
	var parts []string
	for _, u := range upstream {
		parts = append(parts, fmt.Sprintf("%s %s", u.name, u.structName))
	}
	return ", " + strings.Join(parts, ", ")
}

// parseExportedFields splits "rows []Row, total float64" into individual field declarations.
func parseExportedFields(exported string) []string {
	var fields []string
	depth := 0
	start := 0
	for i := 0; i < len(exported); i++ {
		switch exported[i] {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ',':
			if depth == 0 {
				field := strings.TrimSpace(exported[start:i])
				if field != "" {
					fields = append(fields, field)
				}
				start = i + 1
			}
		}
	}
	field := strings.TrimSpace(exported[start:])
	if field != "" {
		fields = append(fields, field)
	}
	return fields
}

// capitalize returns s with the first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// indentBody adds a tab to each line of the body.
func indentBody(body string) string {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = "\t" + l
		}
	}
	return strings.Join(lines, "\n")
}

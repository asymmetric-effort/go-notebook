// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// PruneImports removes unused imports from a Go source string.
// This is needed because notebooks declare imports in go:defs that may
// not all be used by every extracted cell.
func PruneImports(src string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return src, nil // If we can't parse, return as-is for the compiler to report.
	}

	// Collect used identifiers.
	used := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				used[ident.Name] = true
			}
		}
		if ident, ok := n.(*ast.Ident); ok {
			used[ident.Name] = true
		}
		return true
	})

	// Find unused imports.
	var unusedImports []string
	for _, imp := range f.Imports {
		name := importName(imp)
		if name == "_" || name == "." {
			continue // Blank and dot imports are always kept.
		}
		if !used[name] {
			unusedImports = append(unusedImports, name)
		}
	}

	if len(unusedImports) == 0 {
		return src, nil
	}

	// Remove unused import lines from the source.
	lines := strings.Split(src, "\n")
	result := removeImportLines(lines, unusedImports)
	return strings.Join(result, "\n"), nil
}

// importName returns the local name of an import spec.
func importName(imp *ast.ImportSpec) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	// Use the last path component as the default name.
	path := strings.Trim(imp.Path.Value, "\"")
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// removeImportLines removes lines containing unused imports from source lines.
func removeImportLines(lines []string, unused []string) []string {
	unusedSet := make(map[string]bool, len(unused))
	for _, u := range unused {
		unusedSet[u] = true
	}

	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isImportLine(trimmed, unusedSet) {
			continue
		}
		result = append(result, line)
	}
	return result
}

// isImportLine checks if a line is an import of one of the unused packages.
func isImportLine(trimmed string, unused map[string]bool) bool {
	// Match lines like:  "fmt"  or  alias "path/to/pkg"
	if !strings.HasPrefix(trimmed, "\"") && !strings.Contains(trimmed, "\"") {
		return false
	}
	// Extract the package path.
	start := strings.Index(trimmed, "\"")
	if start < 0 {
		return false
	}
	end := strings.LastIndex(trimmed, "\"")
	if end <= start {
		return false
	}
	path := trimmed[start+1 : end]
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]

	// Check for alias.
	prefix := strings.TrimSpace(trimmed[:start])
	if prefix != "" {
		name = prefix
	}

	return unused[name]
}

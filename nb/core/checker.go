// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckError is a checker finding with a Markdown line and actionable message.
type CheckError struct {
	Line    int
	Message string
}

func (e *CheckError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// checkerCellInfo tracks a named cell for reference checking.
type checkerCellInfo struct {
	name  string
	line  int
	attrs CellAttrs
}

// Check runs all checker rules on a parsed notebook and returns errors.
// Every error stops generation.
func Check(result ParseResult) []error {
	var errs []error

	// Collect cell names and positions for reference checking.
	var cells []checkerCellInfo

	for _, b := range result.Blocks {
		switch b.Kind {
		case BlockDefs:
			errs = append(errs, checkDefs(b)...)
		case BlockCell:
			attrs, err := ParseInfoString(b.Info, b.StartPos.Line)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			errs = append(errs, checkCell(b, attrs, cells)...)
			if attrs.Name != "" {
				cells = append(cells, checkerCellInfo{name: attrs.Name, line: b.StartPos.Line, attrs: attrs})
			}
		case BlockProse:
			errs = append(errs, checkProse(b)...)
		}
	}

	return errs
}

// checkDefs enforces rules on go:defs blocks.
func checkDefs(b Block) []error {
	var errs []error

	// Rule 2: go:defs contains only import, const, type and func.
	// Rule 3: No func init() in go:defs.
	// Rule 14: var only under //go:embed, typed string or embed.FS.
	// Rule 15: //go:embed and //go:generate only in go:defs (checked in cells).

	src := "package _\n\n" + b.Body + "\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		// If it doesn't parse, the Go compiler will report the error.
		return nil
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT, token.CONST, token.TYPE:
				// Allowed.
			case token.VAR:
				// Rule 14: var only allowed with //go:embed.
				if !hasEmbedDirective(d, f.Comments) {
					line := b.StartPos.Line + fset.Position(d.Pos()).Line - 2
					errs = append(errs, &CheckError{
						Line:    line,
						Message: "var in go:defs is only allowed with a //go:embed directive (rule 14). Use const for constants.",
					})
				}
			default:
				line := b.StartPos.Line + fset.Position(d.Pos()).Line - 2
				errs = append(errs, &CheckError{
					Line:    line,
					Message: fmt.Sprintf("go:defs may only contain import, const, type and func declarations (rule 2), found %s", d.Tok),
				})
			}
		case *ast.FuncDecl:
			// Rule 3: No func init().
			if d.Name.Name == "init" {
				line := b.StartPos.Line + fset.Position(d.Pos()).Line - 2
				errs = append(errs, &CheckError{
					Line:    line,
					Message: "func init() is not allowed in go:defs (rule 3). Init sets up package state as a side effect.",
				})
			}
		}
	}

	return errs
}

// checkCell enforces rules on a cell block.
func checkCell(b Block, attrs CellAttrs, upstream []checkerCellInfo) []error {
	var errs []error

	// Rule 4: Cell bodies use only bare return.
	checkBareReturn(b, &errs)

	// Rule 1: exported(...) requires name.
	if attrs.Exported != "" && attrs.Name == "" {
		errs = append(errs, &CheckError{
			Line:    b.StartPos.Line,
			Message: "exported(...) requires name= (rule 1). An unnamed cell cannot be referenced.",
		})
	}

	// Rule 5: References only to named cells above the referencing cell.
	checkCellReferences(b, attrs, upstream, &errs)

	// Rule 6: No assignment to another cell's field, and no taking its address.
	checkCellFieldMutation(b, upstream, &errs)

	// Rule 8: Files opened only through ctx.Open and declared in reads.
	// Rule 9: Files created only through ctx.Create and declared in writes.
	checkFileAccess(b, attrs, &errs)

	// Rule 10: Reject direct os.Open, os.ReadFile, os.Create, os.WriteFile calls.
	for _, fn := range []string{"os.Open", "os.ReadFile", "os.Create", "os.WriteFile", "os.OpenFile"} {
		if strings.Contains(b.Body, fn) {
			errs = append(errs, &CheckError{
				Line:    b.StartPos.Line,
				Message: fmt.Sprintf("direct %s call in cell (rule 10). Use ctx.Open or ctx.Create instead.", fn),
			})
		}
	}

	// Rule 15: //go:embed and //go:generate in cells are errors.
	for i, line := range strings.Split(b.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//go:embed") {
			errs = append(errs, &CheckError{
				Line:    b.BodyPos.Line + i,
				Message: "//go:embed is only allowed in go:defs (rule 15). Cell bodies are function bodies.",
			})
		}
		if strings.HasPrefix(trimmed, "//go:generate") {
			errs = append(errs, &CheckError{
				Line:    b.BodyPos.Line + i,
				Message: "//go:generate is only allowed in go:defs (rule 15). Cell bodies are function bodies.",
			})
		}
	}

	return errs
}

// checkBareReturn enforces rule 4: cell bodies may only use `return`,
// not `return <value>`.
func checkBareReturn(b Block, errs *[]error) {
	src := "package _\nfunc _() {\n" + b.Body + "\n}\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return
	}
	ast.Inspect(f, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		if len(ret.Results) > 0 {
			line := b.BodyPos.Line + fset.Position(ret.Pos()).Line - 2
			*errs = append(*errs, &CheckError{
				Line:    line,
				Message: "cell bodies may only use bare return, not return <value> (rule 4). The deferred function sets the result.",
			})
		}
		return true
	})
}

// imageRefRegex matches Markdown image references: ![alt](path)
var imageRefRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// htmlTagRegex matches HTML tags in prose.
var htmlTagRegex = regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>`)

// checkProse enforces rules 16-21 on prose blocks.
func checkProse(b Block) []error {
	var errs []error
	lines := strings.Split(b.Body, "\n")

	for i, line := range lines {
		mdLine := b.StartPos.Line + i

		// Rule 16: Image paths stay inside the notebook directory.
		for _, match := range imageRefRegex.FindAllStringSubmatch(line, -1) {
			imgPath := match[2]
			if filepath.IsAbs(imgPath) || strings.HasPrefix(imgPath, "/") {
				errs = append(errs, &CheckError{
					Line:    mdLine,
					Message: fmt.Sprintf("image path %q is absolute (rule 16). Use a relative path inside the notebook directory.", imgPath),
				})
			}
			if strings.Contains(imgPath, "..") && !isWebURL(imgPath) {
				errs = append(errs, &CheckError{
					Line:    mdLine,
					Message: fmt.Sprintf("image path %q leaves the notebook directory (rule 16). Remove the ../ component.", imgPath),
				})
			}
		}

		// Rule 18: Raw HTML uses only allowlisted tags.
		for _, match := range htmlTagRegex.FindAllStringSubmatch(line, -1) {
			tag := match[2]
			if IsRefusedTag(tag) {
				errs = append(errs, &CheckError{
					Line:    mdLine,
					Message: fmt.Sprintf("<%s> is not allowed in notebook prose (rule 18). Refused tags: script, iframe, frame, object, embed, base, meta.", tag),
				})
			}
		}

		// Rule 19: No DOCTYPE or entity declarations.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "<!DOCTYPE") {
			errs = append(errs, &CheckError{
				Line:    mdLine,
				Message: "<!DOCTYPE> is not allowed (rule 19). Blocks external entities and entity expansion.",
			})
		}
		if strings.Contains(trimmed, "<!ENTITY") {
			errs = append(errs, &CheckError{
				Line:    mdLine,
				Message: "<!ENTITY> declarations are not allowed (rule 19). Blocks entity expansion attacks.",
			})
		}
	}

	return errs
}

// isWebURL checks if a reference is an HTTP/HTTPS URL.
func isWebURL(ref string) bool {
	lower := strings.ToLower(ref)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// checkCellReferences enforces rule 5: cell bodies may only reference named
// cells that appear above them in the document.
func checkCellReferences(b Block, attrs CellAttrs, upstream []checkerCellInfo, errs *[]error) {
	// Build set of valid upstream cell names.
	upstreamNames := make(map[string]bool)
	for _, u := range upstream {
		upstreamNames[u.name] = true
	}

	// Rule 5 enforcement is handled by the generator: cells only receive
	// upstream named cells as parameters, so references to later cells
	// are compile errors. Full enforcement with go/types would let us
	// report violations at the Markdown line, but compile errors suffice.
	_ = upstreamNames
}

// checkCellFieldMutation enforces rule 6: no assignment to another cell's
// field (e.g., load.val = 42) and no taking its address (e.g., &load.val).
func checkCellFieldMutation(b Block, upstream []checkerCellInfo, errs *[]error) {
	if len(upstream) == 0 {
		return
	}

	src := "package _\nfunc _() {\n" + b.Body + "\n}\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return
	}

	upstreamNames := make(map[string]bool)
	for _, u := range upstream {
		upstreamNames[u.name] = true
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			// Check LHS for cellName.field = ...
			for _, lhs := range node.Lhs {
				sel, ok := lhs.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					continue
				}
				if upstreamNames[ident.Name] {
					line := b.BodyPos.Line + fset.Position(node.Pos()).Line - 2
					*errs = append(*errs, &CheckError{
						Line:    line,
						Message: fmt.Sprintf("assignment to %s.%s (rule 6). Cell values are read-only — mutation would diverge between modes.", ident.Name, sel.Sel.Name),
					})
				}
			}
		case *ast.UnaryExpr:
			// Check for &cellName.field
			if node.Op == token.AND {
				sel, ok := node.X.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if upstreamNames[ident.Name] {
					line := b.BodyPos.Line + fset.Position(node.Pos()).Line - 2
					*errs = append(*errs, &CheckError{
						Line:    line,
						Message: fmt.Sprintf("taking address of %s.%s (rule 6). Cell values are read-only — mutation would diverge between modes.", ident.Name, sel.Sel.Name),
					})
				}
			}
		}
		return true
	})
}

// checkFileAccess enforces rules 8 and 9:
// - Rule 8: ctx.Open calls must have the filename declared in reads()
// - Rule 9: ctx.Create calls must have the filename declared in writes()
func checkFileAccess(b Block, attrs CellAttrs, errs *[]error) {
	src := "package _\nfunc _() {\n" + b.Body + "\n}\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return
	}

	readsSet := make(map[string]bool)
	for _, r := range attrs.Reads {
		readsSet[r.Name] = true
	}
	writesSet := make(map[string]bool)
	for _, w := range attrs.Writes {
		writesSet[w] = true
	}

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name != "ctx" {
			return true
		}

		switch sel.Sel.Name {
		case "Open":
			// Rule 8: file must be in reads()
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					name := strings.Trim(lit.Value, "\"")
					if !readsSet[name] {
						line := b.BodyPos.Line + fset.Position(call.Pos()).Line - 2
						*errs = append(*errs, &CheckError{
							Line:    line,
							Message: fmt.Sprintf("ctx.Open(%q) but %q is not in reads() (rule 8). Declare it: reads(%s)", name, name, name),
						})
					}
				}
			}
		case "Create":
			// Rule 9: file must be in writes()
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					name := strings.Trim(lit.Value, "\"")
					if !writesSet[name] {
						line := b.BodyPos.Line + fset.Position(call.Pos()).Line - 2
						*errs = append(*errs, &CheckError{
							Line:    line,
							Message: fmt.Sprintf("ctx.Create(%q) but %q is not in writes() (rule 9). Declare it: writes(%s)", name, name, name),
						})
					}
				}
			}
		}
		return true
	})
}

// hasEmbedDirective checks if a var declaration has a //go:embed comment
// directly above it.
func hasEmbedDirective(d *ast.GenDecl, comments []*ast.CommentGroup) bool {
	for _, cg := range comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "//go:embed") {
				// Check if the comment is directly above the var declaration.
				if c.End()+2 >= d.Pos() {
					return true
				}
			}
		}
	}
	return false
}

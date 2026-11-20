// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

	// Rule 1: exported(...) requires name.
	if attrs.Exported != "" && attrs.Name == "" {
		errs = append(errs, &CheckError{
			Line:    b.StartPos.Line,
			Message: "exported(...) requires name= (rule 1). An unnamed cell cannot be referenced.",
		})
	}

	// Rule 4: Cell bodies use only bare return.
	checkBareReturn(b, &errs)

	// Rule 5: References only to named cells above.
	checkCellReferences(b, attrs, upstream, &errs)

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

// checkCellReferences is a placeholder for rule 5. The generated code enforces
// this at compile time: upstream cells are passed as function parameters, so
// referencing a cell that isn't above produces a compile error. Full AST-based
// enforcement requires go/types integration (tracked in #18).
func checkCellReferences(b Block, attrs CellAttrs, upstream []checkerCellInfo, errs *[]error) {
	// Enforced by generated code at compile time.
}

// checkBareReturn enforces rule 4: cell bodies may only use bare `return`,
// not `return x` or `return x, y`.
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

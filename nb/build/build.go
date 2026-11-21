// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package build implements gobook build, which compiles a notebook into a
// standalone report binary.
package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/document"
	"github.com/asymmetric-effort/go-notebook/nb/kernel"
	"github.com/asymmetric-effort/go-notebook/nb/render"
)

// Options configures the build.
type Options struct {
	Format     string // "md" or "html" (default "html")
	Output     string // Output file path
	Timeout    string // Cell timeout
	Passphrase string // Encryption passphrase (empty = no encryption)
}

// Run executes gobook build: parse, check, run cells, build the document
// model with actual outputs, and render to the specified format.
func Run(notebookPath string, opts Options) error {
	src, err := os.ReadFile(notebookPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", notebookPath, err)
	}

	result := core.Parse(string(src))

	// Check the notebook.
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", notebookPath, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}

	// Generate the Go package.
	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", notebookPath, err)
	}

	// Execute cells if there are any.
	var cellOutput string
	var cellFailed bool
	var doc *document.Document

	if len(gen.Cells) > 0 {
		tmpDir, err := os.MkdirTemp("", "gobook-build-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		pkgDir := filepath.Join(tmpDir, "pkg")
		if err := core.WritePackage(gen, pkgDir, notebookPath, "gobook-build"); err != nil {
			return fmt.Errorf("write package: %w", err)
		}

		moduleRoot := findModuleRoot()
		goModPath := filepath.Join(pkgDir, "go.mod")
		goModData, _ := os.ReadFile(goModPath)
		goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
		os.WriteFile(goModPath, []byte(goModStr), 0o644)

		timeout := 5 * time.Minute
		if opts.Timeout != "" {
			if d, err := time.ParseDuration(opts.Timeout); err == nil {
				timeout = d
			}
		}

		k := kernel.New()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cellResult, runErr := k.RunCell(ctx, "all", timeout,
			func(ctx context.Context) (kernel.CellResult, error) {
				return kernel.GoRun(ctx, pkgDir)
			})

		cellOutput = cellResult.Stdout
		cellFailed = cellResult.ExitCode != 0

		doc = buildDocument(result, cellOutput, cellFailed)

		if runErr != nil {
			doc.AddError("build", 0, runErr)
		} else if cellFailed {
			doc.AddError("build", 0, fmt.Errorf("cell exited with status %d", cellResult.ExitCode))
		}
	} else {
		doc = buildDocument(result, "", false)
	}

	// Render.
	format := opts.Format
	if format == "" {
		format = "html"
	}

	var output string
	switch format {
	case "md":
		output = render.Markdown(doc)
	case "html":
		output = render.HTML(doc, notebookPath)
	default:
		return fmt.Errorf("unknown format %q (want md or html)", format)
	}

	// Write output.
	outPath := opts.Output
	if outPath == "" {
		outPath = notebookPath + "." + format
	}
	if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	fmt.Printf("gobook build: %s -> %s\n", notebookPath, outPath)

	// Exit with error if cells failed (partial report was still written).
	if cellFailed {
		return fmt.Errorf("build completed with errors (partial report written to %s)", outPath)
	}
	return nil
}

// buildDocument creates a document model from parsed blocks. If cellOutput
// is provided, it's associated with the last named cell's display output.
func buildDocument(result core.ParseResult, cellOutput string, hasError bool) *document.Document {
	doc := document.NewDocument()

	// Find the last named cell for output association.
	var lastNamed string
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			if attrs.Name != "" {
				lastNamed = attrs.Name
			}
		}
	}

	for _, b := range result.Blocks {
		switch b.Kind {
		case core.BlockProse:
			doc.AddProse(b.Body)
		case core.BlockCell:
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			if attrs.Name != "" && attrs.Name == lastNamed && cellOutput != "" {
				doc.AddDisplay(attrs.Name, "text/markdown", strings.TrimRight(cellOutput, "\n"))
			}
		case core.BlockOutput:
			// Output blocks from previous runs are stripped in build mode.
		case core.BlockDefs:
			// Definitions don't appear in the report.
		}
	}
	return doc
}

// findModuleRoot finds the go-notebook module root.
func findModuleRoot() string {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "."
	}
	modFile := strings.TrimSpace(string(out))
	if modFile == "" || modFile == os.DevNull {
		return "."
	}
	return filepath.Dir(modFile)
}

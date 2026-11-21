// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package build implements gobook build, which compiles a notebook into a
// standalone report binary.
package build

import (
	"fmt"
	"os"

	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/document"
	"github.com/asymmetric-effort/go-notebook/nb/render"
)

// Options configures the build.
type Options struct {
	Format     string // "md" or "html" (default "html")
	Output     string // Output file path
	Timeout    string // Cell timeout
	Passphrase string // Encryption passphrase (empty = no encryption)
}

// Run executes gobook build: parse, check, generate the document model,
// and render to the specified format.
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

	// Build the document model from the parse result.
	doc := buildDocument(result)

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
	return nil
}

// buildDocument creates a document model from parsed blocks.
func buildDocument(result core.ParseResult) *document.Document {
	doc := document.NewDocument()
	for _, b := range result.Blocks {
		switch b.Kind {
		case core.BlockProse:
			doc.AddProse(b.Body)
		case core.BlockCell:
			// In build mode, cells would be executed here.
			// For now, record them as display output placeholders.
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			if attrs.Name != "" {
				doc.AddDisplay(attrs.Name, "text/markdown", fmt.Sprintf("(cell %s output)", attrs.Name))
			}
		case core.BlockOutput:
			// Output blocks from previous runs are stripped in build mode.
		case core.BlockDefs:
			// Definitions don't appear in the report.
		}
	}
	return doc
}

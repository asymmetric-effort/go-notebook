// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package build implements gobook build, which compiles a notebook into a
// standalone report binary.
package build

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/crypto"
	"github.com/asymmetric-effort/go-notebook/nb/document"
	"github.com/asymmetric-effort/go-notebook/nb/engine"
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

// BuildOptions configures gobook build (compile a standalone binary).
type BuildOptions struct {
	Output     string // Output binary path.
	Passphrase string // Encryption passphrase for data=embedded cells.
	NoPDF      bool   // Exclude PDF/Typst engine from binary.
	NoHTML     bool   // Exclude HTML/CommonMark from binary.
	NoMarkdown bool   // Exclude Markdown backend from binary.
	Cross      CrossCompileOptions
}

// BuildBinary compiles a notebook into a standalone report binary.
func BuildBinary(notebookPath string, opts BuildOptions) error {
	// Validate at least one format.
	if opts.NoPDF && opts.NoHTML && opts.NoMarkdown {
		return fmt.Errorf("at least one output format must be included (all excluded by --no-pdf --no-html --no-markdown)")
	}

	// Validate passphrase (rule 13).
	if err := crypto.ValidatePassphrase(opts.Passphrase); err != nil {
		return err
	}

	src, err := os.ReadFile(notebookPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", notebookPath, err)
	}

	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", notebookPath, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}

	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", notebookPath, err)
	}
	gen.ReportMode = true

	tmpDir, err := os.MkdirTemp("", "gobook-build-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, notebookPath, "gobook-report"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	// Add replace directive for local module.
	moduleRoot := findModuleRoot()
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) +
		"\nrequire github.com/spf13/cobra v1.9.1\n" +
		"\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	// Run go mod tidy to resolve dependencies.
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = pkgDir
	tidyCmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %s\n%w", out, err)
	}

	// Compile the binary.
	output := opts.Output
	if output == "" {
		output = strings.TrimSuffix(filepath.Base(notebookPath), filepath.Ext(notebookPath))
	}

	buildArgs := []string{"build", "-o", output, "."}
	cmd := exec.Command("go", buildArgs...)
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")

	// Cross-compilation.
	if opts.Cross.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+opts.Cross.GOOS)
	}
	if opts.Cross.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+opts.Cross.GOARCH)
	}
	if opts.Cross.CGOEnabled != "" {
		cmd.Env = append(cmd.Env, "CGO_ENABLED="+opts.Cross.CGOEnabled)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %s\n%w", out, err)
	}

	fmt.Printf("gobook build: %s -> %s\n", notebookPath, output)
	return nil
}

// Run executes gobook build: parse, check, run cells, build the document
// model with actual outputs, and render to the specified format.
func Run(notebookPath string, opts Options) error {
	// Validate passphrase length (rule 13).
	if err := crypto.ValidatePassphrase(opts.Passphrase); err != nil {
		return err
	}

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

		// If there are embedded cells, run in embed mode first to capture data.
		if hasEmbeddedCells(gen) {
			embedResult, _ := k.RunCell(ctx, "embed", timeout,
				func(ctx context.Context) (kernel.CellResult, error) {
					return kernel.GoRunWithEnv(ctx, pkgDir, []string{"GOBOOK_EMBED_MODE=1"})
				})
			gen.EmbeddedCells = parseEmbedOutput(embedResult.Stdout)
		}

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

	// Render to the requested format.
	outPath := opts.Output
	switch format {
	case "md":
		output := render.Markdown(doc)
		if outPath == "" {
			outPath = notebookPath + ".md"
		}
		if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	case "html":
		output := render.HTML(doc, notebookPath)
		if outPath == "" {
			outPath = notebookPath + ".html"
		}
		if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	case "typst":
		output := render.Typst(doc)
		if outPath == "" {
			outPath = notebookPath + ".typ"
		}
		if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	case "pdf":
		typstSrc := render.Typst(doc)
		if outPath == "" {
			outPath = notebookPath + ".pdf"
		}
		ctx := context.Background()
		eng, err := engine.New(ctx)
		if err != nil {
			return fmt.Errorf("start typst engine: %w", err)
		}
		defer eng.Close(ctx)
		pdfBytes, err := eng.Compile(ctx, typstSrc, "pdf")
		if err != nil {
			return fmt.Errorf("compile pdf: %w", err)
		}
		if err := os.WriteFile(outPath, pdfBytes, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	default:
		return fmt.Errorf("unknown format %q (want md, html, typst, or pdf)", format)
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
// hasEmbeddedCells returns true if any cell has data=embedded.
func hasEmbeddedCells(gen core.GenerateResult) bool {
	for _, cell := range gen.Cells {
		if cell.DataMode == core.DataEmbedded {
			return true
		}
	}
	return false
}

// parseEmbedOutput extracts GOBOOK_EMBED:name:base64 lines from stdout.
func parseEmbedOutput(stdout string) []core.EmbeddedCellData {
	var embedded []core.EmbeddedCellData
	for _, line := range strings.Split(stdout, "\n") {
		if !strings.HasPrefix(line, "GOBOOK_EMBED:") {
			continue
		}
		rest := strings.TrimPrefix(line, "GOBOOK_EMBED:")
		idx := strings.Index(rest, ":")
		if idx < 0 {
			continue
		}
		name := rest[:idx]
		b64Data := rest[idx+1:]
		data, err := base64.StdEncoding.DecodeString(b64Data)
		if err != nil {
			continue
		}
		embedded = append(embedded, core.EmbeddedCellData{
			CellName:      name,
			EncryptedData: data, // Serialized cell data (would be encrypted in production).
		})
	}
	return embedded
}

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

// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package commands implements the gobook CLI commands.
package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/build"
	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/crypto"
	"github.com/asymmetric-effort/go-notebook/nb/identity"
	"github.com/asymmetric-effort/go-notebook/nb/kernel"
	"github.com/asymmetric-effort/go-notebook/nb/lsp"
	"github.com/asymmetric-effort/go-notebook/nb/server"
	"github.com/spf13/cobra"
)

// version is set at build time via ldflags, or resolved from module
// build info when installed via go install.
var version = resolveVersion()

func resolveVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		// Accept tagged versions (v1.2.3) but not pseudo-versions
		// (v0.0.9-0.20261127...) or development markers ((devel)).
		if strings.HasPrefix(v, "v") && !strings.Contains(v, "-0.") {
			return v
		}
	}
	return "dev"
}

var rootCmd = &cobra.Command{
	Use:           "gobook",
	Short:         "Go notebook compiler and runner",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var buildOutput string
var buildPassphrase bool
var buildPassphraseFile string
var buildSignKey string
var buildRecipients []string
var buildNoPDF bool
var buildNoHTML bool
var buildNoMarkdown bool
var buildOS string
var buildArch string

var buildCmd = &cobra.Command{
	Use:   "build <notebook.md>",
	Short: "Compile a notebook into a standalone report binary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		passphrase, err := resolvePassphrase(buildPassphrase, buildPassphraseFile)
		if err != nil {
			return err
		}
		opts := build.BuildOptions{
			Output:     buildOutput,
			Passphrase: passphrase,
			Recipients: buildRecipients,
			NoPDF:      buildNoPDF,
			NoHTML:     buildNoHTML,
			NoMarkdown: buildNoMarkdown,
		}
		if buildOS != "" || buildArch != "" {
			opts.Cross.GOOS = buildOS
			opts.Cross.GOARCH = buildArch
			opts.Cross.CGOEnabled = "0"
		}
		err = build.BuildBinary(args[0], opts)
		if err != nil {
			return err
		}
		// Sign the output if --sign-key is provided.
		if buildSignKey != "" {
			outPath := buildOutput
			if outPath == "" {
				outPath = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
			}
			if err := identity.SignFile(outPath, buildSignKey); err != nil {
				return fmt.Errorf("sign: %w", err)
			}
			fmt.Fprintf(os.Stderr, "gobook build: signed %s -> %s.sig\n", outPath, outPath)
		}
		return nil
	},
}

// Variables for gobook run (renders directly via temp binary).
var runFormat string
var runOutput string
var runTimeout string

var checkCmd = &cobra.Command{
	Use:   "check <notebook.md>",
	Short: "Check a notebook for errors",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdCheck(args[0])
	},
}

var fmtCmd = &cobra.Command{
	Use:   "fmt <notebook.md>",
	Short: "Format a notebook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdFmt(args[0])
	},
}

var genOutDir string

var genCmd = &cobra.Command{
	Use:   "gen <notebook.md>",
	Short: "Generate Go package from a notebook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdGen(args[0], genOutDir)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate <notebook.md>",
	Short: "Generate Go package and run //go:generate directives",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notebookPath := args[0]
		// First, generate the package (same as gen).
		if err := cmdGen(notebookPath, genOutDir); err != nil {
			return err
		}
		// Then, find and run //go:generate directives from go:defs blocks.
		return runGoGenerateDirectives(notebookPath)
	},
}

var runCmd = &cobra.Command{
	Use:   "run <notebook.md> [-- flags for the report binary]",
	Short: "Build and run a notebook, producing a report file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notebookPath := args[0]

		// Build to a temp directory.
		tmpDir, err := os.MkdirTemp("", "gobook-run-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		binPath := filepath.Join(tmpDir, "report")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}

		err = build.BuildBinary(notebookPath, build.BuildOptions{
			Output: binPath,
		})
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		// Execute the temp binary, passing through flags after --.
		binArgs := cmd.ArgsLenAtDash()
		var passthrough []string
		if binArgs >= 0 && binArgs < len(args) {
			passthrough = args[binArgs:]
		} else {
			// No -- separator: use default flags.
			passthrough = []string{"--format", "html"}
			if runOutput != "" {
				passthrough = append(passthrough, "-o", runOutput)
			}
			if runFormat != "" {
				passthrough[1] = runFormat
			}
		}

		execCmd := exec.Command(binPath, passthrough...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin
		return execCmd.Run()
	},
}

var servePort int
var serveHost string

var serveCmd = &cobra.Command{
	Use:   "serve <notebook.md>",
	Short: "Start the notebook editor server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("notebook not found: %s", absPath)
		}
		addr := fmt.Sprintf("%s:%d", serveHost, servePort)
		srv := server.New(absPath, addr)
		return srv.ListenAndServe()
	},
}

var keygenOutput string

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a post-quantum identity (ML-DSA-65)",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := keygenOutput
		if name == "" {
			name = "gobook-key"
		}
		id, err := identity.Generate()
		if err != nil {
			return err
		}
		if err := id.WriteKeys(name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "gobook keygen: wrote %s.pub and %s.key\n", name, name)
		fmt.Fprintf(os.Stderr, "fingerprint: %s\n", id.Fingerprint())
		return nil
	},
}

var goplsPath string

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the LSP proxy to gopls",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy := lsp.New(goplsPath)
		return proxy.Start()
	},
}

var verifySignature string
var verifySigner string

var verifyCmd = &cobra.Command{
	Use:   "verify <binary>",
	Short: "Verify a detached ML-DSA-65 signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := identity.VerifyFile(args[0], verifySignature, verifySigner); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "gobook verify: %s signature valid\n", args[0])
		return nil
	},
}

var vetCmd = &cobra.Command{
	Use:   "vet <notebook.md>",
	Short: "Run go vet on a notebook's generated code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdVet(args[0])
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output binary path")
	buildCmd.Flags().BoolVar(&buildPassphrase, "passphrase", false, "prompt for encryption passphrase")
	buildCmd.Flags().StringVar(&buildPassphraseFile, "passphrase-file", "", "read passphrase from file")
	buildCmd.MarkFlagsMutuallyExclusive("passphrase", "passphrase-file")
	buildCmd.Flags().StringVar(&buildSignKey, "sign-key", "", "sign output with ML-DSA-65 private key")
	buildCmd.Flags().StringArrayVar(&buildRecipients, "recipient", nil, "encrypt for recipient public key (repeatable)")
	buildCmd.MarkFlagsMutuallyExclusive("passphrase", "recipient")
	buildCmd.MarkFlagsMutuallyExclusive("passphrase-file", "recipient")
	buildCmd.Flags().BoolVar(&buildNoPDF, "no-pdf", false, "exclude PDF/Typst engine from binary")
	buildCmd.Flags().BoolVar(&buildNoHTML, "no-html", false, "exclude HTML/CommonMark from binary")
	buildCmd.Flags().BoolVar(&buildNoMarkdown, "no-markdown", false, "exclude Markdown backend from binary")
	buildCmd.Flags().StringVar(&buildOS, "os", "", "target OS for cross-compilation")
	buildCmd.Flags().StringVar(&buildArch, "arch", "", "target architecture for cross-compilation")

	runCmd.Flags().StringVar(&runFormat, "format", "html", "output format (html, pdf, md)")
	runCmd.Flags().StringVarP(&runOutput, "output", "o", "", "output file path")

	genCmd.Flags().StringVar(&genOutDir, "out", "gen", "output directory")

	serveCmd.Flags().IntVar(&servePort, "port", 8080, "HTTP port")
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "bind address")

	keygenCmd.Flags().StringVarP(&keygenOutput, "output", "o", "", "output key file base name")

	lspCmd.Flags().StringVar(&goplsPath, "gopls", "gopls", "path to gopls binary")

	verifyCmd.Flags().StringVar(&verifySignature, "signature", "", "detached signature file (.sig)")
	verifyCmd.Flags().StringVar(&verifySigner, "signer", "", "signer's public key file (.pub)")
	_ = verifyCmd.MarkFlagRequired("signature")
	_ = verifyCmd.MarkFlagRequired("signer")

	rootCmd.AddCommand(buildCmd, checkCmd, fmtCmd, genCmd, generateCmd, keygenCmd, lspCmd, runCmd, serveCmd, verifyCmd, vetCmd)
}

// Execute runs the gobook CLI.
func Execute() error {
	return rootCmd.Execute()
}

// resolvePassphrase handles the --passphrase and --passphrase-file flags.
// Returns empty string if neither flag is set (no encryption needed).
func resolvePassphrase(prompt bool, file string) (string, error) {
	if !prompt && file == "" {
		return "", nil
	}
	var passphrase string
	var err error
	if file != "" {
		passphrase, err = crypto.ReadPassphraseFromFile(file)
		if err != nil {
			return "", err
		}
	} else {
		passphrase, err = crypto.PromptPassphrase(true) // confirm on build
		if err != nil {
			return "", err
		}
	}
	if err := crypto.ValidatePassphrase(passphrase); err != nil {
		return "", fmt.Errorf("passphrase rejected: %w", err)
	}
	return passphrase, nil
}

func cmdBuild(path string, opts build.Options) error {
	return build.Run(path, opts)
}

func cmdCheck(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}
	return nil
}

func cmdFmt(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	formatted, err := core.FormatNotebook(string(src))
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	if formatted != string(src) {
		if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func cmdGen(path string, outDir string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	result := core.Parse(string(src))
	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}

	if err := core.WritePackage(gen, outDir, path, "gobook-gen"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	fmt.Printf("gobook gen: wrote %d cells to %s/\n", len(gen.Cells), outDir)
	return nil
}

// runGoGenerateDirectives scans go:defs blocks for //go:generate directives
// and executes them from the notebook's directory.
func runGoGenerateDirectives(notebookPath string) error {
	src, err := os.ReadFile(notebookPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", notebookPath, err)
	}

	result := core.Parse(string(src))
	notebookDir := filepath.Dir(notebookPath)
	if notebookDir == "" {
		notebookDir = "."
	}

	var directives []string
	for _, b := range result.Blocks {
		if b.Kind != core.BlockDefs {
			continue
		}
		for _, line := range strings.Split(b.Body, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//go:generate ") {
				directives = append(directives, strings.TrimPrefix(trimmed, "//go:generate "))
			}
		}
	}

	if len(directives) == 0 {
		return nil
	}

	for _, directive := range directives {
		parts := strings.Fields(directive)
		if len(parts) == 0 {
			continue
		}
		fmt.Fprintf(os.Stderr, "gobook generate: running %s\n", directive)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = notebookDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go:generate %q: %w", directive, err)
		}
	}

	return nil
}

func cmdRun(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check errors", len(errs))
	}

	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}

	tmpDir, err := os.MkdirTemp("", "gobook-run-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, path, "gobook-run"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	moduleRoot := findModuleRootFromPath(path)
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	k := kernel.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cellResult, err := k.RunCell(ctx, "all", 5*time.Minute,
		func(ctx context.Context) (kernel.CellResult, error) {
			return kernel.GoRun(ctx, pkgDir)
		})
	if err != nil {
		return fmt.Errorf("run %s: %w", path, err)
	}

	if cellResult.Stdout != "" {
		fmt.Print(cellResult.Stdout)
	}
	if cellResult.Stderr != "" {
		fmt.Fprint(os.Stderr, cellResult.Stderr)
	}
	if cellResult.ExitCode != 0 {
		return fmt.Errorf("cell exited with status %d", cellResult.ExitCode)
	}

	if cellResult.Stdout != "" {
		outputs := make(map[string]string)
		lastNamed := lastNamedCell(result)
		if lastNamed != "" {
			outputs[lastNamed] = strings.TrimRight(cellResult.Stdout, "\n")
		}
		if len(outputs) > 0 {
			updated := core.WriteOutputBlocks(string(src), outputs)
			if updated != string(src) {
				if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
					return fmt.Errorf("write output blocks: %w", err)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "gobook run: %s completed successfully\n", path)
	return nil
}

func cmdVet(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, e)
		}
		return fmt.Errorf("%d check findings", len(errs))
	}

	gen, err := core.Generate(result, "main")
	if err != nil {
		return fmt.Errorf("generate %s: %w", path, err)
	}

	tmpDir, err := os.MkdirTemp("", "gobook-vet-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, path, "gobook-vet"); err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	moduleRoot := findModuleRootFromPath(path)
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))
		return fmt.Errorf("go vet found issues")
	}

	fmt.Fprintf(os.Stderr, "gobook vet: %s passed\n", path)
	return nil
}

func findModuleRootFromPath(notebookPath string) string {
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

func lastNamedCell(result core.ParseResult) string {
	var last string
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			attrs, err := core.ParseInfoString(b.Info, b.StartPos.Line)
			if err == nil && attrs.Name != "" {
				last = attrs.Name
			}
		}
	}
	return last
}

func countCells(result core.ParseResult) int {
	n := 0
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			n++
		}
	}
	return n
}

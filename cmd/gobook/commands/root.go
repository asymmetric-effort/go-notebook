// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package commands implements the gobook CLI commands.
package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

var rootCmd = &cobra.Command{
	Use:           "gobook",
	Short:         "Go notebook compiler and runner",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var buildFormat string
var buildOutput string
var buildTimeout string
var buildPassphrase bool
var buildPassphraseFile string
var buildSignKey string
var buildRecipients []string

var buildCmd = &cobra.Command{
	Use:   "build <notebook.md>",
	Short: "Build a notebook into a report",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		passphrase, err := resolvePassphrase(buildPassphrase, buildPassphraseFile)
		if err != nil {
			return err
		}
		err = cmdBuild(args[0], build.Options{
			Format:     buildFormat,
			Output:     buildOutput,
			Timeout:    buildTimeout,
			Passphrase: passphrase,
		})
		if err != nil {
			return err
		}
		// Sign the output if --sign-key is provided.
		if buildSignKey != "" {
			outPath := buildOutput
			if outPath == "" {
				format := buildFormat
				if format == "" {
					format = "html"
				}
				outPath = args[0] + "." + format
			}
			if err := identity.SignFile(outPath, buildSignKey); err != nil {
				return fmt.Errorf("sign: %w", err)
			}
			fmt.Fprintf(os.Stderr, "gobook build: signed %s -> %s.sig\n", outPath, outPath)
		}
		return nil
	},
}

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
	Use:     "gen <notebook.md>",
	Aliases: []string{"generate"},
	Short:   "Generate Go package from a notebook",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdGen(args[0], genOutDir)
	},
}

var runCmd = &cobra.Command{
	Use:   "run <notebook.md>",
	Short: "Run all cells in a notebook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdRun(args[0])
	},
}

var serveAddr string

var serveCmd = &cobra.Command{
	Use:   "serve <notebook.md>",
	Short: "Start the notebook editor server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := server.New(args[0], serveAddr)
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

var verifyCmd = &cobra.Command{
	Use:   "verify <file> <signature> <public-key>",
	Short: "Verify a detached ML-DSA-65 signature",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := identity.VerifyFile(args[0], args[1], args[2]); err != nil {
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
	buildCmd.Flags().StringVar(&buildFormat, "format", "", "output format (md, html, typst, pdf)")
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output file path")
	buildCmd.Flags().StringVar(&buildTimeout, "cell-timeout", "", "per-cell timeout")
	buildCmd.Flags().BoolVar(&buildPassphrase, "passphrase", false, "prompt for encryption passphrase")
	buildCmd.Flags().StringVar(&buildPassphraseFile, "passphrase-file", "", "read passphrase from file")
	buildCmd.MarkFlagsMutuallyExclusive("passphrase", "passphrase-file")
	buildCmd.Flags().StringVar(&buildSignKey, "sign-key", "", "sign output with ML-DSA-65 private key")
	buildCmd.Flags().StringArrayVar(&buildRecipients, "recipient", nil, "encrypt for recipient public key (repeatable)")

	genCmd.Flags().StringVar(&genOutDir, "out", "gen", "output directory")

	serveCmd.Flags().StringVar(&serveAddr, "addr", "localhost:8080", "listen address")

	keygenCmd.Flags().StringVarP(&keygenOutput, "output", "o", "", "output key file base name")

	lspCmd.Flags().StringVar(&goplsPath, "gopls", "gopls", "path to gopls binary")

	rootCmd.AddCommand(buildCmd, checkCmd, fmtCmd, genCmd, keygenCmd, lspCmd, runCmd, serveCmd, verifyCmd, vetCmd)
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

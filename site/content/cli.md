# CLI Reference

The `gobook` command-line tool provides all operations for working with
go-notebooks.

```bash
gobook [command] [flags]
```

Use `gobook --version` to print the version string.

## gobook build

Compile a notebook into a standalone report binary.

```bash
gobook build <notebook.md> [flags]
```

Requires exactly one positional argument: the notebook file.

| Flag | Description |
| ---- | ----------- |
| `-o, --output` | Output binary path (default: notebook name without .md) |
| `--passphrase` | Prompt for encryption passphrase |
| `--passphrase-file` | Read passphrase from file (mutually exclusive with `--passphrase`) |
| `--sign-key` | Sign output with ML-DSA-65 private key |
| `--recipient` | Encrypt for recipient public key (repeatable) |
| `--no-pdf` | Exclude PDF/Typst engine from binary |
| `--no-html` | Exclude HTML/CommonMark from binary |
| `--no-markdown` | Exclude Markdown backend from binary |
| `--os` | Target OS for cross-compilation |
| `--arch` | Target architecture for cross-compilation |

The `--passphrase` and `--recipient` flags are mutually exclusive, as are
`--passphrase-file` and `--recipient`.

Examples:

```bash
# Basic build
gobook build report.md -o report

# Cross-compile for Linux ARM64
gobook build report.md -o report --os linux --arch arm64

# Build PDF-only binary (smaller size)
gobook build report.md -o report --no-html --no-markdown

# Encrypted build (interactive prompt)
gobook build report.md -o report --passphrase

# Encrypted build (passphrase from file)
gobook build report.md -o report --passphrase-file secret.txt

# Sign the output binary
gobook build report.md -o report --sign-key mykey.key
```

## gobook run

Build a temporary binary, execute it, and clean up.

```bash
gobook run <notebook.md> [-- flags-for-binary]
```

Requires at least one positional argument: the notebook file.

| Flag | Description |
| ---- | ----------- |
| `--format` | Output format: html, pdf, or md (default: html) |
| `-o, --output` | Output file path |

When no `--` separator is present, `--format` and `-o` are used as defaults
for the generated binary. When `--` is present, everything after it is passed
through to the binary verbatim (and `--format`/`-o` are ignored).

```bash
# Use built-in flags (no -- separator)
gobook run report.md --format pdf -o report.pdf

# Pass flags through to the binary
gobook run report.md -- --format html -o report.html
```

## gobook serve

Launch the browser-based notebook editor.

```bash
gobook serve <notebook.md> [flags]
```

Requires exactly one positional argument: the notebook file.

| Flag | Description |
| ---- | ----------- |
| `--port` | HTTP port (default: 8080) |
| `--host` | Bind address (default: localhost) |

## gobook check

Check a notebook for errors by running checker rules.

```bash
gobook check <notebook.md>
```

Requires exactly one positional argument. Reports any check errors to stderr
and exits with a non-zero status if problems are found.

## gobook fmt

Format a notebook in place.

```bash
gobook fmt <notebook.md>
```

Requires exactly one positional argument. Rewrites the file only if the
formatted output differs from the original.

## gobook gen

Generate a Go package from a notebook.

```bash
gobook gen <notebook.md> [flags]
```

Requires exactly one positional argument.

| Flag | Description |
| ---- | ----------- |
| `--out` | Output directory (default: gen) |

## gobook generate

Generate a Go package and run `//go:generate` directives from `go:defs` blocks.

```bash
gobook generate <notebook.md>
```

Requires exactly one positional argument. This command first runs `gen`
(producing the Go package), then scans `go:defs` blocks for `//go:generate`
directives and executes them from the notebook's directory.

## gobook vet

Run `go vet` on a notebook's generated code.

```bash
gobook vet <notebook.md>
```

Requires exactly one positional argument. Generates the Go package to a
temporary directory, then runs `go vet ./...` on it.

## gobook keygen

Generate a post-quantum identity key pair (ML-DSA-65).

```bash
gobook keygen [flags]
```

Takes no positional arguments.

| Flag | Description |
| ---- | ----------- |
| `-o, --output` | Output key file base name (default: gobook-key) |

Writes `<name>.pub` and `<name>.key` and prints the fingerprint.

## gobook verify

Verify a detached ML-DSA-65 signature.

```bash
gobook verify <binary> [flags]
```

Requires exactly one positional argument: the file to verify.

| Flag | Required | Description |
| ---- | -------- | ----------- |
| `--signature` | Yes | Detached signature file (.sig) |
| `--signer` | Yes | Signer's public key file (.pub) |

## gobook lsp

Start the LSP proxy to gopls for editor integration.

```bash
gobook lsp [flags]
```

Takes no positional arguments.

| Flag | Description |
| ---- | ----------- |
| `--gopls` | Path to gopls binary (default: gopls) |

## Report Binary Flags

The binary produced by `gobook build` accepts these flags:

| Flag | Description |
| ---- | ----------- |
| `--format` | Output format: `html`, `pdf`, or `markdown` |
| `-o, --output` | Output file path |
| `--cell-timeout` | Per-cell execution timeout |
| `--passphrase` | Decrypt with passphrase (interactive prompt) |
| `--key-file` | Decrypt with private key file |

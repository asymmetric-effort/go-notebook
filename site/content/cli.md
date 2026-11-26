# CLI Reference

The `gobook` command-line tool provides all operations for working with
go-notebooks.

## gobook build

Compile a notebook into a standalone report binary.

```bash
gobook build <notebook.md> [flags]
```

| Flag | Description |
| ---- | ----------- |
| `-o, --output` | Output binary path (default: notebook name without .md) |
| `--os` | Target operating system for cross-compilation |
| `--arch` | Target architecture for cross-compilation |
| `--no-pdf` | Exclude PDF rendering from the binary |
| `--no-html` | Exclude HTML rendering from the binary |
| `--no-markdown` | Exclude Markdown rendering from the binary |
| `--recipient` | Encrypt for recipient (repeatable for multiple recipients) |
| `--sign-key` | Sign the report with this key |
| `--passphrase` | Encrypt with a passphrase (interactive prompt) |

Examples:

```bash
# Basic build
gobook build report.md -o report

# Cross-compile for Linux ARM64
gobook build report.md -o report-arm --os linux --arch arm64

# Build PDF-only binary (smaller size)
gobook build report.md -o report --no-html --no-markdown

# Encrypted build
gobook build report.md -o report --passphrase
```

## gobook run

Build a temporary binary, execute it, and clean up.

```bash
gobook run <notebook.md> [-- flags-for-binary]
```

Flags after `--` are passed to the generated binary:

```bash
# Render to HTML
gobook run report.md -- --format html -o report.html

# Render to PDF
gobook run report.md -- --format pdf -o report.pdf
```

## gobook serve

Launch the browser-based notebook editor.

```bash
gobook serve <notebook.md> [flags]
```

| Flag | Description |
| ---- | ----------- |
| `--port` | HTTP port (default: 8080) |
| `--host` | Bind address (default: localhost) |

## gobook keygen

Generate an encryption key pair.

```bash
gobook keygen [flags]
```

| Flag | Description |
| ---- | ----------- |
| `-o, --output` | Output path prefix (creates .pub and .key files) |

## gobook verify

Verify a report's digital signature.

```bash
gobook verify [flags]
```

| Flag | Description |
| ---- | ----------- |
| `--signature` | Path to the signature file |
| `--signer` | Path to the signer's public key |

## gobook generate

Run Go generate directives in the project.

```bash
gobook generate
```

## gobook version

Print the gobook version.

```bash
gobook version
```

## Report Binary Flags

The binary produced by `gobook build` accepts these flags:

| Flag | Description |
| ---- | ----------- |
| `--format` | Output format: `html`, `pdf`, or `markdown` |
| `-o, --output` | Output file path |
| `--cell-timeout` | Per-cell execution timeout |
| `--passphrase` | Decrypt with passphrase (interactive prompt) |
| `--key-file` | Decrypt with private key file |

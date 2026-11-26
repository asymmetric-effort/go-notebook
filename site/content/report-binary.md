# Report Binaries

One of go-notebook's key features is compiling notebooks into standalone
report binaries. These are self-contained executables that can render the
report on any platform without requiring Go or gobook to be installed.

## How It Works

When you run `gobook build`, the tool:

1. **Parses** the notebook Markdown into cells
2. **Generates** Go source code with the notebook content embedded
3. **Compiles** the source into a native binary using `go build`
4. **Optionally encrypts** the payload before embedding

The resulting binary contains everything needed to render the report:
the compiled cell code, the Typst WASM engine (for PDF), and all static
assets.

## Building a Report

```bash
gobook build analysis.md -o analysis
```

This produces an `analysis` binary (or `analysis.exe` on Windows).

## Running a Report

The report binary has its own CLI powered by Cobra:

```bash
# Generate HTML output
./analysis --format html -o analysis.html

# Generate PDF output
./analysis --format pdf -o analysis.pdf

# Generate Markdown output
./analysis --format markdown -o analysis.md
```

## Cross-Compilation

Build reports for different platforms:

```bash
# Build for Linux ARM64
gobook build report.md -o report --os linux --arch arm64

# Build for Windows
gobook build report.md -o report.exe --os windows --arch amd64

# Build for macOS
gobook build report.md -o report --os darwin --arch arm64
```

## Format Exclusion

Reduce binary size by excluding formats you don't need:

```bash
# PDF-only report (no HTML or Markdown rendering)
gobook build report.md -o report --no-html --no-markdown

# HTML-only report
gobook build report.md -o report --no-pdf --no-markdown
```

## Encrypted Reports

Reports can be encrypted so only authorized users can render them.
See the [Encryption](#/encryption) page for details.

### Passphrase Encryption

```bash
# Build with passphrase encryption
gobook build report.md -o report --passphrase

# Render (prompts for passphrase)
./report --format pdf -o report.pdf --passphrase
```

### Recipient Encryption

```bash
# Build for specific recipients
gobook build report.md -o report --recipient alice.pub

# Render with private key
./report --format pdf -o report.pdf --key-file alice.key
```

## Signed Reports

Add a digital signature to verify report authenticity:

```bash
# Build and sign
gobook build report.md -o report --sign-key author.sign

# Verify signature
gobook verify --signature report.sig --signer author.sign.pub
```

## Distribution

Report binaries are fully self-contained. You can distribute them by:

- Email attachment
- File share
- Package manager
- Container image
- Any other file transfer mechanism

Recipients only need to run the binary — no Go toolchain, no gobook, no
runtime dependencies.

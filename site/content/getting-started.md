# Getting Started

## Installation

Install `gobook` using the Go toolchain:

```bash
go install github.com/asymmetric-effort/go-notebook/cmd/gobook@latest
```

Verify the installation:

```bash
gobook version
```

## Prerequisites

- **Go 1.27 or later** — the gobook binary is written in Go
- **A text editor** — any editor that supports Markdown will work

No other system dependencies are required. The PDF renderer uses WebAssembly
(wazero) and needs no C compiler or external libraries.

## Your First Notebook

Create a file called `hello.md` with the following structure:

- A Markdown heading and prose
- A `go:defs` fenced code block with shared imports
- A `go:notebook` fenced code block with executable code

For example, the `go:defs` block would contain:

```go
// Shared definitions available to all cells.
import "fmt"
```

And the `go:notebook` block would contain:

```go
// Cells run in order and can produce output.
fmt.Println("Hello from go-notebook!")
```

## Running the Notebook

Use `gobook run` for quick iteration:

```bash
gobook run hello.md -- --format html -o hello.html
```

This builds a temporary binary from your notebook, executes it, and cleans up.
The `--` separator passes flags to the generated binary.

## Building a Report Binary

Use `gobook build` to create a distributable executable:

```bash
gobook build hello.md -o hello
```

The resulting binary is self-contained. Send it to anyone and they can render
the report without installing Go or gobook:

```bash
./hello --format pdf -o hello.pdf
./hello --format html -o hello.html
./hello --format markdown -o hello-rendered.md
```

## Editor Mode

Launch the browser-based notebook editor:

```bash
gobook serve hello.md
```

This starts an LSP-powered editor with Go autocompletion, diagnostics, and
hover documentation. See the [Editor](#/editor) page for details.

## Next Steps

- **[User Guide](#/user-guide)** — Learn about cell types, attributes, and data modes
- **[Charts](#/charts)** — Add line, bar, pie, and other charts to your notebooks
- **[Encryption](#/encryption)** — Protect reports with post-quantum encryption
- **[CLI Reference](#/cli)** — Full command documentation

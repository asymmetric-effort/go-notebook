# Getting Started

## Prerequisites

- **Go 1.27 or later** — required to compile notebooks into binaries
- **A text editor** — any editor that supports Markdown will work

No other system dependencies are required. The PDF renderer uses WebAssembly
(wazero) and needs no C compiler or external libraries.

## Installation

Install `gobook` using the Go toolchain:

```bash
go install github.com/asymmetric-effort/go-notebook/cmd/gobook@latest
```

Verify the installation:

```bash
gobook version
```

## Hello World Example

Download the complete example: [hello.md](/examples/hello.md)

Or create a file called `hello.md` with the following content:

```markdown
# Hello World

This is a go-notebook. The prose sections are standard CommonMark
Markdown and appear as-is in the rendered output.
```

The file starts with ordinary Markdown. Next, add a `go:defs` block
for shared imports. This block is not executed directly — it provides
definitions available to all cells in the notebook:

```go
import "fmt"
import "runtime"
import "time"
```

The info string on the code fence must be `go:defs` (not plain `go`).

Now add executable cells using `go:notebook` fences. Each cell runs
in order and its printed output appears in the rendered report:

```go
fmt.Println("Hello from go-notebook!")
fmt.Printf("Built with Go %s on %s/%s\n",
    runtime.Version(), runtime.GOOS, runtime.GOARCH)
```

Cells can be named using attributes for clearer error messages:

```go
fmt.Printf("Report generated at: %s\n",
    time.Now().Format(time.RFC1123))
```

The info string for these cells is `go:notebook` (optionally with
`name=cellname`). You can add as many cells as you need, with
Markdown prose between them.

## Building and Running

### Quick run (build, execute, clean up)

```bash
gobook run hello.md -- --format html -o hello.html
```

The `--` separator passes flags to the generated binary. Expected
terminal output:

```text
Hello from go-notebook!
Built with Go go1.27.4 on linux/amd64
Report generated at: Thu, 27 Nov 2026 12:00:00 UTC
```

The file `hello.html` now contains a self-contained HTML report
with the Markdown prose and cell output rendered together.

### Build a standalone binary

```bash
gobook build hello.md -o hello
```

This produces a `hello` binary (or `hello.exe` on Windows). The
binary is self-contained — send it to anyone and they can render
the report without installing Go or gobook:

```bash
./hello --format html -o hello.html
./hello --format pdf -o hello.pdf
./hello --format markdown -o hello-rendered.md
```

### Output formats

| Format | Flag | Description |
| ------ | ---- | ----------- |
| HTML | `--format html` | Self-contained HTML with inline styles |
| PDF | `--format pdf` | PDF rendered via Typst (zero cgo) |
| Markdown | `--format markdown` | Markdown with output blocks filled in |

## Editor Mode

Launch the browser-based notebook editor for interactive development:

```bash
gobook serve hello.md
```

This starts an LSP-powered editor with Go autocompletion, diagnostics,
and hover documentation. See the [Editor](#/editor) page for details.

## Next Steps

- **[User Guide](#/user-guide)** — Cell types, attributes, and data modes
- **[Charts](#/charts)** — Add line, bar, pie, and other charts
- **[Encryption](#/encryption)** — Protect reports with post-quantum encryption
- **[CLI Reference](#/cli)** — Full command documentation

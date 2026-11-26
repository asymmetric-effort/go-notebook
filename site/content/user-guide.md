# User Guide

## Notebook Structure

A go-notebook is a standard CommonMark Markdown file (`.md`) with special
fenced code blocks. The prose between code blocks is rendered as-is in the
output.

There are three types of code fences:

### go:defs — Shared Definitions

A `go:defs` block contains imports, constants, types, and helper functions
shared across all cells:

```go
import "fmt"
import "math"

const pi = math.Pi
```

Definitions are available to all subsequent cells. Use `go:defs` for imports,
constants, types, and helper functions that multiple cells need.

### go:notebook — Executable Cells

A `go:notebook` block contains executable code that produces output:

```go
fmt.Printf("Pi is approximately %.4f\n", pi)
```

Each cell runs in order. Output from `fmt.Print` and friends appears in the
rendered report. Cells can reference anything declared in `go:defs` blocks
and in prior cells.

### output — Display Blocks

An `output` block shows the expected result of the preceding cell:

```text
Pi is approximately 3.1416
```

Output blocks show the expected result of the preceding cell. In notebook
mode these are updated automatically; in batch mode they serve as
documentation.

## Cell Attributes

Cells can have attributes that control execution:

| Attribute | Description |
|-----------|-------------|
| `name` | Names the cell for reference in error messages |
| `timeout` | Per-cell timeout (e.g., `timeout=30s`) |
| `skip` | Skip this cell during execution |

Example with attributes: a cell named `fetch` with a 10-second timeout:

```go
// This cell has a 10-second timeout.
resp, err := http.Get("https://example.com")
```

## Data Modes

Cells operate in one of two data modes:

- **Copy mode** (default) — Each cell receives a deep copy of the shared
  state. Mutations in one cell do not affect others.
- **Shared mode** — Cells share a single state. Mutations persist across
  cells. Use this when cells need to build on each other's work.

## Supported Output Formats

The report binary (from `gobook build`) supports three output formats:

| Format | Flag | Description |
|--------|------|-------------|
| HTML | `--format html` | Self-contained HTML with inline styles and images |
| PDF | `--format pdf` | PDF rendered via Typst (zero cgo, WASM-based) |
| Markdown | `--format markdown` | Rendered Markdown with output blocks filled in |

## Error Handling

If a cell panics, the error is caught and displayed inline in the output.
The report continues executing subsequent cells. Errors include the cell
name and line number for easy debugging.

## Imports and Dependencies

Go notebooks use standard Go imports. The generated code is a regular Go
module, so any package available via `go get` can be imported in `go:defs`.

The build process runs `go mod tidy` automatically to resolve dependencies.

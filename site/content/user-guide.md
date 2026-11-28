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

Cells can have attributes on the info string that control execution and
data flow. The full info string syntax is:

```
go:notebook [name=<id>] [exported(...)] [reads(...)] [writes(...)] [secrets(...)] [data=embedded|live] [timeout=<dur>]
```

| Attribute | Syntax | Description |
| --------- | ------ | ----------- |
| `name` | `name=<identifier>` | Names the cell for reference in error messages and cross-cell access |
| `exported` | `exported(Field Type, ...)` | Declares fields exported to downstream cells |
| `reads` | `reads(file:embed, file:runtime)` | Declares files this cell reads, with access mode |
| `writes` | `writes(file, ...)` | Declares files this cell writes |
| `secrets` | `secrets(name, ...)` | Declares secrets this cell needs (provided at runtime) |
| `data` | `data=embedded` or `data=live` | Controls when data is captured (see Data Modes) |
| `timeout` | `timeout=<duration>` | Per-cell execution timeout (e.g., `timeout=30s`) |

Example with attributes: a cell named `fetch` with exported fields, a file
read, a secret, and a 10-second timeout:

```
```go:notebook name=fetch exported(Result string) reads(config.json:embed) secrets(api_key) timeout=10s
```

### File Read Modes

The `reads()` attribute supports two modes for each file:

| Mode | Syntax | Description |
| ---- | ------ | ----------- |
| `embed` | `file:embed` | File contents embedded at build time |
| `runtime` | `file:runtime` | File read at render time |

## Data Modes

Cells can specify a data mode that controls when their data is captured:

| Mode | Attribute | Description |
| ---- | --------- | ----------- |
| None | *(default)* | No special data handling |
| Embedded | `data=embedded` | Data is captured at build time and baked into the binary |
| Live | `data=live` | Data is fetched fresh each time the report is rendered |

Use `data=embedded` when data should be frozen at build time (e.g., a
snapshot of query results). Use `data=live` when the report should fetch
current data on every render.

## Supported Output Formats

The report binary (from `gobook build`) supports three output formats:

| Format | Flag | Description |
| ------ | ---- | ----------- |
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

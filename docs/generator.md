# Generator

The generator (`nb/core/generator.go`, `nb/core/writepackage.go`) transforms a
parsed notebook into compilable Go source code.

## Cell Structs

Named cells produce a struct and a run function:

```go
type loadCell struct {
    err   error
    rows  []Row
    total float64
}
```

The `err` field is always first. Exported fields are appended in declaration
order from the `exported(...)` attribute.

Unnamed cells produce only a function (no struct, no return value).

## Run Functions

Named cell signature:

```go
func runLoad(ctx *nb.Context, /* upstream named cells by value */) (out loadCell)
```

Upstream named cells are passed by value as additional parameters. The function
uses named returns so the deferred closure can populate them.

Unnamed cell signature:

```go
func runCell3(ctx *nb.Context, load loadCell) {
```

## Deferred Panic Recovery

Every generated function (named and unnamed) installs a deferred closure:

```go
defer func() {
    if r := recover(); r != nil {
        err = nb.PanicError(r)
    }
    out = loadCell{err: err, rows: rows, total: total}
}()
```

Named cells assign both `err` and exported fields. Unnamed cells discard the
recovered value with `_ = nb.PanicError(r)`.

## Line Directives

Each cell body is preceded by a `//line` directive:

```text
//line notebook.md:42:1
```

This maps compiler errors and stack frames back to Markdown line numbers.

## Source Map

`WritePackage` scans the assembled source for `//line` directives and
populates a `SourceMap`. Each entry records `(mdLine, mdCol) → (goLine,
goCol, length)`. The LSP proxy uses this for bidirectional position
translation.

## Import Pruning

After assembly, `PruneImports` parses the Go source with `go/ast`, collects
all used identifiers, and removes import lines for packages that are not
referenced. This handles cells that use only a subset of the imports declared
in `go:defs` blocks.

## Report Mode

When `GenerateResult.ReportMode` is true, `generateReportMain` produces a
`main()` with a cobra root command. Flags include `--format` (html/pdf/md),
`-o`, `--passphrase`, `--passphrase-file`, `--key-file`, and
`--cell-timeout`. The generated `runCells` function executes all cells and
builds a `document.Document` from display output. If embedded cells are
present, `--passphrase`, `--passphrase-file`, or `--key-file` is required.

## Socket Mode

When `GenerateResult.SocketMode` is true, `generateSocketMain` produces a
`main()` for notebook-mode single-cell execution. The process reads
`GOBOOK_SOCKET` from the environment, connects to the kernel's Unix socket,
fetches upstream cell data via `client.FetchUpstream`, runs the target cell,
and sends display output and the serialized result back to the kernel via
`client.SendDisplay` and `client.SendResult`.

# Parser

The parser (`nb/core/parser.go`, `nb/core/infostring.go`) converts a notebook
Markdown file into a sequence of typed blocks.

## Block Types

- **prose** — any content that is not a recognized fenced block
- **go:defs** — shared declarations (imports, consts, types, funcs)
- **go:notebook** — an executable cell
- **output** — rendered output (generated, never executed)

## Fenced Block Rules

Fences follow CommonMark: at least three backticks (`` ` ``) or tildes (`~`),
with optional leading indent of up to three spaces. A closing fence must use
the same character, be at least as long, and have no info string. Backtick
fences cannot contain backticks in the info string.

Nested fences of the same character are supported: an inner fence must be
shorter than the outer fence to remain nested.

The block kind is determined by the first word of the info string:

- `go:defs` → `BlockDefs`
- `go:notebook` → `BlockCell`
- `output` → `BlockOutput`
- anything else (including plain `go`) → `BlockProse`

## Info String Grammar

```text
info   = "go:notebook" { " " attr }
attr   = "name=" ident
       | "exported(" fields ")"
       | "reads(" filelist ")"
       | "writes(" namelist ")"
       | "secrets(" namelist ")"
       | "data=" ("embedded" | "live")
       | "timeout=" duration
```

- `name=<ident>` — names the cell; enables export and upstream references
- `exported(<fields>)` — Go field declarations added to the cell struct
- `reads(<name>:embed, <name>:runtime, ...)` — files the cell may open
- `writes(<name>, ...)` — files the cell may create
- `secrets(<name>, ...)` — secrets the cell may read via `ctx.Secret`
- `data=embedded` — cell data is captured at build time and embedded encrypted
- `data=live` — cell data is fetched at render time
- `timeout=<duration>` — per-cell execution limit (e.g., `30s`, `2m`)

## Position Tracking

Every block records `StartPos`, `EndPos`, and `BodyPos` as 1-based line and
column numbers. The body position points to the first line after the opening
fence. These positions are used by the source map and error messages.

## Example

```markdown
```go:notebook name=load exported(rows []Row, total float64)
rows, total = fetchData()
` `` `
```

This parses as `BlockCell` with `attrs.Name="load"` and two exported fields.

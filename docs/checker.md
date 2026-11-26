# Checker

The checker (`nb/core/checker.go`) runs static analysis on a parsed notebook
before code generation. Every violation blocks compilation.

## Cell Rules

- **Rule 1** — `exported(...)` requires `name=`. An unnamed cell cannot be
  referenced so exports would be unreachable.
- **Rule 4** — Cell bodies may only use bare `return`, not `return <value>`.
  The deferred function sets the named return value.
- **Rule 5** — Cells may only reference named cells that appear above them.
  Enforced structurally: later cells are not passed as parameters.
- **Rule 6** — Assignment to another cell's field (e.g., `load.val = x`) and
  taking its address (`&load.val`) are refused. Cell values are read-only.
- **Rule 7** — Exported fields must be serializable. Types `chan`, `func`, and
  `unsafe.Pointer` are rejected (best effort on declared type string).

## Defs Rules

- **Rule 2** — `go:defs` may only contain `import`, `const`, `type`, and
  `func` declarations.
- **Rule 3** — `func init()` is not allowed in `go:defs`.
- **Rule 14** — `var` declarations in `go:defs` require a `//go:embed`
  directive immediately above them.
- **Rule 15** — `//go:embed` and `//go:generate` directives are only allowed
  in `go:defs`, not in cell bodies.

## Security Rules

- **Rule 8** — `ctx.Open(name)` requires `name` to be declared in `reads()`.
- **Rule 9** — `ctx.Create(name)` requires `name` to be declared in
  `writes()`.
- **Rule 10** — Direct calls to `os.Open`, `os.ReadFile`, `os.Create`,
  `os.WriteFile`, and `os.OpenFile` are refused in cells. Use `ctx.Open` and
  `ctx.Create` instead.
- **Rule 11** — `ctx.Secret(name)` requires `name` to be declared in
  `secrets()`.
- **Rule 12** — Secrets must not flow directly into exported fields. Direct
  assignment `exported_field = ctx.Secret(...)` is refused (best effort).
- **Rule 13** — Passphrases must be at least 12 Unicode code points.
  Enforced at build time, not by `Check()`.

## File Rules

- **Rule 16** — Image paths in prose must be relative and must not leave the
  notebook directory. Absolute paths and `../` traversal are refused.
- **Rule 17** — Referenced images must exist and have an accepted extension:
  `png`, `jpg`, `gif`, `webp`, `bmp`, `tiff`, `svg`. Only checked when
  `CheckOptions.NotebookDir` is set.

## HTML/Prose Rules

- **Rule 18** — Raw HTML in prose may not use refused tags (`script`,
  `iframe`, `frame`, `object`, `embed`, `base`, `meta`), event-handler
  attributes (`on*`), or `javascript:` URLs in `href`/`src`/`action`.
- **Rule 19** — `<!DOCTYPE>` and `<!ENTITY>` declarations are refused to
  block external entity expansion.
- **Rule 20** — Remote URLs in `<link href>` are refused. Stylesheets must be
  embedded locally.
- **Rule 21** — SVG files referenced in prose must be valid XML.

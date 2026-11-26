# go-notebook Documentation

## Specification

The full design is in the [project wiki](https://github.com/asymmetric-effort/go-notebook/wiki/Go-Notebook-Specification).

## Milestones

The delivery plan is in the [milestones page](https://github.com/asymmetric-effort/go-notebook/wiki/Milestones).

## Feature Documentation

- [parser.md](parser.md) — Block extraction, fenced code rules, and info-string grammar
- [checker.md](checker.md) — All 21 rules with descriptions, grouped by category
- [generator.md](generator.md) — Cell structs, run functions, source map, and report/socket modes
- [encryption.md](encryption.md) — Passphrase mode, recipient mode, payload format, and key generation
- [charts.md](charts.md) — 7 chart types, ctx.Chart usage, SVG/PNG rendering, and interactive script
- [server-api.md](server-api.md) — All server endpoints and their request/response shapes
- [lsp-proxy.md](lsp-proxy.md) — stdio proxy to gopls, position translation, and LSP framing

## Notes and Audits

- [compile-cost.md](compile-cost.md) — Compile-time cost analysis
- [editor-responsiveness.md](editor-responsiveness.md) — Editor responsiveness notes
- [govulncheck-exceptions.md](govulncheck-exceptions.md) — Vulnerability check exceptions
- [typst-feature-gaps.md](typst-feature-gaps.md) — Typst feature gaps
- [typst-fonts.md](typst-fonts.md) — Typst font notes
- [typst-license-audit.md](typst-license-audit.md) — Typst license audit
- [typst-measurements.md](typst-measurements.md) — Typst measurements
- [yamlfmt-lint-mode.md](yamlfmt-lint-mode.md) — yamlfmt lint mode notes

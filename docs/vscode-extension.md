# VSCode extension

## Overview

The VSCode extension provides syntax highlighting for `go:notebook` and
`go:defs` blocks inside Markdown, and connects to `gobook lsp` for
completion, hover and diagnostics.

## Implementation plan

1. TypeScript extension that registers a Markdown language contribution
2. TextMate grammar for `go:notebook` and `go:defs` block highlighting
3. LSP client connecting to `gobook lsp` on stdio
4. Extension published to the VS Code marketplace

## Status

Planned for M9. Depends on M8 (Language server proxy).

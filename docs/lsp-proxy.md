# LSP Proxy

The LSP proxy (`nb/lsp/proxy.go`) bridges the browser editor and `gopls`
so that the editor sees Markdown line numbers while `gopls` works on the
generated Go source.

## Architecture

```text
Editor (browser) <--> Proxy <--> gopls
```

The proxy spawns `gopls serve` as a subprocess, pipes stdin and stdout, and
rewrites position fields in both directions using the source map.

## LSP Message Framing

LSP uses HTTP-style framing over stdio:

```text
Content-Length: <N>\r\n
\r\n
<N bytes of JSON>
```

`readLSPMessage` reads headers line by line until a blank line, parses
`Content-Length`, then reads exactly that many bytes. `writeLSPMessage` writes
the header and body in one pass.

## Position Translation

The source map (`nb/core/sourcemap.go`) records `(mdLine, mdCol) →
(goLine, goCol, length)` entries built from `//line` directives in the
generated source.

For messages flowing editor → gopls (`toGo=true`): Markdown positions are
translated to Go positions via `sm.MdToGo`. For messages flowing gopls →
editor (`toGo=false`): Go positions are translated back via `sm.GoToMd`.

Translation is recursive: `translateObj` walks the JSON tree looking for
objects that match LSP `Position` shape (`{"line":N,"character":N}`) or
`Range` shape (`{"start":...,"end":...}`). LSP uses 0-based lines; the source
map uses 1-based, so offsets are applied on entry and exit.

Positions that fall outside any mapped range are passed through unchanged.

## Incremental Regeneration

When the user edits a cell, the server calls `core.Generate` again and passes
the new `SourceMap` to the proxy via `SetSourceMap`. This keeps the
translation table current without restarting `gopls`.

## Diagnostics

The server's `/api/lsp/diagnostics` endpoint does not go through the stdio
proxy. It runs the checker directly and maps `CheckError.Line` values back to
Markdown positions without round-tripping through gopls. Full `gopls`
diagnostics (unused imports, type errors) require the proxy to be running and
the generated source file to be on disk.

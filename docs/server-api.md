# Server API

The notebook server (`nb/server/server.go`) exposes a JSON HTTP API and
serves the browser editor. All responses include a
`Content-Security-Policy` header that blocks remote loads, inline scripts,
and framing.

## Endpoints

### GET /api/status

Returns notebook name, cell count, and an `ok` flag. Also refreshes the
in-memory parse result.

```json
{"notebook":"analysis.md","cells":4,"ok":true}
```

### GET /api/cells

Returns an array of cell descriptors: `name`, `line`, `exported`, `body`,
`reads`, `writes`, `secrets`, `data`, `timeout`.

### PUT /api/cells/update

Rewrites one cell in the notebook file identified by `line` (opening fence
line, 1-based). Body: same fields as `/api/cells` items plus the new
`body`. Marks the updated cell and all downstream cells stale.

### GET /api/blocks

Returns all blocks (prose, cell, defs, output) with `kind`, `line`,
`end_line`, `body`, `info`, `name`, and `exported`.

### PUT /api/prose/update

Rewrites a prose block identified by `line`. Body: `{"line":N,"body":"..."}`.

### POST /api/run

Runs one cell or all cells. Optional JSON body `{"cell":"name"}` selects a
single cell. Without a cell name, runs all cells as a batch. Returns
`stdout`, `stderr`, `exit_code`, `error`, and structured `errors` array with
`message`, `cell`, `file`, `line`, and `kind` fields (`compile`, `runtime`,
`panic`, `timeout`).

Single-cell mode refuses to run if any upstream cell has previously failed,
returning HTTP 409 with `refused` naming the failed upstream.

### POST /api/run/all

Runs all named cells sequentially, stopping on the first error. Returns an
array of run responses.

### GET /api/run/state

Returns the current execution state of all cells: `name`, `stdout`,
`stderr`, `exit_code`, `error`, `stale`, `failed`.

### PUT /api/secrets

Stores a named secret value in memory (never persisted).
Body: `{"name":"api_key","value":"..."}`.

### GET /api/secrets

Returns the names of secrets that have been set (never the values).

### DELETE /api/secrets

Clears all in-memory secrets.

### GET /api/lsp/diagnostics

Runs the checker on the current notebook and returns structured diagnostics
mapped to Markdown line numbers.

```json
{"diagnostics":[{"file":"nb.md","line":12,"col":1,"message":"...","severity":"error"}]}
```

### GET /api/lsp/completion

Returns completion items: named cells, their exported fields
(`cell.field`), and common context methods (`ctx.Display`, `ctx.Open`,
`ctx.Create`, `ctx.Secret`, `ctx.Chart`).

### GET /api/lsp/hover?word=\<token\>

Returns hover markdown for a cell name or `cell.field` reference.

### GET /api/editor

Serves the embedded editor HTML from `static/editor.html`.

### GET /api/chart.js

Serves the interactive chart script. Used in the CSP hash for rendered pages.

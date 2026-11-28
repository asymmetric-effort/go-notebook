# Browser Editor

go-notebook includes a browser-based notebook editor powered by `gobook serve`.
The editor provides a rich editing experience with Go language intelligence.

## Launching the Editor

```bash
gobook serve notebook.md
```

This starts a local HTTP server on `localhost:8080` (configurable with
`--host` and `--port`). The root path (`/`) redirects to the `/editor`
page.

## Features

### Go Autocompletion

The editor proxies Language Server Protocol (LSP) requests to `gopls`,
providing full Go autocompletion as you type. This includes:

- Package imports
- Function and method names
- Struct field completion
- Variable and constant names

### Diagnostics

Real-time error checking shows compilation errors, unused imports, and other
diagnostics inline as you edit cells. The `/api/lsp/diagnostics` endpoint
runs the notebook checker and reports errors with Markdown line numbers.

### Hover Documentation

Hover over any identifier to see its type signature and documentation.
The editor resolves cell names and exported field references.

### Cell Execution

Execute cells individually or run the entire notebook. Output appears inline
below each cell. The kernel compiles and runs the generated Go package for
each execution request.

### Block Editing

The editor supports adding, deleting, and reordering blocks (prose, cell,
and defs) through the API. Prose blocks can be edited in-place.

## Architecture

The editor uses a two-tier HTTP architecture:

1. **Browser UI** — The frontend editor interface, served as static HTML/JS
   from embedded files at `/static/`.
2. **HTTP Server** — `gobook serve` handles all communication over HTTP.
   It parses the notebook, manages cell state, and executes code by
   generating a temporary Go package and running it via `go run`. An
   embedded `kernel.Kernel` manages cell execution with timeouts. An
   optional LSP proxy to `gopls` provides completions, hover, and
   diagnostics.

There are no WebSockets or Unix sockets. All client-server communication
uses standard HTTP requests to the JSON API endpoints under `/api/`.

### Key API Endpoints

| Endpoint | Method | Description |
| -------- | ------ | ----------- |
| `/api/status` | GET | Notebook status (name, cell count, version) |
| `/api/cells` | GET | List all cells with attributes and body |
| `/api/cells/update` | PUT | Update a cell's body and attributes |
| `/api/blocks` | GET | List all blocks (prose, cell, defs, output) |
| `/api/blocks/add` | POST | Insert a new block |
| `/api/blocks/delete` | DELETE | Remove a block |
| `/api/blocks/reorder` | PUT | Move a block up or down |
| `/api/prose/update` | PUT | Update a prose block's content |
| `/api/run` | POST | Run a single cell or all cells |
| `/api/run/all` | POST | Run all cells sequentially |
| `/api/run/state` | GET | Current execution state of all cells |
| `/api/secrets` | GET/PUT/DELETE | Manage in-memory secrets |
| `/api/lsp/diagnostics` | GET | Checker diagnostics with line numbers |
| `/api/lsp/completion` | GET | Completion items (cells, fields, keywords) |
| `/api/lsp/hover` | GET | Hover info for cell/field references |
| `/api/notebook/init` | POST | Initialize a new notebook with starter content |

## Keyboard Shortcuts

| Shortcut | Action |
| -------- | ------ |
| `Shift+Enter` | Execute current cell |
| `Ctrl+Enter` | Execute current cell and stay |
| `Ctrl+S` | Save notebook |

## Configuration

```bash
# Custom port
gobook serve notebook.md --port 9090

# Bind to all interfaces (for remote access)
gobook serve notebook.md --host 0.0.0.0
```

## Security

The server sets a Content-Security-Policy header that blocks remote loads,
inline scripts, and frame/object elements. Raw SVG content in cell output
is stripped to prevent XSS.

## Limitations

- The editor requires a running `gopls` instance (installed automatically
  if Go is on the PATH)
- Large notebooks with many cells may have slower initial LSP startup

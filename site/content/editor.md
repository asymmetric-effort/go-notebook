# Browser Editor

go-notebook includes a browser-based notebook editor powered by `gobook serve`.
The editor provides a rich editing experience with Go language intelligence.

## Launching the Editor

```bash
gobook serve notebook.md
```

This starts a local HTTP server and opens the editor in your default browser.
The editor communicates with a Go kernel over a Unix socket.

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
diagnostics inline as you edit cells.

### Hover Documentation

Hover over any identifier to see its type signature and documentation,
pulled directly from Go source and godoc.

### Cell Execution

Execute cells individually or run the entire notebook. Output appears inline
below each cell. The kernel maintains state between cell executions in
notebook mode.

## Architecture

The editor uses a three-tier architecture:

1. **Browser UI** — The frontend editor interface
2. **HTTP Server** — `gobook serve` handles HTTP requests and serves the UI
3. **Kernel** — A Go process connected via Unix socket that executes cells
   and proxies LSP requests to gopls

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Shift+Enter` | Execute current cell |
| `Ctrl+Enter` | Execute current cell and stay |
| `Ctrl+S` | Save notebook |

## Configuration

The editor reads settings from the gobook configuration:

```bash
# Custom port
gobook serve notebook.md --port 9090

# Bind to all interfaces (for remote access)
gobook serve notebook.md --host 0.0.0.0
```

## Limitations

- The editor requires a running `gopls` instance (installed automatically
  if Go is on the PATH)
- Unix socket communication means the kernel runs on the same machine as
  the server
- Large notebooks with many cells may have slower initial LSP startup

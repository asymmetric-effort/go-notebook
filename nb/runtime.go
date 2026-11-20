// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package nb provides the runtime for go-notebook generated code.
// It is imported by every generated notebook and report binary.
package nb

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Context provides the runtime interface for cell bodies. It gives access
// to file I/O, secrets and display output.
// Context provides the runtime interface for cell bodies. It gives access
// to file I/O, secrets and display output.
type Context struct {
	mu       sync.Mutex
	fs       fs.FS
	outDir   string // Output directory for files created with Create.
	displays []displayEntry
	secrets  map[string]string
}

type displayEntry struct {
	mime string
	data string
}

// NewContext creates a new runtime context. outDir is the directory where
// files created with Create are written.
func NewContext(fsys fs.FS) *Context {
	return &Context{
		fs: fsys,
	}
}

// Open opens a file declared in reads(). The file is resolved through the
// context's file system.
func (ctx *Context) Open(name string) (fs.File, error) {
	if ctx.fs == nil {
		return nil, fmt.Errorf("no file system configured")
	}
	return ctx.fs.Open(name)
}

// SetOutputDir sets the directory where Create writes files.
func (ctx *Context) SetOutputDir(dir string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.outDir = dir
}

// Create creates a file declared in writes(). Returns a writer that the
// cell writes output to. In notebook mode, files go to an output directory.
// In the report binary, files go relative to the -o location.
func (ctx *Context) Create(name string) (io.WriteCloser, error) {
	ctx.mu.Lock()
	dir := ctx.outDir
	ctx.mu.Unlock()

	if dir == "" {
		return nil, fmt.Errorf("no output directory configured for Create")
	}

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create directory for %s: %w", name, err)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", name, err)
	}
	return f, nil
}

// Secret returns a secret value declared in secrets(). It reads from
// environment variables, the secrets file, or terminal prompt.
func (ctx *Context) Secret(name string) string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.secrets != nil {
		return ctx.secrets[name]
	}
	return ""
}

// SetSecret stores a secret value. Used by the kernel and test harness.
func (ctx *Context) SetSecret(name, value string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.secrets == nil {
		ctx.secrets = make(map[string]string)
	}
	ctx.secrets[name] = value
}

// Display records output for rendering. The mime type must be one of the
// accepted types: text/markdown, image/png, image/jpeg, image/gif.
func (ctx *Context) Display(mime, data string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.displays = append(ctx.displays, displayEntry{mime: mime, data: data})
}

// Displays returns all display entries recorded by the cell.
func (ctx *Context) Displays() []struct{ Mime, Data string } {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	result := make([]struct{ Mime, Data string }, len(ctx.displays))
	for i, d := range ctx.displays {
		result[i] = struct{ Mime, Data string }{d.mime, d.data}
	}
	return result
}

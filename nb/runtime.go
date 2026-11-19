// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package nb provides the runtime for go-notebook generated code.
// It is imported by every generated notebook and report binary.
package nb

import (
	"fmt"
	"io"
	"io/fs"
	"sync"
)

// Context provides the runtime interface for cell bodies. It gives access
// to file I/O, secrets and display output.
type Context struct {
	mu       sync.Mutex
	fs       fs.FS
	displays []displayEntry
	secrets  map[string]string
}

type displayEntry struct {
	mime string
	data string
}

// NewContext creates a new runtime context.
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

// Create creates a file declared in writes(). Returns a writer that the
// cell writes output to.
func (ctx *Context) Create(name string) (io.WriteCloser, error) {
	// Placeholder: in notebook mode this writes to an output directory,
	// in build mode relative to -o location.
	return nil, fmt.Errorf("Create not yet implemented")
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

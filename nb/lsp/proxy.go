// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package lsp implements gobook lsp: the language server proxy that
// translates positions between Markdown and generated Go and forwards
// requests to gopls.
package lsp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

// Proxy is the LSP proxy that sits between an editor and gopls.
type Proxy struct {
	sourceMap *core.SourceMap
	goplsCmd  *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
}

// New creates a new LSP proxy.
func New() *Proxy {
	return &Proxy{
		sourceMap: core.NewSourceMap(),
	}
}

// Start launches gopls as a subprocess and begins proxying.
func (p *Proxy) Start(ctx context.Context) error {
	p.goplsCmd = exec.CommandContext(ctx, "gopls", "serve")
	var err error
	p.stdin, err = p.goplsCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("gopls stdin: %w", err)
	}
	p.stdout, err = p.goplsCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("gopls stdout: %w", err)
	}
	p.goplsCmd.Stderr = os.Stderr
	if err := p.goplsCmd.Start(); err != nil {
		return fmt.Errorf("start gopls: %w", err)
	}
	return nil
}

// Close shuts down the proxy and gopls.
func (p *Proxy) Close() error {
	if p.stdin != nil {
		p.stdin.Close()
	}
	if p.goplsCmd != nil && p.goplsCmd.Process != nil {
		return p.goplsCmd.Process.Kill()
	}
	return nil
}

// TranslateMdToGo translates a Markdown position to a Go position.
func (p *Proxy) TranslateMdToGo(line, col int) (int, int, bool) {
	return p.sourceMap.MdToGo(line, col)
}

// TranslateGoToMd translates a Go position to a Markdown position.
func (p *Proxy) TranslateGoToMd(line, col int) (int, int, bool) {
	return p.sourceMap.GoToMd(line, col)
}

// UpdateSourceMap regenerates the source map from a fresh parse.
func (p *Proxy) UpdateSourceMap(sm *core.SourceMap) {
	p.sourceMap = sm
}

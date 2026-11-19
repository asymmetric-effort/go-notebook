// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package engine provides the Typst compilation engine running on wazero.
// It embeds the Typst WASM module and provides compile-to-PDF and
// SVG-to-PNG conversion with no cgo.
package engine

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed typst-module.wasm
var wasmModule []byte

// Command is the JSON protocol for communicating with the Typst module.
type Command struct {
	Cmd    string `json:"cmd"`
	Input  string `json:"input"`
	Format string `json:"format,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Response is the JSON response from the Typst module.
type Response struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	InputLen int    `json:"input_len,omitempty"`
	Data     []byte `json:"data,omitempty"`
}

// Engine runs the Typst WASM module on wazero.
type Engine struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// New creates a new Typst engine.
func New(ctx context.Context) (*Engine, error) {
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	compiled, err := r.CompileModule(ctx, wasmModule)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("compile typst module: %w", err)
	}

	return &Engine{runtime: r, module: compiled}, nil
}

// Close releases the engine resources.
func (e *Engine) Close(ctx context.Context) error {
	return e.runtime.Close(ctx)
}

// Run executes a command against the Typst module and returns the response.
func (e *Engine) Run(ctx context.Context, cmd Command) (Response, error) {
	input, err := json.Marshal(cmd)
	if err != nil {
		return Response{}, fmt.Errorf("marshal command: %w", err)
	}

	stdin := bytes.NewReader(input)
	var stdout bytes.Buffer

	cfg := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithName("")

	mod, err := e.runtime.InstantiateModule(ctx, e.module, cfg)
	if err != nil {
		return Response{}, fmt.Errorf("run typst module: %w", err)
	}
	defer mod.Close(ctx)

	var resp Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return Response{}, fmt.Errorf("unmarshal response: %w (output: %q)", err, stdout.String())
	}

	return resp, nil
}

// Compile compiles Typst source to the specified format (pdf or png).
func (e *Engine) Compile(ctx context.Context, source string, format string) ([]byte, error) {
	resp, err := e.Run(ctx, Command{Cmd: "compile", Input: source, Format: format})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("typst compile: %s", resp.Error)
	}
	return resp.Data, nil
}

// SVGToPNG converts an SVG string to a PNG image.
func (e *Engine) SVGToPNG(ctx context.Context, svg string, width, height int) ([]byte, error) {
	resp, err := e.Run(ctx, Command{Cmd: "svg_to_png", Input: svg, Width: width, Height: height})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("typst svg_to_png: %s", resp.Error)
	}
	return resp.Data, nil
}
